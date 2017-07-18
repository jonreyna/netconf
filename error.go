package netconf

import (
	"bytes"
	"fmt"
	"sort"
)

// ErrorSeverity identifies severity of the error as either warning or error.
type ErrorSeverity uint64

const (
	// ErrorSeverityZero represents an uninitialized ErrorSeverity value.
	ErrorSeverityZero ErrorSeverity = iota
	// ErrorSeverityError indicates the severity is on the error level.
	ErrorSeverityError
	// ErrorSeverityUnknown means the ErrorSeverity could not be identified, and may indicate an internal error.
	ErrorSeverityUnknown
	// ErrorSeverityWarning is not yet utilized, according to RFC 6241.
	ErrorSeverityWarning
)

var errorSeverityStringArray = [...]string{
	ErrorSeverityZero:    "",
	ErrorSeverityError:   "error",
	ErrorSeverityUnknown: "unknown",
	ErrorSeverityWarning: "warning",
}

// String returns a string representing the ErrorSeverity.
// If the ErrorSeverity is not known for some erroneous reason,
// the String will return "unknown".
func (es ErrorSeverity) String() string {
	if int(es) < len(errorSeverityStringArray) {
		return errorSeverityStringArray[es]
	}
	return errorSeverityStringArray[ErrorSeverityUnknown]
}

// UnmarshalText satisfies the encoding.TextUnmarshaler interface.
func (es *ErrorSeverity) UnmarshalText(text []byte) error {
	sevSlice := errorSeverityStringArray[:]
	if i := sort.SearchStrings(sevSlice, string(bytes.TrimSpace(text))); i != len(sevSlice) {
		*es = ErrorSeverity(uint64(i))
		return nil
	}
	*es = ErrorSeverityUnknown
	return fmt.Errorf("netconf: unknown ErrorSeverity parsing %q", text)
}

// ErrorInfo contains protocol or data model specific error content.
type ErrorInfo struct {
	BadAttribute string `xml:"bad-attribute"` // BadAttribute is the name of the bad, missing, or unexpected attribute.
	BadElement   string `xml:"bad-element"`   // BadElement is the name of the element containing the bad, missing or unexpected attribute or element.
	OkElement    string `xml:"ok-element"`    // OkElement is the parent element for which all children have completed the requested operation.
	ErrElement   string `xml:"err-element"`   // ErrElement is the parent element for which all children have failed to complete the requested operation.
	NoopElement  string `xml:"noop-element"`  // NoopElement is the parent element that identifies all children for which the requested operation was not attempted.
	BadNamespace string `xml:"bad-namespace"` // BadNamespace contains the name of the unexpected
	SessionID    uint64 `xml:"session-id"`
}

// ErrorType defines the conceptual layer that the error occurred.
type ErrorType uint64

const (
	// ErrorTypeZero represents an uninitialized ErrorType value.
	ErrorTypeZero ErrorType = iota
	// ErrorTypeApplication indicates the error occurred on the Content layer.
	ErrorTypeApplication
	// ErrorTypeProtocol indicates the error occurred on the Operations layer, which defines a set of base protocol operations invoked as RPC methods.
	ErrorTypeProtocol
	// ErrorTypeRPC indicates the error occurred on the Messages layer: the transport-independent framing mechanism for encoding RPCs and notifications.
	ErrorTypeRPC
	// ErrorTypeTransport indicates the error occurred on the Secure Transport layer, which provides a communication path between the client and server.
	ErrorTypeTransport
	// ErrorTypeUnknown indicates an unexpected condition.
	ErrorTypeUnknown
)

var errorTypeStringArray = [...]string{
	ErrorTypeZero:        "",
	ErrorTypeApplication: "application",
	ErrorTypeProtocol:    "protocol",
	ErrorTypeRPC:         "rpc",
	ErrorTypeTransport:   "transport",
	ErrorTypeUnknown:     "unknown",
}

