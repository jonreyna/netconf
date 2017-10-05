package netconf

import (
	"encoding/xml"
	"io"

	"golang.org/x/crypto/ssh"
)

const DefaultPort = "830"

// MessageSeparator is a constant defining the standard
// NETCONF message seaprator. It should be written to the
// session after writing any method.
const MessageSeparator = `]]>]]>
`

var messageSeparatorBytes = []byte(MessageSeparator)

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

var defaultHelloMessageBytes = []byte(DefaultHelloMessage)

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
		sshSession.Close()
		return nil, nil, err
	}

	ncSession := Session{sshSession: sshSession}
	err = ncSession.initPipes()
	if err != nil {
		ncSession.Close()
		return nil, nil, err
	}

	ncSession.decoder = xml.NewDecoder(ncSession.reader)
	ncSession.encoder = xml.NewEncoder(ncSession.writeCloser)

	hello, err := ncSession.DecodeHello()
	if err != nil {
		ncSession.Close()
		return nil, nil, err
	}

	_, err = ncSession.writeCloser.Write(defaultHelloMessageBytes)

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

	if s.sshSession != nil {
		if err != nil {
			s.sshSession.Close()
		} else {
			err = s.sshSession.Close()
		}
	}

	return err
}

func (s *Session) DecodeHello() (*HelloMessage, error) {
	defer s.reader.Reset()
	var hello HelloMessage
	return &hello, s.decoder.Decode(&hello)
}

// func (s *Session) LocalAddr() net.Addr { }

// func (s *Session) Exec(ctx context.Context, method ...interface{}) (*ReplyReader, error) { }

func (s *Session) execOne(method interface{}) error {

	m, ok := method.(*Method)
	if !ok {
		m = WrapMethod(method)
	}

	return s.encoder.Encode(m)
}

func (s *Session) writeSep() error {
	_, err := s.Write(messageSeparatorBytes)
	return err
}
