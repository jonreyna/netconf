package netconf

import (
	"io"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

// Session wraps an *ssh.Session providing additional NETCONF functionality.
// An initialized Session is a io.ReadWriteCloser, with the io.Reader connected
// to the ssh.Session's stdout, and the io.WriteCloser connected to the
// ssh.Sessions's stdin.
type Session struct {
	netConn net.Conn
	io.Reader
	io.WriteCloser
	s *ssh.Session
}

// Close checks to see if it has a valid io.WriteCloser, and closes it first.
// It then closes the underlying ssh.Session. The ssh.Session's close error
// takes precedence over the error returned by the io.WriteCloser's Close.
func (s *Session) Close() (err error) {

	if s.WriteCloser != nil {
		err = s.WriteCloser.Close()
	}

	if sErr := s.s.Close(); sErr != nil {
		return sErr
	}

	return err
}

// NewDecoder returns a new Decoder object attached to the stdout pipe
// of the underlying SSH session.
func (s *Session) NewDecoder() *Decoder {
	return NewDecoder(s.Reader)
}

// NewTimeoutDecoder returns a new Decoder attached to the stdout pipe
// of the underlying SSH session. The Decoder's io.Reader is wrapped to set a read
// timeout on the underlying net.Conn before every read.
func (s *Session) NewTimeoutDecoder(timeout time.Duration) *Decoder {
	return NewDecoder(s.NewTimeoutReader(timeout))
}

// ReadDeadliner abstracts the action SetReadDeadline, which is
// typically implemented by a net.Conn.
type ReadDeadliner interface {
	SetReadDeadline(t time.Time) error
}

// TimeoutReader is a wrapper for an io.Reader that sets a timeout
// on reads.
type TimeoutReader struct {
	io.Reader
	ReadDeadliner
	time.Duration
}

// Read sets a read deadline before every call to the underlying stream's
// io.Reader.
func (tr *TimeoutReader) Read(b []byte) (n int, err error) {
	err = tr.ReadDeadliner.SetReadDeadline(time.Now().Add(tr.Duration))
	if err != nil {
		return
	}
	return tr.Reader.Read(b)
}

// NewTimeoutReader wraps the underlying io.Reader, which is attached to the stdout
// pipe of the underlying SSH session, into a new io.Reader that sets a timeout on
// the underlying net.Conn before every read.
//
// The reader does not discard NETCONF message separators.
func (s *Session) NewTimeoutReader(timeout time.Duration) io.Reader {
	return &TimeoutReader{
		Reader:   s.Reader,
		Duration: timeout,
	}
}

// NewEncoder returns a new Encoder object attached to the stdin pipe
// of the underlying SSH session.
func (s *Session) NewEncoder() *Encoder {
	return NewEncoder(s.WriteCloser)
}

// NewTimeoutEncoder returns a new Encoder attached to the stdin pipe
// of the underlying SSH session. The Encoder's io.Writer is wrapped
// to set a write timeout on the underlying net.Conn before every write.
func (s *Session) NewTimeoutEncoder(timeout time.Duration) *Encoder {
	return NewEncoder(s.NewTimeoutWriter(timeout))
}

// WriteDeadliner abstracts the action SetWriteDeadline, which is
// typically implemented by a net.Conn.
type WriteDeadliner interface {
	SetWriteDeadline(t time.Time) error
}

// TimeoutWriter is a wrapper for an io.Writer that sets a
// timeout on writes.
type TimeoutWriter struct {
	io.Writer
	WriteDeadliner
	time.Duration
}

// Write sets a write deadline before every call to the underlying stream's
// io.Writer.
func (tw *TimeoutWriter) Write(p []byte) (n int, err error) {
	err = tw.WriteDeadliner.SetWriteDeadline(time.Now().Add(tw.Duration))
	if err != nil {
		return
	}
	return tw.Writer.Write(p)
}

// NewTimeoutWriter wraps the underlying io.Writer, which is attached to
// the stdin pipe of the underlying SSH Session, into a new io.Writer that
// sets a timeout before every write on the underlying net.Conn.
//
// The writer does not automatically write NETCONF message separators.
func (s *Session) NewTimeoutWriter(timeout time.Duration) io.Writer {
	return &TimeoutWriter{
		Writer:   s.WriteCloser,
		Duration: timeout,
	}
}
