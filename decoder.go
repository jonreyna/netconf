package netconf

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"io"
)

// Reply models the structure of a NETCONF reply.
// It is useful for wrapping structs that don't
// decode the outer rpc-reply tags.
type Reply struct {
	XMLName xml.Name     `xml:"rpc-reply"`
	Attr    []xml.Attr   `xml:",attr"`
	Ok      *struct{}    `xml:"ok"`
	Error   []ReplyError `xml:"rpc-error"`
	Data    interface{}  `xml:",any"`
}

// Decoder embeds an xml.Decoder, but overrides Decode
// with a custom implementation designed specifically
// to decode NETCONF RPC replies.
type Decoder struct {
	*xml.Decoder
	bufReader *bufio.Reader
}

// NewDecoder buffers the given io.Reader, and wraps it
// in a Decoder.
func NewDecoder(r io.Reader) *Decoder {

	var d Decoder

	d.bufReader = bufio.NewReader(r)
	d.Decoder = xml.NewDecoder(d.bufReader)

	return &d
}

// DecodeHello handles hello/capabilities messages sent by
// the NETCONF server. It's a special decode case since the
// closing tags are named "hello" rather than "rpc-reply".
func (d *Decoder) DecodeHello(h *HelloMessage) error {

	if err := d.Decoder.Decode(h); err != nil {
		return err
	} else if err = d.SkipSep(); err != nil {
		return err
	}

	return nil
}

// Decode wraps the interface{} parameter in a Reply object
// to capture all of the RPC Reply content. It also searches
// for errors in the Reply, and returns the first ReplyError
// found.
// as a standard error interface.
//
//
// unmarshals a single NETCONF RPC reply message into
// the given interface{}.
//
// A full RPC Reply can be obtained
// Parsing XML as a stream of tokens is still possible using
// the embedded xml.Decoder. However, Done should be called
// finished to discard the NETCONF message separator.
func (d *Decoder) Decode(v interface{}) error {

	reply, ok := v.(*Reply)
	if !ok {
		// wrap in a standard RPC Reply for proper decoding
		reply = &Reply{
			Data: v,
		}
	}

	if err := d.Decoder.Decode(reply); err != nil {
		return err
	}

	// TODO: Consider returning here if the caller provided a Reply

	for i, err := range reply.Error {
		if err.Severity == ErrorSeverityError {
			return &reply.Error[i]
		}
	}

	return nil
}

// messageSeparatorBytes is a micro-optimization that eliminates the
// need to create a new byte slice every time we search for the NETCONF
// message message separator.
var messageSeparatorBytes = []byte(MessageSeparator)

// SkipSep discarding everything from the underlying buffer until it
// encounters a NETCONF message separator, or an error.
//
// Since the separator is explicitly designed to be invalid XML,
// failure to discard it before decoding will cause the standard
// decoder to fail with a syntax error.
//
// Using this method is only necessary when manually decoding XML
// tokens as a stream, with DecodeToken, et al.
//
// Calls to SkipSep may block if more bytes have to be read from
// the underlying net.Conn.
//
// Most uses will call Decode, which calls SkipSep internally.
func (d *Decoder) SkipSep() error {

	for {
		if s, err := d.bufReader.ReadSlice('\n'); err != nil && err != bufio.ErrBufferFull {
			return err
		} else if bytes.Equal(bytes.TrimSpace(s), messageSeparatorBytes) {
			break
		}
	}

	return nil
}

// Unmarshal maps the NETCONF RPC reply XML into the given argument,
// discarding the terminating message separator.
func Unmarshal(data []byte, v interface{}) error {
	return NewDecoder(bytes.NewReader(data)).Decode(v)
}
