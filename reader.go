package netconf

import (
	"bytes"
	"io"
	"unicode"
)

// Reader adapts a Session's stdout pipe into a standard reader that returns
// io.EOF errors at the end of every NETCONF reply. The end of a NETCONF
// message is detected by looking for the standard message separator (after
// trimming space) at the end of every NETCONF read.
//
// A single reply can be read using standard library objects, like bytes.Buffer,
// or io.Copy. Reset must be called after a complete message is read, to clear
// the io.EOF error, and to keep using the same Reader.
//
// Reusing the same reader is recommended to avoid unncessary internal buffer
// allocations.
type Reader struct {
	// session can be any io.Reader, but is treated as a pipe attached to
	// stdout.
	session io.Reader

	// buffer contains is used to store the entire message.
	buffer *bytes.Buffer

	// done indicates a message separator was found.
	done bool

	// err preserves errors between reads.
	err error

	// readBuffer is passed to the session's Read method before being
	// copied into the bytes.Buffer.
	readBuffer []byte
}

// NewReader decorates the given io.Reader's Read method with one that
// transparently handles NETCONF message separators. Its goal is to abstract
// away the NETCONF protocol to make using standard library utilities and
// objects easy.
func NewReader(ncSession io.Reader) *Reader {
	return &Reader{
		session:    ncSession,
		buffer:     new(bytes.Buffer),
		readBuffer: make([]byte, bytes.MinRead),
	}
}

// Read implements the standard io.Reader. Internally, on the first call to
// Read, the entire message is read into the internal buffer, and the message
// separator is discarded. This implementation guarantees the message separator
// is found independent of the length of p.
//
// Once a single message is read, the reader will continue to act like a
// standard io.Reader, by returning io.EOF on subsequent reads. Use the Reset
// method to clear the error before reading the next message.
func (r *Reader) Read(p []byte) (n int, err error) {

	for !r.done && err == nil {

		n, err = r.session.Read(r.readBuffer)

		// error is always nil
		r.buffer.Write(r.readBuffer[:n])

		// only preserve non io.EOF errors for subsequent reads
		if err != nil && err != io.EOF {
			r.err = err
		}

		bTrim := bytes.TrimRightFunc(r.buffer.Bytes(), unicode.IsSpace)
		if bytes.HasSuffix(bTrim, []byte(MessageSeparator)) {
			r.buffer.Truncate(bytes.LastIndex(bTrim, []byte(MessageSeparator)))
			r.done = true
		}
	}

	n, err = r.buffer.Read(p)

	// perform read of available data,
	// but always prefer returning the
	// non io.EOF error (if exists)
	if r.err != nil {
		return n, r.err
	}

	return n, err
}

// Reset clears any errors returned by Read, and prepares it for the next
// message.
func (r *Reader) Reset() {
	r.done = false
	r.buffer.Reset()
	r.err = nil
}
