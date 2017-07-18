package netconf

import (
	"encoding/xml"
)

// HelloMessage represents a capabilities exchange message.
type HelloMessage struct {
	XMLName      xml.Name
	Capabilities []string `xml:"capabilities>capability"`
	SessionID    uint     `xml:"session-id,omitempty"`
}

// Copy makes a deep copy of this HelloMessage.
func (h *HelloMessage) Copy() *HelloMessage {
	var c HelloMessage
	if capLen := len(h.Capabilities); capLen != 0 {
		c.Capabilities = make([]string, 0, capLen)
		copy(c.Capabilities, h.Capabilities)
	}
	return &c
}

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
