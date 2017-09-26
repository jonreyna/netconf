package netconf

import (
	"reflect"
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
		t.Errorf("erroSeverityStringArray is NOT sorted!\nwant:\t%q\ngot:\t%q",
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
	if err := Unmarshal([]byte(err1), &reply1); err.Error() != "error unknown-element pbr" {
		t.Errorf("unexpected error unmarshalling reply: %v", err)
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

func TestErrorSeverity_UnmarshalText(t *testing.T) {
	tests := []struct {
		ErrorSeverityText []byte
		WantErrorSeverity ErrorSeverity
		WantError         error
	}{
		{
			ErrorSeverityText: []byte(""),
			WantErrorSeverity: ErrorSeverityZero,
		},
		{
			ErrorSeverityText: []byte("error"),
			WantErrorSeverity: ErrorSeverityError,
		},
		{
			ErrorSeverityText: []byte("unknown"),
			WantErrorSeverity: ErrorSeverityUnknown,
		},
		{
			ErrorSeverityText: []byte("warning"),
			WantErrorSeverity: ErrorSeverityWarning,
		},
		{
			ErrorSeverityText: []byte("    unknown"),
			WantErrorSeverity: ErrorSeverityUnknown,
		},
		{
			ErrorSeverityText: []byte("warning      "),
			WantErrorSeverity: ErrorSeverityWarning,
		},
		{
			ErrorSeverityText: []byte(" error      "),
			WantErrorSeverity: ErrorSeverityError,
		},
		{
			ErrorSeverityText: []byte("sadf d error      "),
			WantErrorSeverity: ErrorSeverityUnknown,
			WantError:         &UnmarshalTextError{Type: "ErrorSeverity", Value: "sadf d error      "},
		},
		{
			ErrorSeverityText: []byte("errora"),
			WantErrorSeverity: ErrorSeverityUnknown,
			WantError:         &UnmarshalTextError{Type: "ErrorSeverity", Value: "errora"},
		},
	}

	for i, test := range tests {
		var es ErrorSeverity
		if err := es.UnmarshalText(test.ErrorSeverityText); err != nil {
			if !reflect.DeepEqual(err, test.WantError) {
				t.Errorf("unexpected error returned from UnmarshalText on test %d\nwant:\t%v\ngot:\t%v",
					i, test.WantError, err)
			}
		} else if es != test.WantErrorSeverity {
			t.Errorf("unexpected ErrorSeverity returned from UnmarshalText on test %d\nwant:\t%q\ngot:\t%q",
				i, test.WantErrorSeverity, es)
		}
	}
}

func TestErrorType_UnmarshalText(t *testing.T) {
	tests := []struct {
		ErrorTypeText []byte
		WantErrorType ErrorType
		WantError     error
	}{
		{
			ErrorTypeText: []byte(""),
			WantErrorType: ErrorTypeZero,
		},
		{
			ErrorTypeText: []byte("application"),
			WantErrorType: ErrorTypeApplication,
		},
		{
			ErrorTypeText: []byte("protocol"),
			WantErrorType: ErrorTypeProtocol,
		},
		{
			ErrorTypeText: []byte("rpc"),
			WantErrorType: ErrorTypeRPC,
		},
		{
			ErrorTypeText: []byte("transport"),
			WantErrorType: ErrorTypeTransport,
		},
		{
			ErrorTypeText: []byte("unknown"),
			WantErrorType: ErrorTypeUnknown,
		},
		{
			ErrorTypeText: []byte("   "),
			WantErrorType: ErrorTypeZero,
		},
		{
			ErrorTypeText: []byte(" rpc  "),
			WantErrorType: ErrorTypeRPC,
		},
		{
			ErrorTypeText: []byte("rpc  "),
			WantErrorType: ErrorTypeRPC,
		},
		{
			ErrorTypeText: []byte("      transport"),
			WantErrorType: ErrorTypeTransport,
		},
		{
			ErrorTypeText: []byte("stransport"),
			WantErrorType: ErrorTypeUnknown,
			WantError:     &UnmarshalTextError{Type: "ErrorType", Value: "stransport"},
		},
		{
			ErrorTypeText: []byte("  rpcc"),
			WantErrorType: ErrorTypeUnknown,
			WantError:     &UnmarshalTextError{Type: "ErrorType", Value: "  rpcc"},
		},
		{
			ErrorTypeText: []byte("unknown  "),
			WantErrorType: ErrorTypeUnknown,
		},
	}

	for i, test := range tests {
		var et ErrorType
		if err := et.UnmarshalText(test.ErrorTypeText); err != nil {
			if !reflect.DeepEqual(err, test.WantError) {
				t.Errorf("unexpected error returned from UnmarshalText on test %d\nwant:\t%v\ngot:\t%v",
					i, test.WantError, err)
			}
		} else if et != test.WantErrorType {
			t.Errorf("unexpected ErrorType returned from UnmarshalText on test %d\nwant:\t%q\ngot:\t%q",
				i, test.WantErrorType, et)
		}
	}
}

func TestErrorTag_UnmarshalText(t *testing.T) {

	tests := []struct {
		ErrorTagText []byte
		WantErrorTag ErrorTag
		WantError    error
	}{
		{
			ErrorTagText: []byte(""),
			WantErrorTag: ErrorTagZero,
		},
		{
			ErrorTagText: []byte("bad-attribute"),
			WantErrorTag: ErrorTagBadAttribute,
		},
		{
			ErrorTagText: []byte("lock-denied"),
			WantErrorTag: ErrorTagLockDenied,
		},
		{
			ErrorTagText: []byte("operation-failed"),
			WantErrorTag: ErrorTagOpFailed,
		},
		{
			ErrorTagText: []byte("resource-denied"),
			WantErrorTag: ErrorTagResourceDenied,
		},
		{
			ErrorTagText: []byte("unknown"),
			WantErrorTag: ErrorTagUnknown,
		},
		{
			ErrorTagText: []byte("   "),
			WantErrorTag: ErrorTagZero,
		},
		{
			ErrorTagText: []byte("  too-big"),
			WantErrorTag: ErrorTagTooBig,
		},
		{
			ErrorTagText: []byte("malformed-message     "),
			WantErrorTag: ErrorTagMalformedMessage,
		},
		{
			ErrorTagText: []byte("    in-use      "),
			WantErrorTag: ErrorTagInUse,
		},
		{
			ErrorTagText: []byte("ƢƦƴǼ"),
			WantErrorTag: ErrorTagUnknown,
			WantError:    &UnmarshalTextError{Type: "ErrorTag", Value: "ƢƦƴǼ"},
		},
		{
			ErrorTagText: []byte("    0xDEADBEEFCAFE"),
			WantErrorTag: ErrorTagUnknown,
			WantError:    &UnmarshalTextError{Type: "ErrorTag", Value: "    0xDEADBEEFCAFE"},
		},
		{
			ErrorTagText: []byte(" i n - u s e "),
			WantErrorTag: ErrorTagUnknown,
			WantError:    &UnmarshalTextError{Type: "ErrorTag", Value: " i n - u s e "},
		},
	}

	for i, test := range tests {
		var et ErrorTag
		if err := et.UnmarshalText(test.ErrorTagText); err != nil {
			if !reflect.DeepEqual(err, test.WantError) {
				t.Errorf("unexpected error returned from UnmarshalText on test %d\nwant:\t%v\ngot:\t%v",
					i, test.WantError, err)
			}
		} else if et != test.WantErrorTag {
			t.Errorf("unexpected ErrorTag returned from UnmarshalText on test %d\nwant:\t%q\ngot:\t%q",
				i, test.WantErrorTag, et)
		}
	}
}
