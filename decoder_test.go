package netconf

import (
	"bytes"
	"encoding/xml"
	"reflect"
	"strings"
	"testing"
)

func TestDecoder_DecodeHello(t *testing.T) {

	type DecoderTest struct {
		ResetVal string
		Reader   *strings.Reader
		Want     []interface{}
	}

	dTestTable := []DecoderTest{
		{
			ResetVal: `<?xml version="1.0" encoding="UTF-8"?>
<hello xmlns="urn:ietf:params:xml:ns:netconf:base:1.0">
<capabilities>
<capability>urn:ietf:params:netconf:base:1.1</capability>
<capability>urn:ietf:params:ns:netconf:capability:startup:1.0</capability>
</capabilities>
<session-id>4</session-id>
</hello>
]]>]]>
`,
			Reader: strings.NewReader(`<?xml version="1.0" encoding="UTF-8"?>
<hello xmlns="urn:ietf:params:xml:ns:netconf:base:1.0">
<capabilities>
<capability>urn:ietf:params:netconf:base:1.1</capability>
<capability>urn:ietf:params:ns:netconf:capability:startup:1.0</capability>
</capabilities>
<session-id>4</session-id>
</hello>
]]>]]>
`),
			Want: []interface{}{
				&HelloMessage{
					XMLName: xml.Name{
						Local: "hello",
						Space: BaseNamespace,
					},
					SessionID: 4,
					Capabilities: []string{
						"urn:ietf:params:netconf:base:1.1",
						"urn:ietf:params:ns:netconf:capability:startup:1.0",
					},
				},
			},
		},
		{
			ResetVal: `<?xml version="1.0" encoding="UTF-8"?>
<hello xmlns="urn:ietf:params:xml:ns:netconf:base:1.0">
<capabilities>
<capability>urn:ietf:params:netconf:base:1.1</capability>
<capability>urn:ietf:params:ns:netconf:capability:startup:1.0</capability>
</capabilities>
<session-id>4</session-id>
</hello>
]]>]]>


<?xml version="1.0" encoding="UTF-8"?>
<hello xmlns="urn:ietf:params:xml:ns:netconf:base:1.0">
<capabilities>
<capability>urn:ietf:params:netconf:base:1.1</capability>
<capability>urn:ietf:params:ns:netconf:capability:startup:1.0</capability>
</capabilities>
<session-id>4</session-id>
</hello>
]]>]]>
`,
			Reader: strings.NewReader(`<?xml version="1.0" encoding="UTF-8"?>
<hello xmlns="urn:ietf:params:xml:ns:netconf:base:1.0">
<capabilities>
<capability>urn:ietf:params:netconf:base:1.1</capability>
<capability>urn:ietf:params:ns:netconf:capability:startup:1.0</capability>
</capabilities>
<session-id>4</session-id>
</hello>
]]>]]>


<?xml version="1.0" encoding="UTF-8"?>
<hello xmlns="urn:ietf:params:xml:ns:netconf:base:1.0">
<capabilities>
<capability>urn:ietf:params:netconf:base:1.1</capability>
<capability>urn:ietf:params:ns:netconf:capability:startup:1.0</capability>
</capabilities>
<session-id>4</session-id>
</hello>
]]>]]>
`),
			Want: []interface{}{
				&HelloMessage{
					XMLName: xml.Name{
						Local: "hello",
						Space: BaseNamespace,
					},
					SessionID: 4,
					Capabilities: []string{
						"urn:ietf:params:netconf:base:1.1",
						"urn:ietf:params:ns:netconf:capability:startup:1.0",
					},
				},
				&HelloMessage{
					XMLName: xml.Name{
						Local: "hello",
						Space: BaseNamespace,
					},
					SessionID: 4,
					Capabilities: []string{
						"urn:ietf:params:netconf:base:1.1",
						"urn:ietf:params:ns:netconf:capability:startup:1.0",
					},
				},
			},
		},
	}

	for i, test := range dTestTable {
		dec := NewDecoder(test.Reader)
		for j, want := range test.Want {
			var hello HelloMessage
			if err := dec.DecodeHello(&hello); err != nil {
				t.Error(err)
			} else if !reflect.DeepEqual(&hello, want) {
				t.Errorf("trying test %d: structs don't match\nwant:\t%v\ngot:\t%v",
					i, want, &hello)
			} else {
				t.Logf("test %d subtest %d successful", i, j)
			}
		}
	}
}

