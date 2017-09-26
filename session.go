package netconf

import (
	"fmt"
	"io"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// Session wraps an *ssh.Session providing additional NETCONF functionality.
// An initialized Session is a io.ReadWriteCloser, with the io.Reader connected
// to the ssh.Session's stdout, and the io.WriteCloser connected to the
// ssh.Sessions's stdin.
type Session struct {
	reader      io.Reader
	writeCloser io.WriteCloser
	sshSession  *ssh.Session
	sshClient   *ssh.Client
}

// NewSession creates a new session ready for use with the NETCONF SSH subsystem.
// It uses the credentials given by the ssh.ClientConfig argument to connect to
// the target.
// Hello messages are negotiated, and the server's hello message is returned along
// with a newly allocated Session pointer.
func NewSession(clientConfig *ssh.ClientConfig, target string) (*Session, *HelloMessage, error) {

	var session Session
	var err error

	session.sshClient, err = ssh.Dial("tcp", target, clientConfig)
	if err != nil {
		return nil, nil, err
	}

	if session.sshSession, err = session.sshClient.NewSession(); err != nil {
		_ = session.sshClient.Close()
		return nil, nil, err
	}

	closeAll := func() {
		_ = session.sshClient.Close()
		_ = session.sshSession.Close()
	}

	if err := session.sshSession.RequestSubsystem("netconf"); err != nil {
		closeAll()
		return nil, nil, err
	}

	if session.reader, err = session.sshSession.StdoutPipe(); err != nil {
		closeAll()
		return nil, nil, err
	}

	if session.writeCloser, err = session.sshSession.StdinPipe(); err != nil {
		closeAll()
		return nil, nil, err
	}

	var helloMessage HelloMessage
	if err := session.NewDecoder().DecodeHello(&helloMessage); err != nil {
		closeAll()
		return nil, nil, err
	}

	if _, err := io.Copy(&session, strings.NewReader(DefaultHelloMessage)); err != nil {
		closeAll()
		return nil, nil, err
	}

	return &session, &helloMessage, nil
}

// NewReplyReader returns a ReplyReader that reads exactly one
// NETCONF RPC Reply from the session's stdout stream. The ReplyReader
// strictly satisfies io.Reader interface by reading from the stream
// until the NETCONF message separator "]]>]]>" is reached, and an io.EOF
// error is returned. The io.EOF error is also returned on all subsequent
// calls.
//
// The ReplyReader does not close the underlying session. Multiple
// ReplyReaders are required to read multiple replies from the same session.
func (s *Session) NewReplyReader() *ReplyReader {
	return NewReplyReader(s)
}

// Read is a partial implementation of the io.Reader interface.
// It reads directly from the session without any modifications.
// It is not compliant with the standard io.Reader interface
// because an EOF is only returned if the session is closed.
//
// Most will use ReplyReader or Decoder.
func (s *Session) Read(p []byte) (n int, err error) {
	return s.reader.Read(p)
}

// Write is the most basic implementation of the io.Writer
// interface. It writes directly to the stdin stream of the
// NETCONF session, and does not write a NETCONF message
// separator "]]>]]>".
//
// Most will use Encoder.
func (s *Session) Write(p []byte) (n int, err error) {
	return s.writeCloser.Write(p)
}

// Close closes all session resources in the following order:
//
//  1. stdin pipe
//  2. SSH session
//  3. SSH client
//
// Errors are returned with priority matching the same order.
func (s *Session) Close() error {

	var (
		writeCloseErr      error
		sshSessionCloseErr error
		sshClientCloseErr  error
	)

	if s.writeCloser != nil {
		writeCloseErr = s.writeCloser.Close()
	}

	if s.sshSession != nil {
		sshSessionCloseErr = s.sshSession.Close()
	}

	if s.sshClient != nil {
		sshClientCloseErr = s.sshClient.Close()
	}

	if writeCloseErr != nil {
		return writeCloseErr
	}

	if sshSessionCloseErr != nil {
		return sshSessionCloseErr
	}

	return sshClientCloseErr
}

// NewDecoder returns a new Decoder object attached to the stdout pipe
// of the underlying SSH session.
func (s *Session) NewDecoder() *Decoder {
	return NewDecoder(s.reader)
}

// NewTimeoutDecoder returns a new Decoder attached to the stdout pipe
// of the underlying SSH session. The Decoder's io.TrimReader is wrapped to set a read
// timeout on the underlying net.Conn before every read.
func (s *Session) NewTimeoutDecoder(timeout time.Duration) *Decoder {
	return NewDecoder(s.NewDeadlineReader(timeout))
}

// DeadlineError is returned when a read or write deadline is reached.
type DeadlineError struct {
	Op        string
	BeginTime time.Time
	FailTime  time.Time
	Deadline  time.Duration
}

// Error implements the error interface.
func (te *DeadlineError) Error() string {
	return fmt.Sprintf("netconf: %s deadline %s began %s expired %s",
		te.Op, te.Deadline, te.BeginTime, te.FailTime)
}

// NewDeadlineReader decorates the session's io.Reader with
// a new DeadlineReader.
//
// The DeadlineReader only adds a deadline when reading from
// the stream. It does not handle higher level functionality,
// like a complete implementation of an io.Reader, or discarding
// NETCONF message separators.
func (s *Session) NewDeadlineReader(deadline time.Duration) io.Reader {
	return &DeadlineReader{
		reader:   s.reader,
		deadline: deadline,
	}
}

// NewEncoder returns a new Encoder object attached to the stdin pipe
// of the underlying SSH session.
func (s *Session) NewEncoder() *Encoder {
	return NewEncoder(s.writeCloser)
}

// TODO: Make RPCWriter that handles writing NETCONF message separators.
// TODO: Make all other readers and writers start with the ReplyReader, and
// TODO: RPCWriter, which has the sole job of implementing the standard
// TODO: io.Reader and io.Writer interfaces.
