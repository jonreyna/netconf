package netconf

import (
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

type Config struct {
	SSH          *ssh.ClientConfig
	Keepalive    time.Duration
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	Network      string
	Address      string
}

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
