package netconf

import (
	"context"
	"encoding/xml"
	"io"

	"golang.org/x/crypto/ssh"
)

const DefaultPort = "830"

// MessageSeparator is a constant defining the standard
// NETCONF message seaprator. It should be written to the
// session after writing any method.
const MessageSeparator = `]]>]]>`

// DefaultHelloMessage is this library's default hello sent to the
// server, when it is not sent manually by the client application.
const DefaultHelloMessage = `<?xml version="1.0" encoding="UTF-8"?>
<hello xmlns="urn:ietf:params:xml:ns:netconf:base:1.0">
<capabilities>
<capability>urn:ietf:params:netconf:base:1.1</capability>
</capabilities>
</hello>
]]>]]>
`

// HelloMessage represents a capabilities exchange message.
type HelloMessage struct {
	XMLName      xml.Name
	Capabilities []string `xml:"capabilities>capability"`
	SessionID    uint     `xml:"session-id,omitempty"`
}

type Session struct {
	sshSession  *ssh.Session
	encoder     *xml.Encoder
	decoder     *xml.Decoder
	reader      *Reader
	writeCloser io.WriteCloser
}

func NewSession(sshSession *ssh.Session) (*Session, *HelloMessage, error) {

	err := sshSession.RequestSubsystem("netconf")
	if err != nil {
		_ = sshSession.Close()
		return nil, nil, err
	}

	ncSession := Session{sshSession: sshSession}
	err = ncSession.initPipes()
	if err != nil {
		_ = ncSession.Close()
		return nil, nil, err
	}

	ncSession.decoder = xml.NewDecoder(ncSession.reader)
	ncSession.encoder = xml.NewEncoder(ncSession.writeCloser)

	hello, err := ncSession.DecodeHello()
	if err != nil {
		_ = ncSession.Close()
		return nil, nil, err
	}

	_, err = ncSession.writeCloser.Write([]byte(DefaultHelloMessage))

	return &ncSession, hello, err
}

func (s *Session) initPipes() error {

	readPipe, err := s.sshSession.StdoutPipe()
	if err != nil {
		return err
	}

	s.writeCloser, err = s.sshSession.StdinPipe()
	if err != nil {
		return err
	}

	s.reader = NewReader(readPipe)

	return nil
}

func (s *Session) Read(p []byte) (n int, err error) {
	return s.reader.Read(p)
}

func (s *Session) ResetReader() {
	s.reader.Reset()
}

func (s *Session) Write(p []byte) (n int, err error) {
	return s.writeCloser.Write(p)
}

func (s *Session) Close() error {

	var err error

	if s.writeCloser != nil {
		err = s.writeCloser.Close()
	}

	if s.sshSession != nil && err == nil {
		return s.sshSession.Close()
	}

	_ = s.sshSession.Close()

	return err
}

func (s *Session) DecodeHello() (*HelloMessage, error) {
	defer s.reader.Reset()
	var hello HelloMessage
	return &hello, s.decoder.Decode(&hello)
}

func (s *Session) ExecOne(ctx context.Context, method, reply interface{}) <-chan error {
	return s.goEncodDecodeOne(ctx, method, reply)
}

func (s *Session) goEncodDecodeOne(ctx context.Context, method, reply interface{}) <-chan error {

	errChan := make(chan error, 1)

	go func() {

		defer close(errChan)

		select {
		case err := <-s.goEncodeOne(ctx, method):
			if err != nil {
				errChan <- err
				return
			}
		case <-ctx.Done():
			errChan <- ctx.Err()
			return
		}

		select {
		case err := <-s.goDecodeOne(ctx, method):
			if err != nil {
				errChan <- err
			}
		case <-ctx.Done():
			errChan <- ctx.Err()
		}
	}()

	return errChan
}

func (s *Session) goDecodeOne(ctx context.Context, reply interface{}) <-chan error {

	errChan := make(chan error, 1)

	go func() {

		defer close(errChan)

		r, ok := reply.(*Reply)
		if !ok {
			r = &Reply{
				Data: reply,
			}
		}

		select {
		case <-ctx.Done():
			errChan <- ctx.Err()
			return
		default:
		}

		if err := s.decoder.Decode(r); err != nil {
			errChan <- err
			return
		}

		for i, err := range r.Error {
			if err.Severity == ErrorSeverityError {
				errChan <- &r.Error[i]
				return
			}
		}
	}()

	return errChan
}

func (s *Session) goEncodeOne(ctx context.Context, method interface{}) <-chan error {

	errChan := make(chan error, 1)

	go func() {

		defer close(errChan)

		m, ok := method.(*Method)
		if !ok {
			m = WrapMethod(method)
		}

		select {
		case <-ctx.Done():
			errChan <- ctx.Err()
			return
		default:
		}

		if err := s.encoder.Encode(m); err != nil {
			errChan <- err
			return
		}

		select {
		case <-ctx.Done():
			errChan <- ctx.Err()
			return
		default:
		}

		if _, err := s.WriteSep(); err != nil {
			errChan <- err
		}
	}()

	return errChan
}

func (s *Session) WriteSep() (n int, err error) {
	const sepWithNewLine = `]]>]]>
`
	return s.Write([]byte(sepWithNewLine))
}