func BenchmarkDecoder_DecodeHello(b *testing.B) {

	helloBytes := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<hello xmlns="urn:ietf:params:xml:ns:netconf:base:1.0">
<capabilities>
<capability>urn:ietf:params:netconf:base:1.1</capability>
<capability>urn:ietf:params:ns:netconf:capability:startup:1.0</capability>
</capabilities>
<session-id>4</session-id>
</hello>
]]>]]>
`)

	var hello HelloMessage

	r := bytes.NewReader(helloBytes)
	dec := NewDecoder(r)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {

		if err := dec.DecodeHello(&hello); err != nil {
			b.Error(err)
		}

		r.Reset(helloBytes)
	}
}

func BenchmarkDecoder_Decode(b *testing.B) {

	lldpNbrsRPCReplyBytes := []byte(`<rpc-reply xmlns="urn:ietf:params:xml:ns:netconf:base:1.0" xmlns:junos="http://xml.juniper.net/junos/15.1X49/junos">
<lldp-neighbors-information junos:style="brief">
<lldp-neighbor-information>
<lldp-local-port-id>ge-0/0/7</lldp-local-port-id>
<lldp-local-parent-interface-name>-</lldp-local-parent-interface-name>
<lldp-remote-chassis-id-subtype>Mac address</lldp-remote-chassis-id-subtype>
<lldp-remote-chassis-id>f0:1c:2d:ed:68:80</lldp-remote-chassis-id>
<lldp-remote-port-description>ge-0/0/0.0</lldp-remote-port-description>
<lldp-remote-system-name>EX2200C2</lldp-remote-system-name>
</lldp-neighbor-information>
</lldp-neighbors-information>
</rpc-reply>
]]>]]>
`)
	type Neighbor struct {
		LocalInterface         string `xml:"lldp-local-interface,omitempty"`
		LocalParentInterface   string `xml:"lldp-local-parent-interface-name,omitempty"`
		LocalPortID            string `xml:"lldp-local-port-id,omitempty"`
		RemoteChassisIDSubtype string `xml:"lldp-remote-chassis-id-subtype,omitempty"`
		RemoteChassisID        string `xml:"lldp-remote-chassis-id,omitempty"`
		RemotePortIDSubtype    string `xml:"lldp-remote-port-id-subtype,omitempty"`
		RemotePortID           string `xml:"lldp-remote-port-id,omitempty"`
		RemotePortDesc         string `xml:"lldp-remote-port-description,omitempty"`
		RemoteSystemName       string `xml:"lldp-remote-system-name,omitempty"`
	}

	type LLDPReply struct {
		Neighbor []Neighbor `xml:"lldp-neighbor-information"`
	}

	var reply = Reply{
		Data: &LLDPReply{},
	}

	r := bytes.NewReader(lldpNbrsRPCReplyBytes)
	dec := NewDecoder(r)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {

		if err := dec.Decode(&reply); err != nil {
			b.Error(err)
		}

		r.Reset(lldpNbrsRPCReplyBytes)
	}
}

func TestReply_Unmarshal(t *testing.T) {

	type Neighbor struct {
		LocalInterface         string `xml:"lldp-local-interface,omitempty"`
		LocalParentInterface   string `xml:"lldp-local-parent-interface-name,omitempty"`
		LocalPortID            string `xml:"lldp-local-port-id,omitempty"`
		RemoteChassisIDSubtype string `xml:"lldp-remote-chassis-id-subtype,omitempty"`
		RemoteChassisID        string `xml:"lldp-remote-chassis-id,omitempty"`
		RemotePortIDSubtype    string `xml:"lldp-remote-port-id-subtype,omitempty"`
		RemotePortID           string `xml:"lldp-remote-port-id,omitempty"`
		RemotePortDesc         string `xml:"lldp-remote-port-description,omitempty"`
		RemoteSystemName       string `xml:"lldp-remote-system-name,omitempty"`
	}

	type LLDPReply struct {
		Neighbor []Neighbor `xml:"lldp-neighbor-information"`
	}

	lldpNbrsRPCReplyBytes := []byte(`<rpc-reply xmlns="urn:ietf:params:xml:ns:netconf:base:1.0" xmlns:junos="http://xml.juniper.net/junos/15.1X49/junos">
