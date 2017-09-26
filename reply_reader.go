package netconf

import (
	"bytes"
	"encoding/xml"
	"io"
	"time"
	"unicode"
)

// TODO: add ReplyReader.Reset method

// ReplyReader reads exactly one RPC reply from the session,
// and discards the message separator. If multiple RPCs need to
// be read from the session, multiple ReplyReaders will be required.
// The io.EOF error is returned on every read after the NETCONF message
// separator is encountered. This is how ReplyReader is able to satisfy
// the strict interpretation of the io.Reader interface.
type ReplyReader struct {
	session io.Reader // attached to stdout of netconf session
	err     error     // once an error is generated, always return it on subsequent calls
}

// NewReplyReader assumes the given reader reads from
// a NETCONF session's stdout, and adapts its behavior to
// a standard io.Reader, allowing it to work with standard
// library methods and functions.
// It is intended to read exactly one RPC reply, however
// it can be reused after calling the Reset method.
func NewReplyReader(session io.Reader) *ReplyReader {
	return &ReplyReader{
		session: session,
	}
}

// Read implements the io.Reader interface by returning io.EOF
// whenever the standard NETCONF message separator is found in
// the byte stream.
func (rr *ReplyReader) Read(p []byte) (n int, err error) {

	if rr.err != nil {
		return 0, rr.err
	}

	n, rr.err = rr.session.Read(p)

	bTrim := bytes.TrimRightFunc(p[:n], unicode.IsSpace)
	if bytes.HasSuffix(bTrim, messageSeparatorBytes) {
		n = bytes.LastIndex(bTrim, messageSeparatorBytes)
		rr.err = io.EOF
	}

	return n, rr.err
}

// Reset clears the internal error field, allowing
// this reader to be reused.
func (rr *ReplyReader) Reset() {
	rr.err = nil
}

// WithDeadline decorates the ReplyReader with a DeadlineReader.
// The DeadlineReader sets its deadline before every call to Read.
func (rr *ReplyReader) WithDeadline(deadline time.Duration) *DeadlineReader {
	return &DeadlineReader{
		reader:   rr,
		deadline: deadline,
	}
}

// DeadlineReader is a decorator for an io.Reader that sets a deadline
// before every read. It can only be constructed by a ReplyReader's
// WithDeadline method.
type DeadlineReader struct {
	reader   io.Reader     // NETCONF session's stdout reader
	deadline time.Duration // deadline to set before every call to Read
}

// Read sets a deadline before every call to Read, and returns a DeadlineError
// if reading is not complete before the configured deadline expires.
// It is recommended that you close the session upon receipt of a DeadlineError,
// otherwise a subsequent read will return whatever the NETCONF server wrote to
// its stdout stream after the deadline expired.
func (dr *DeadlineReader) Read(b []byte) (n int, err error) {

	var begin time.Time
	timer := time.NewTimer(dr.deadline)
	defer timer.Stop()

	ch := make(chan struct{})
	go func() {
		begin = time.Now()
		n, err = dr.reader.Read(b)
		ch <- struct{}{}
	}()

	select {
	case <-ch:
		return n, err
	case timeDone := <-timer.C:
		return n, &DeadlineError{
			Op:        "read",
			BeginTime: begin,
			FailTime:  timeDone,
			Deadline:  dr.deadline,
		}
	}
}

func (dr *DeadlineReader) AsDecoder() *Decoder {
	return &Decoder{Decoder: xml.NewDecoder(dr)}
}
