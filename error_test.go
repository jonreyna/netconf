package netconf

import (
	"sort"
	"testing"
)

func TestErrorTagStringArray_IsSorted(t *testing.T) {
	// error tag string optimizations require errorTagStringArray to be sorted
	if isSorted := sort.StringsAreSorted(errorTagStringArray[:]); !isSorted {
		sortedStrs := errorTagStringArray
		sort.Strings(sortedStrs[:])
		t.Errorf("errorTagStringArray is NOT sorted!\nwant:\t%q\ngot:\t%q",
			sortedStrs, errorTagStringArray[:])
	} else {
		t.Log("errorTagStringArray is sorted")
	}
}

func TestErrorSeverityStringArray_IsSorted(t *testing.T) {
	// error severity string optimizations require errorSeverityStringArray to be sorted
	if isSorted := sort.StringsAreSorted(errorSeverityStringArray[:]); !isSorted {
		sortedStrs := errorSeverityStringArray
		sort.Strings(sortedStrs[:])
		t.Errorf("errorSeverityStringArray is NOT sorted!\nwant:\t%q\ngot:\t%q",
			sortedStrs, errorSeverityStringArray[:])
	} else {
		t.Log("errorSeverityStringArray is sorted")
	}
}

func TestErrorTypeStringArray_IsSorted(t *testing.T) {
	// parsing an error type relies on errorTypeStringArray being sorted
	if isSorted := sort.StringsAreSorted(errorTypeStringArray[:]); !isSorted {
		sortedStrs := errorTypeStringArray
		sort.Strings(sortedStrs[:])
		t.Errorf("errorTypeStringArray is NOT sorted!\nwant:\t%q\ngot:\t%q",
			sortedStrs, errorTypeStringArray[:])
	} else {
		t.Log("errorTypeStringArray is sorted")
	}
}

func TestError_Unmarshal(t *testing.T) {
	const err1 = `<rpc-reply xmlns="urn:ietf:params:xml:ns:netconf:base:1.0" message-id="101">
<rpc-error>
<error-type>protocol</error-type>
<error-tag>unknown-element</error-tag>
<error-severity>error</error-severity>
<error-path xmlns:ns1="http://cisco.com/ns/yang/Cisco-IOS-XR-pbr-cfg" xmlns:ns2="http://cisco.com/ns/yang/Cisco-IOS-XR-ifmgr-cfg">ns2:interface-configurations/ns2:interface-configuration/ns1:pbr</error-path>
<error-info>
<bad-element>pbr</bad-element>
</error-info>
</rpc-error>
</rpc-reply>
]]>]]>
`

	var reply1 Reply
	if err := Unmarshal([]byte(err1), &reply1); err != nil {
		t.Error(err)
	} else if reply1.Error[0].Type != ErrorTypeProtocol {
		t.Errorf("unexpected error type:\nwant:\t%q\ngot:\t%q",
			ErrorTypeProtocol, reply1.Error[0].Type)
	} else if reply1.Error[0].Tag != ErrorTagUnknownElement {
		t.Errorf("unexpected error tag:\nwant:\t%q\ngot:\t%q",
			ErrorTagUnknownElement, reply1.Error[0].Tag)
	} else if reply1.Error[0].Severity != ErrorSeverityError {
		t.Errorf("unexpected error severity:\nwant:\t%q\ngot:\t%q",
			ErrorSeverityError, reply1.Error[0].Tag)
	} else if want := "ns2:interface-configurations/ns2:interface-configuration/ns1:pbr"; want != reply1.Error[0].Path {
		t.Errorf("unexpected error path:\nwant:\t%q\ngot:\t%q",
			want, reply1.Error[0].Tag)
	} else if want := "pbr"; want != reply1.Error[0].Info.BadElement {
		t.Errorf("unexpected error path:\nwant:\t%q\ngot:\t%q",
			want, reply1.Error[0].Info.BadElement)
	}

}
