package netconf

import (
	"fmt"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

type Client struct {
	sshClient       *ssh.Client
	keepaliveTicker *time.Ticker
	stopKeepalive   chan struct{}
}

func Dial(c *Config) (*Client, error) {

	conn, err := net.DialTimeout(c.dialTimeoutArgs())
	if err != nil {
		return nil, err
	}

	if c.hasReadWriteTimeout() {
		conn = &DeadlineConn{
			Conn:         conn,
			ReadTimeout:  c.ReadTimeout,
			WriteTimeout: c.WriteTimeout,
		}
	}

	sshConn, chans, reqs, err := ssh.NewClientConn(conn, c.normalizeAddress(), c.SSH)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	ncClient := Client{
		sshClient: ssh.NewClient(sshConn, chans, reqs),
	}

	if c.Keepalive != 0 {
		ncClient.Keepalive(c.Keepalive)
	}

	return &ncClient, nil
}

func (c *Client) Close() error {

	if c.keepaliveTicker != nil {
		close(c.stopKeepalive)
		c.keepaliveTicker.Stop()
	}

	return c.sshClient.Close()
}

func (c *Client) NewSession() (*Session, *HelloMessage, error) {

	sshSession, err := c.sshClient.NewSession()
	if err != nil {
		return nil, nil, err
	}

	return NewSession(sshSession)
}

func (c *Client) Keepalive(interval time.Duration) {

	c.keepaliveTicker = time.NewTicker(interval)
	c.stopKeepalive = make(chan struct{})

	go func() {

		for {
			select {
			case <-c.keepaliveTicker.C:
				_, _, err := c.sshClient.SendRequest("keepalive@github.com/sourcemonk/netconf", true, nil)
				if err != nil {
					return
				}

			case <-c.stopKeepalive:
				c.keepaliveTicker.Stop()
				return
			}
		}
	}()
}

type DeadlineConn struct {
	net.Conn
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func (c *DeadlineConn) Read(p []byte) (n int, err error) {
	err = c.Conn.SetReadDeadline(time.Now().Add(c.ReadTimeout))
	if err != nil {
		return 0, err
	}

	return c.Conn.Read(p)
}

func (c *DeadlineConn) Write(p []byte) (n int, err error) {
	err = c.Conn.SetWriteDeadline(time.Now().Add(c.WriteTimeout))
	if err != nil {
		return 0, err
	}

	return c.Conn.Write(p)
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
