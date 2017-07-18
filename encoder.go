package netconf

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"io"
)

const (
	// MessageSeparator is explicitly invalid XML used to unambiguously separate NETCONF messages.
	MessageSeparator = `]]>]]>`

	// BaseNamespace is the most basic NETCONF namespace, and is the default namespace for all NETCONF methods.
	BaseNamespace = `urn:ietf:params:xml:ns:netconf:base:1.0`
)

// Method models the structure of a NETCONF method.
// It is useful for wrapping structs that don't
// encode the outer rpc tags.
//
// Multiple methods can be encoded into one RPC.
type Method struct {
	XMLName xml.Name
	Attr    []xml.Attr `xml:",attr"`
	Method  []interface{}
}

// XMLNameTag returns an xml.Name for an RPC's outer tag.
// It is appropriately named "rpc", with an attribute
// that has the given namespace.
func XMLNameTag(namespace string) xml.Name {
	return xml.Name{
		Local: "rpc",
		Space: namespace,
	}
}

// XMLAttr returns a slice of xml.Attr containing only
// one xml.Attr for the RPC's message-id.
func XMLAttr(messageID string) []xml.Attr {
	return []xml.Attr{
		{
			Name: xml.Name{
				Local: "message-id",
			},
			Value: messageID,
		},
	}
}

// WrapMethod wraps the given methods' with outer rpc
// tags, and sets default values for namespace and
// message id attributes. It returns a pointer to a
// Method that can be directly marshaled into an RPC
// by Encoder.
func WrapMethod(method ...interface{}) *Method {
	GlobalCounter.Add(1)
	return &Method{
		XMLName: XMLNameTag(BaseNamespace),
		Attr:    XMLAttr(GlobalCounter.String()),
		Method:  method,
	}
}

// Encoder embeds an xml.Encoder, but overrides Encode
// with a custom implementation designed specifically
// to encode NETCONF RPC requests.
type Encoder struct {
	*xml.Encoder
	bufWriter *bufio.Writer
}

// NewEncoder buffers the given io.Writer, and wraps it
// into a Encoder.
func NewEncoder(w io.Writer) *Encoder {

	var e Encoder

	e.bufWriter = bufio.NewWriter(w)
	e.Encoder = xml.NewEncoder(e.bufWriter)

	return &e
}

// EncodeHello writes the given hello message to the
// underlying writer, writes a message separator, and
// flushes the buffer.
func (e *Encoder) EncodeHello(h *HelloMessage) error {

	if err := e.Encoder.Encode(h); err != nil {
		return err
	} else if err = e.WriteSep(); err != nil {
		return err
	}

	return nil
}

// Encode encodes a single NETCONF RPC, and marshals it
// into the underlying buffer. Then it writes the NETCONF message
// separator followed by a newline, and flushes the buffer.
//
// Encoding XML as a stream of tokens is still possible using the
// underlying xml.Encoder. However, WriteSep must should be called
// after encoding an RPC.
func (e *Encoder) Encode(v interface{}) error {

	method, ok := v.(*Method)
	if !ok {
		method = WrapMethod(v)
	}

	if err := e.Encoder.Encode(method); err != nil {
		return err
	} else if err = e.WriteSep(); err != nil {
		return err
	}

	return nil
}

// WriteSep writes a message separator with a trailing newline to
// the underlying buffered io.Writer, and flushes the buffer before
// returning. Using this method is only necessary when manually
// encoding XML tokens as a stream with EncodeToken, et al.
//
// Calls to WriteSep may block depending on the underlying net.Conn.
//
// Most uses will call Encode, which calls WriteSep internally.
func (e *Encoder) WriteSep() error {

	if _, err := e.bufWriter.Write(messageSeparatorBytes); err != nil {
		return err
	} else if err = e.bufWriter.WriteByte('\n'); err != nil {
		return err
	} else if err = e.bufWriter.Flush(); err != nil {
		return err
	}

	return nil
}

// Marshal returns the NETCONF encoding of v, including message
// separators and enclosing RPC tags.
//
// If the argument's type is *Method, MarshalRPC just calls xml.Marshal
// internally to build the XML. Otherwise, it wraps its argument in the
// default *Method before calling xml.Marshal.
//
// A NETCONF message separator and newline is always written to the end
// of the message.
func Marshal(v interface{}) ([]byte, error) {

	var b bytes.Buffer
	if err := NewEncoder(&b).Encode(v); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}
