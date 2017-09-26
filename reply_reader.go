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
	session  io.Reader     // attached to stdout of netconf session
	bytesBuf *bytes.Buffer // scratchpad for reader to make implementing the standard io.Reader easier
	buf      []byte        // used to read from session before copying to bytes buffer
	err      error         // once an error is generated, always return it on subsequent calls
}

// NewReplyReader assumes the given reader reads from
// a NETCONF session's stdout, and adapts its behavior to
// a standard io.Reader, allowing it to work with standard
// library methods and functions.
func NewReplyReader(session io.Reader) *ReplyReader {
	return &ReplyReader{
		session:  session,
		bytesBuf: &bytes.Buffer{},
	}
}

// Read performs line oriented reads (using bufio.Scanner),
// and discards newlines characters. This may be undesirable
// if the NETCONF server writes CLI-like output for humans
// (e.g. in an <output> tag).
//
// On the other hand, trimming newlines may be desirable when
// the NETCONF server writes newlines around values (like integers),
// because the standard xml.Decoder uses strconv to parse
// integers, and it returns an error when parsing integers that
// have surrounding white space.
//
// Trimming newlines may be optional in future implementations.
func (rr *ReplyReader) Read(p []byte) (n int, err error) {

	// continue to return error to comply with io.Reader interface
	if rr.err != nil {
		// if there's more to read, return it
		if rr.bytesBuf.Len() != 0 {
			return rr.bytesBuf.Read(p)
		}

		return 0, rr.err
	}

	n, rr.err = rr.session.Read(rr.buf)
	rr.bytesBuf.Write(rr.buf[:n]) // always returns a nil error

	bTrim := bytes.TrimRightFunc(rr.bytesBuf.Bytes(), unicode.IsSpace)
	if bytes.HasSuffix(bTrim, messageSeparatorBytes) {
		// found the message separator
		rr.bytesBuf.Truncate(bytes.LastIndex(bTrim, messageSeparatorBytes))
		rr.err = io.EOF
		return rr.bytesBuf.Read(p)
	}

	// verify buffer is big enough to detect message sep on next read
	if rr.bytesBuf.Len() < len(messageSeparatorBytes) {
		// buffer too small guarantee message separator is detected
		// e.g. contents could be "]]>]]"
		return 0, nil
	}

	n, _ = rr.bytesBuf.Read(p) // may return io.EOF prematurely
	return n, rr.err
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
