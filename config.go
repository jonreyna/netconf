package netconf

import (
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

// Config contains all available configuration options for a Client.
type Config struct {

	// SSH exposes all available options to configure the SSH client
	// connection.
	SSH *ssh.ClientConfig

	// Keepalive is the interval that keepalive messages are sent to
	// the SSH server. If 0, no keepalive messages are sent.
	Keepalive time.Duration

	// DialTimeout is the duration to wait for dialing to complete before
	// failing with an error. There is no default timeout.
	DialTimeout time.Duration

	// ReadTimeout is the duration added to the current time to set a read
	// deadline. There is no default read deadline.
	ReadTimeout time.Duration

	// WriteTimeout is the duration added to the current time to set a write
	// deadline. There is no default write deadline.
	WriteTimeout time.Duration

	// Network is the type of network to Dial with (e.g. tcp, udp). The
	// default is tcp.
	Network string

	// Address is the dial target, including port. If no port is specified,
	// the default NETCONF port, port 830, is used.
	Address string
}

// dialTimeoutArgs generates the arguments passed to ssh.DialTimeout.
func (c *Config) dialTimeoutArgs() (string, string, time.Duration) {
	if c.Network == "" {
		c.Network = "tcp"
	}
	return c.Network, c.normalizeAddress(), c.DialTimeout
}

// normalizeAddress checks if the target includes a port.
// If it doesn't, the default NETCONF port is joined with it.
// If a port is included, the target is not changed.
func (c *Config) normalizeAddress() string {
	_, _, err := net.SplitHostPort(c.Address)
	if err != nil {
		return net.JoinHostPort(c.Address, DefaultPort)
	}

	return c.Address
}

func (c *Config) hasReadWriteTimeout() bool {
	if c.ReadTimeout != 0 || c.WriteTimeout != 0 {
		return true
	}
	return false
}
