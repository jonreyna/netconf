package netconf

import "encoding/xml"

const (
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