// String satisfies the Stringer interface.
func (es ErrorType) String() string {
	if int(es) < len(errorTypeStringArray) {
		return errorTypeStringArray[es]
	}
	return errorTypeStringArray[ErrorTypeUnknown]
}

// UnmarshalText satisfies the encoding.TextUnmarshaler interface.
func (es *ErrorType) UnmarshalText(text []byte) error {
	errTypeSlice := errorTypeStringArray[:]
	if i := sort.SearchStrings(errTypeSlice, string(bytes.TrimSpace(text))); i != len(errTypeSlice) {
		*es = ErrorType(uint64(i))
		return nil
	}
	*es = ErrorTypeUnknown
	return fmt.Errorf("netconf: unknown ErrorType parsing %q", text)
}

// ErrorTag identifies the error condition.
type ErrorTag uint64

const (
	// ErrorTagZero is an uninitialized ErrorTag value.
	ErrorTagZero ErrorTag = iota
	// ErrorTagAccessDenied indicates access to the requested protocol operation or data is denied because authorization failed.
	ErrorTagAccessDenied
	// ErrorTagBadAttribute indicates an attribute value is not correct; e.g., wrong type, out of range, pattern mismatch.
	ErrorTagBadAttribute
	// ErrorTagBadElement indicates an element value is not correct. ErrorInfo's BadElement field will contain the element with a bad value's name.
	ErrorTagBadElement
	// ErrorTagDataExists indicates the request could not be completed because the relevant data model content already exists. For example, a "create" operation was attempted on data that already exists.
	ErrorTagDataExists
	// ErrorTagDataMissing indicates the request could not be completed because the relevant data model content does not exist. For example, a "delete" operation was attempted on data that does not exist.
	ErrorTagDataMissing
	// ErrorTagInUse indicates the request requires a resource that is already in use.
	ErrorTagInUse
	// ErrorTagInvalidValue indicates the request specifies an unacceptable value for one or more parameters.
	ErrorTagInvalidValue
	// ErrorTagLockDenied indicates access to the requested lock is denied because the lock is currently held by another entity.
	ErrorTagLockDenied
	// ErrorTagMalformedMessage indicates a message could not be handled because it failed to be parsed correctly. For example, the message is not well-formed XML, or uses an invalid character set.
	ErrorTagMalformedMessage
	// ErrorTagMissingAttribute indicates an expected attribute is missing.
	ErrorTagMissingAttribute
	// ErrorTagMissingElement indicates an expected element is missing. ErrorInfo's BadElement field will contain the name of the missing element.
	ErrorTagMissingElement
	// ErrorTagOpFailed indicates the request could not be completed because the operation failed for some reason not covered by any other error condition.
	ErrorTagOpFailed
	// ErrorTagOpNotSupported indicates the request could not be completed because it is not supported by the implementation.
	ErrorTagOpNotSupported
	// ErrorTagOpPartial indicates some part of the requested operation failed or was not attempted. Full cleanup has not been performed by the server. ErrorInfo identifies which portions succeeded, failed, and were not attempted.
	ErrorTagOpPartial
	// ErrorTagResourceDenied indicates the request could not be completed because of insufficient resources.
	ErrorTagResourceDenied
	// ErrorTagRollbackFailed indicates the request to roll back some configuration change was not completed.
	ErrorTagRollbackFailed
	// ErrorTagTooBig indicates the request or response (that would be generated) is too large for the implementation to handle.
	ErrorTagTooBig
	// ErrorTagUnknown probably indicates an internal error, because the error type could not be identified.
	ErrorTagUnknown
	// ErrorTagUnknownAttribute indicates an unexpected attribute is present. The ErrorInfo's BadAttribute and BadElement fields will contain more detail.
	ErrorTagUnknownAttribute
	// ErrorTagUnknownElement indicates an unexpected element is present. ErrorInfo's BadElement field will contain the unexpected element's name.
	ErrorTagUnknownElement
	// ErrorTagUnknownNamespace indicates an unexpected namespace is present. ErrorInfo's BadElement and BadNamespace fields will contain more detail.
	ErrorTagUnknownNamespace
)

