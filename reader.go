package netconf

import (
	"bytes"
	"io"
	"unicode"
)

type Reader struct {
	session    io.Reader
	buffer     *bytes.Buffer
	done       bool
	err        error
	readBuffer []byte
}

func NewReader(ncSession io.Reader) *Reader {
	return &Reader{
		session:    ncSession,
		buffer:     new(bytes.Buffer),
		readBuffer: make([]byte, bytes.MinRead),
	}
}

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
		if bytes.HasSuffix(bTrim, messageSeparatorBytes) {
			r.buffer.Truncate(bytes.LastIndex(bTrim, messageSeparatorBytes))
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

func (r *Reader) Reset() {
	r.done = false
	r.buffer.Reset()
	r.err = nil
}
