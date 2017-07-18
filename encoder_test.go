package netconf

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"testing"
)

func TestEncoder_Encode(t *testing.T) {

	type ShowInterfacesRPC struct {
		XMLName xml.Name  `xml:"get-interface-information"`
		Detail  *struct{} `xml:"detail,omitempty"`
	}

	var buf bytes.Buffer
	showIfaceRPC := ShowInterfacesRPC{
		Detail: &struct{}{},
	}

	want := []byte(`<rpc xmlns="urn:ietf:params:xml:ns:netconf:base:1.0" message-id="1"><get-interface-information><detail></detail></get-interface-information></rpc>]]>]]>
`)

	if err := NewEncoder(&buf).Encode(WrapMethod(&showIfaceRPC)); err != nil {
		t.Error(err)
	} else if !bytes.Equal(want, buf.Bytes()) {
		t.Logf("unexpected bytes decoded\nwant:\t%q\ngot:\t%q", want, buf.Bytes())
	} else {
		t.Log("successfully marshalled get-interface-information rpc")
	}
}

func BenchmarkEncoder_Encode(b *testing.B) {

	type ShowInterfacesRPC struct {
		XMLName xml.Name  `xml:"get-interface-information"`
		Detail  *struct{} `xml:"detail,omitempty"`
	}

	var buf bytes.Buffer
	showIfaceRPC := ShowInterfacesRPC{
		Detail: &struct{}{},
	}

	enc := NewEncoder(&buf)
	if err := NewEncoder(&buf).Encode(WrapMethod(&showIfaceRPC)); err != nil {
		b.Error(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := enc.Encode(WrapMethod(&showIfaceRPC)); err != nil {
			b.Error(err)
		}
	}
}

func Test_Marshal(t *testing.T) {

	type ShowInterfacesRPC struct {
		XMLName xml.Name  `xml:"get-interface-information"`
		Detail  *struct{} `xml:"detail,omitempty"`
	}

	showIfaceRPC := ShowInterfacesRPC{
		Detail: &struct{}{},
	}

	b, err := Marshal(showIfaceRPC)
	if err != nil {
		t.Error(err)
	}

	want := fmt.Sprintf(`<rpc xmlns="urn:ietf:params:xml:ns:netconf:base:1.0" message-id="%d"><get-interface-information><detail></detail></get-interface-information></rpc>]]>]]>
`, GlobalCounter.Value())

	if wantBytes := []byte(want); !bytes.Equal(wantBytes, b) {
		t.Errorf("unexpected bytes decoded\nwant:\t%q\ngot:\t%q", want, b)
	} else {
		t.Log("successfully marshalled get-interface-information rpc")
	}
}