var errorTagStringArray = [...]string{
	ErrorTagZero:             "",
	ErrorTagAccessDenied:     "access-denied",
	ErrorTagBadAttribute:     "bad-attribute",
	ErrorTagBadElement:       "bad-element",
	ErrorTagDataExists:       "data-exists",
	ErrorTagDataMissing:      "data-missing",
	ErrorTagInUse:            "in-use",
	ErrorTagInvalidValue:     "invalid-value",
	ErrorTagLockDenied:       "lock-denied",
	ErrorTagMalformedMessage: "malformed-message",
	ErrorTagMissingAttribute: "missing-attribute",
	ErrorTagMissingElement:   "missing-element",
	ErrorTagOpFailed:         "operation-failed",
	ErrorTagOpNotSupported:   "operation-not-supported",
	ErrorTagOpPartial:        "partial-operation",
	ErrorTagResourceDenied:   "resource-denied",
	ErrorTagRollbackFailed:   "rollback-failed",
	ErrorTagTooBig:           "too-big",
	ErrorTagUnknown:          "unknown",
	ErrorTagUnknownAttribute: "unknown-attribute",
	ErrorTagUnknownElement:   "unknown-element",
	ErrorTagUnknownNamespace: "unknown-namespace",
}

// Tag returns the XML tag that stores this value.
func (et ErrorTag) Tag() string {
	if int(et) < len(errorTagStringArray) {
		return errorTagStringArray[et]
	}
	return errorTagStringArray[ErrorTagUnknown]
}

// Severity returns the severity of this ErrorTag.
func (et ErrorTag) Severity() ErrorSeverity {
	switch et {
	case ErrorTagZero:
		return ErrorSeverityZero
	case ErrorTagInUse,
		ErrorTagInvalidValue,
		ErrorTagTooBig,
		ErrorTagMissingAttribute,
		ErrorTagBadAttribute,
		ErrorTagUnknownAttribute,
		ErrorTagMissingElement,
		ErrorTagBadElement,
		ErrorTagUnknownElement,
		ErrorTagUnknownNamespace,
		ErrorTagAccessDenied,
		ErrorTagLockDenied,
		ErrorTagResourceDenied,
		ErrorTagRollbackFailed,
		ErrorTagDataExists,
		ErrorTagDataMissing,
		ErrorTagOpNotSupported,
		ErrorTagOpFailed,
		ErrorTagOpPartial,
		ErrorTagMalformedMessage:
		return ErrorSeverityError
	default:
		return ErrorSeverityUnknown
	}
}

// UnmarshalText satisfies the encoding.TextUnmarshaler interface.
func (et *ErrorTag) UnmarshalText(text []byte) error {
	tagSlice := errorTagStringArray[:]
	if i := sort.SearchStrings(tagSlice, string(bytes.TrimSpace(text))); i != len(tagSlice) {
		*et = ErrorTag(uint64(i))
		return nil
	}
	*et = ErrorTagUnknown
	return fmt.Errorf("netconf: unknown ErrorTag %q", text)
}

// String satisfies the Stringer interface.
func (et *ErrorTag) String() string {
	return et.Tag()
}

// Error encapsulates a NETCONF RPC error.
type Error struct {
	Type     ErrorType     `xml:"error-type"`
	Tag      ErrorTag      `xml:"error-tag"`
	Severity ErrorSeverity `xml:"error-severity"`
	Info     ErrorInfo     `xml:"error-info"`
	Path     string        `xml:"error-path"`
	Message  string        `xml:"error-message"`
}

// Error is the implementation of the error interface.
func (e *Error) Error() string {
	return e.Message
}
