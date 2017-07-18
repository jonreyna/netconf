package netconf

import (
	"net"

	"golang.org/x/crypto/ssh"
)

// Client embeds an *ssh.Client to add additional
// NETCONF specific capabilities, while keeping the
// public functionality of the underlying *ssh.Client
// exposed.
type Client struct {
	*ssh.Client
	NetConn         net.Conn
	SSHClientConfig *ssh.ClientConfig
}

// Close closes the underlying ssh.Client and net.Conn.
func (c *Client) Close() (err error) {

	// Client already redirects close to the underlying net.Conn
	// If it is successful, calling f.NetConn.Close() is
	// redundant, and would likely produce an extraneous error
	if c.Client != nil {
		if err = c.Client.Close(); err != nil {
			return err
		}
	}

	if c.NetConn != nil {
		err = c.NetConn.Close()
	}

	return err
}

// NewSession creates an SSH session from the underlying SSH Client,
// requests the netconf subsystem, stdin, stdout, and wraps the
// session object in a NETCONF Session before returning it.
//
// Note that this overrides the ssh.Client.NewSession method.
func (c *Client) NewSession() (*Session, error) {
	if session, err := c.Client.NewSession(); err != nil {
		return nil, err
	} else if err := session.RequestSubsystem("netconf"); err != nil {
		_ = session.Close()
		return nil, err
	} else if reader, err := session.StdoutPipe(); err != nil {
		_ = session.Close()
		return nil, err
	} else if writeCloser, err := session.StdinPipe(); err != nil {
		_ = session.Close()
		return nil, err
	} else {
		return &Session{
			netConn:     c.NetConn,
			s:           session,
			Reader:      reader,
			WriteCloser: writeCloser,
		}, nil
	}
}

// NewClient returns a pointer to a new Client that is used as a convenience to build
// NETCONF objects.
func NewClient(clientConfig *ssh.ClientConfig, target string) (*Client, error) {
	if netConn, err := net.Dial("tcp", target); err != nil {
		return nil, err
	} else if sshConn, newChan, requestChan, err := ssh.NewClientConn(
		netConn,
		netConn.RemoteAddr().String(),
		clientConfig,
	); err != nil {
		_ = netConn.Close()
		return nil, err
	} else {
		return &Client{
			NetConn: netConn,
			Client:  ssh.NewClient(sshConn, newChan, requestChan),
		}, nil
	}
}
