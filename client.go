package netconf

import (
	"fmt"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

// Client wraps an ssh.Session. It overrides NewSession so it returns a NETCONF
// specific session. It also allows setting Keepalives and deadlines using the
// Config argument's fields.
type Client struct {

	// sshClient may be public in future implementations.
	sshClient *ssh.Client

	// keepaliveTicker signals when to send keepalive messages.
	keepaliveTicker *time.Ticker

	// stopKeepalive signals the keepalive ticker to stop, and allows its
	// encapsulating goroutine to exit cleanly.
	stopKeepalive chan struct{}
}

// Dial creates a ssh.Client using credentials found in Config's ssh.ClientConfig
// and wraps it in a netconf.Client, and sets up other resources to satisfy
// other options set in the config (like deadlines, keepalives, etc.)
func Dial(c *Config) (*Client, error) {

	// create a standard net.Conn for more granular control
	conn, err := net.DialTimeout(c.dialTimeoutArgs())
	if err != nil {
		return nil, err
	}

	// wrap the net.Conn in a DeadlineConn if required by Config
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

	// setup keepalive goroutine if needed
	if c.Keepalive != 0 {
		ncClient.Keepalive(c.Keepalive)
	}

	return &ncClient, nil
}

// Close releases any reasources associated with this Client, including
// its underlying ssh.Client, and signals any supporting goroutines to
// exit.
func (c *Client) Close() error {

	// make sure a ticker exists
	if c.keepaliveTicker != nil {
		close(c.stopKeepalive)
		c.keepaliveTicker.Stop()
	}

	return c.sshClient.Close()
}

// NewSession creates a new ssh.Session using the underlying ssh.Client, and
// upgrades it to a netconf.Session. Unlike the NewSession function, closing
// the Session returned from this method will not close the Client that
// produced it.
func (c *Client) NewSession() (*Session, *HelloMessage, error) {

	sshSession, err := c.sshClient.NewSession()
	if err != nil {
		return nil, nil, err
	}

	return Upgrade(sshSession)
}

// Keepalive sends a global request to the SSH server in a separate goroutine,
// at the given interval, to keep the connection alive. Calling Close on the
// Client stops the timer, and causes the goroutine handling the keepalives to
// exit cleanly.
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
				// receive indicates stopKeepalive is closed, so exit by returning
				c.keepaliveTicker.Stop()
				return
			}
		}
	}()
}

// DeadlineConn wraps a net.Conn to override its Read and Write methods, setting
// a deadline based on its ReadTimeout and WriteTimeout fields.
type DeadlineConn struct {

	// Conn is the embedded, underlying net.Conn, which is likely a standard Conn.
	net.Conn

	// ReadTimeout is added to the time Read is called to produce the Read deadline.
	ReadTimeout time.Duration

	// WriteTimeout is added to the time Write is called to produce the Write deadline.
	WriteTimeout time.Duration
}

// Read sets a read deadline before every call to Read on the underlying
// net.Conn.
func (c *DeadlineConn) Read(p []byte) (n int, err error) {
	err = c.Conn.SetReadDeadline(time.Now().Add(c.ReadTimeout))
	if err != nil {
		return 0, err
	}

	return c.Conn.Read(p)
}

// Write sets a write deadline before every call to Write on the underlying
// net.Conn.
func (c *DeadlineConn) Write(p []byte) (n int, err error) {
	err = c.Conn.SetWriteDeadline(time.Now().Add(c.WriteTimeout))
	if err != nil {
		return 0, err
	}

	return c.Conn.Write(p)
}

// DeadlineError is returned when a read or write deadline is reached.
// TODO: This is error is not currenlty used, and may be removed.
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