<lldp-neighbors-information junos:style="brief">
<lldp-neighbor-information>
<lldp-local-port-id>ge-0/0/7</lldp-local-port-id>
<lldp-local-parent-interface-name>-</lldp-local-parent-interface-name>
<lldp-remote-chassis-id-subtype>Mac address</lldp-remote-chassis-id-subtype>
<lldp-remote-chassis-id>f0:1c:2d:ed:68:80</lldp-remote-chassis-id>
<lldp-remote-port-description>ge-0/0/0.0</lldp-remote-port-description>
<lldp-remote-system-name>EX2200C2</lldp-remote-system-name>
</lldp-neighbor-information>
</lldp-neighbors-information>
</rpc-reply>
]]>]]>
`)

	wantVal := &Neighbor{
		LocalInterface:         "",
		LocalParentInterface:   "-",
		LocalPortID:            "ge-0/0/7",
		RemoteChassisIDSubtype: "Mac address",
		RemoteChassisID:        "f0:1c:2d:ed:68:80",
		RemotePortIDSubtype:    "",
		RemotePortID:           "",
		RemotePortDesc:         "ge-0/0/0.0",
		RemoteSystemName:       "EX2200C2",
	}

	var reply LLDPReply
	//var gotVal Reply
	//gotVal.Data = &reply

	if err := Unmarshal(lldpNbrsRPCReplyBytes, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(wantVal, &reply.Neighbor[0]) {
		t.Errorf("unexpected lldp unmarshal values:\nwant:\t%v\ngot:\t%v",
			wantVal, &reply.Neighbor[0])
	} else {
		t.Log("rpc reply wrapped successfully")
	}

	type NewNeighbor struct {
		LocalInterface         string `xml:"lldp-neighbor-information>lldp-local-interface,omitempty"`
		LocalParentInterface   string `xml:"lldp-neighbor-information>lldp-local-parent-interface-name,omitempty"`
		LocalPortID            string `xml:"lldp-neighbor-information>lldp-local-port-id,omitempty"`
		RemoteChassisIDSubtype string `xml:"lldp-neighbor-information>lldp-remote-chassis-id-subtype,omitempty"`
		RemoteChassisID        string `xml:"lldp-neighbor-information>lldp-remote-chassis-id,omitempty"`
		RemotePortIDSubtype    string `xml:"lldp-neighbor-information>lldp-remote-port-id-subtype,omitempty"`
		RemotePortID           string `xml:"lldp-neighbor-information>lldp-remote-port-id,omitempty"`
		RemotePortDesc         string `xml:"lldp-neighbor-information>lldp-remote-port-description,omitempty"`
		RemoteSystemName       string `xml:"lldp-neighbor-information>lldp-remote-system-name,omitempty"`
	}

	newNbrWantVal := &NewNeighbor{
		LocalInterface:         "",
		LocalParentInterface:   "-",
		LocalPortID:            "ge-0/0/7",
		RemoteChassisIDSubtype: "Mac address",
		RemoteChassisID:        "f0:1c:2d:ed:68:80",
		RemotePortIDSubtype:    "",
		RemotePortID:           "",
		RemotePortDesc:         "ge-0/0/0.0",
		RemoteSystemName:       "EX2200C2",
	}

	var newNbrSlice []NewNeighbor
	if err := Unmarshal(lldpNbrsRPCReplyBytes, &newNbrSlice); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(newNbrWantVal, &newNbrSlice[0]) {
		t.Errorf("unexpected lldp unmarshal values:\nwant:\t%v\ngot:\t%v",
			newNbrWantVal, &newNbrSlice[0])
	} else {
		t.Log("rpc reply wrapped successfully")
	}
}

func TestReply_UnmarshalOk(t *testing.T) {
	okReplyBytes1 := []byte(`<rpc-reply xmlns="urn:ietf:params:xml:ns:netconf:base:1.0" xmlns:junos="http://xml.juniper.net/junos/15.1X49/junos">
<ok/>
</rpc-reply>
]]>]]>
`)

	var okReplyObj1 Reply
	if err := Unmarshal(okReplyBytes1, &okReplyObj1); err != nil {
		t.Error(err)
	} else if okReplyObj1.Ok == nil {
		t.Errorf("unexpected reply ok value:\nwant:\t%t\ngot:\t%t", true, okReplyObj1.Ok != nil)
	}

	okReplyBytes2 := []byte(`<rpc-reply xmlns="urn:ietf:params:xml:ns:netconf:base:1.0" xmlns:junos="http://xml.juniper.net/junos/15.1X49/junos">
</rpc-reply>
]]>]]>
`)

	var okReplyObj2 Reply
	if err := Unmarshal(okReplyBytes2, &okReplyObj2); err != nil {
		t.Error(err)
	} else if okReplyObj2.Ok != nil {
		t.Errorf("unexpected reply ok value:\nwant:\t%t\ngot:\t%t", false, okReplyObj2.Ok != nil)
	}
}
