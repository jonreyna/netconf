package netconf

import (
	"bytes"
	"fmt"
	"sort"
)

// TODO: Add a flag to return errors for warnings when constructing a Decoder.
// TODO: Remove *Unknown constants, and return errors with zero values.
// TODO: Use uint instead of uint64
// TODO: Make a better Error() implementation using more of the ReplyError data.

// UnmarshalTextError is returned when UnmarshalText fails to parse
// the text it's given.
type UnmarshalTextError struct {
	Type  string // Type is the unknown value's type.
	Value string // Value is the what caused the failure.
}

// Error is UnmarshalTextError's implementation of the error interface.
func (ute *UnmarshalTextError) Error() string {
	return fmt.Sprintf("UnmarshalText: unknown %s parsing %q", ute.Type, ute.Value)
}

// ErrorSeverity identifies severity of the error as either warning, or error.
type ErrorSeverity uint

const (
	ErrorSeverityZero    ErrorSeverity = iota // ErrorSeverityZero represents an uninitialized ErrorSeverity value.
	ErrorSeverityError                        // ErrorSeverityError indicates the severity is on the error level.
	ErrorSeverityUnknown                      // ErrorSeverityUnknown means the ErrorSeverity could not be identified, and may indicate an internal error.
	ErrorSeverityWarning                      // ErrorSeverityWarning is not yet utilized, according to RFC 6241.
)

// errorSeverityStringArray contains all error severity
// levels, and is used to translate ErrorSeverities to
// and from strings.
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

// UnmarshalText sets the receiver to the constant represented
// by the text argument given. If the text argument does not
// represent a known ErrorSeverity, it is set to the
// ErrorSeverityUnknown constant, and an UnmarshalTextError
// is returned.
func (es *ErrorSeverity) UnmarshalText(text []byte) error {

	sText := string(bytes.ToLower(bytes.TrimSpace(text)))
	if i := sort.SearchStrings(errorSeverityStringArray[:], sText); i != len(errorSeverityStringArray) && errorSeverityStringArray[i] == sText {
		*es = ErrorSeverity(i)
		return nil
	}

	*es = ErrorSeverityUnknown
	return &UnmarshalTextError{Type: "ErrorSeverity", Value: string(text)}
}

// ErrorInfo contains protocol or data model specific error content.
type ErrorInfo struct {
	BadAttribute string   `xml:"bad-attribute"` // BadAttribute has the name(s) of the bad, missing, or unexpected attribute(s).
	BadElement   string   `xml:"bad-element"`   // BadElement is the name of the element that should (or does) contain the missing (or bad) attribute.
	BadNamespace string   `xml:"bad-namespace"` // BadNamespace contains the name of the unexpected namespace.
	OkElement    []string `xml:"ok-element"`    // OkElement is the parent element for which all children have completed the requested operation.
	ErrElement   []string `xml:"err-element"`   // ErrElement is the parent element for which all children have failed to complete the requested operation.
	NOPElement   []string `xml:"noop-element"`  // NOPElement is the parent element that identifies all children for which the requested operation was not attempted.
}

// ErrorType defines the conceptual layer that the error occurred in.
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

// errorTypeStringArray contains all error types,
// and is used to translate ErrorTypes to and from
// strings.
var errorTypeStringArray = [...]string{
	ErrorTypeZero:        "",
	ErrorTypeApplication: "application",
	ErrorTypeProtocol:    "protocol",
	ErrorTypeRPC:         "rpc",
	ErrorTypeTransport:   "transport",
	ErrorTypeUnknown:     "unknown",
}

// String returns a string representation of the
// ErrorType. If the ErrorType is unknown, the
// ErrorTypeUnknown constant is returned.
func (es ErrorType) String() string {
	if int(es) < len(errorTypeStringArray) {
		return errorTypeStringArray[es]
	}
	return errorTypeStringArray[ErrorTypeUnknown]
}

// UnmarshalText sets the ErrorType receiver to the constant
// represented by the text argument given. If the text argument
// does not represent a known ErrorType, it is set
// to the ErrorTypeUnknown constant, and an UnmarshalTextError
// is returned.
func (es *ErrorType) UnmarshalText(text []byte) error {

	sText := string(bytes.ToLower(bytes.TrimSpace(text)))
	if i := sort.SearchStrings(errorTypeStringArray[:], sText); i != len(errorTypeStringArray) && errorTypeStringArray[i] == sText {
		*es = ErrorType(i)
		return nil
	}

	*es = ErrorTypeUnknown
	return &UnmarshalTextError{Type: "ErrorType", Value: string(text)}
}

// ErrorTag identifies the error condition.
type ErrorTag uint64

const (
	ErrorTagZero             ErrorTag = iota // ErrorTagZero is an uninitialized ErrorTag value.
	ErrorTagAccessDenied                     // ErrorTagAccessDenied indicates access was denied because authorization failed.
	ErrorTagBadAttribute                     // ErrorTagBadAttribute indicates an attribute value is not correct; e.g., wrong type, out of range, pattern mismatch.
	ErrorTagBadElement                       // ErrorTagBadElement indicates a bad element value was in the RPC. ErrorInfo's BadElement field will contain element's name.
	ErrorTagDataExists                       // ErrorTagDataExists indicates data could not be created because it already exists.
	ErrorTagDataMissing                      // ErrorTagDataMissing indicates data could not be deleted because it doesn't exist.
	ErrorTagInUse                            // ErrorTagInUse indicates the required resource is already in use.
	ErrorTagInvalidValue                     // ErrorTagInvalidValue indicates the RPC specifies an unacceptable value for one or more parameters.
	ErrorTagLockDenied                       // ErrorTagLockDenied indicates the requested lock is denied because it is held by another entity.
	ErrorTagMalformedMessage                 // ErrorTagMalformedMessage indicates a failure to parse the RPC correctly.
	ErrorTagMissingAttribute                 // ErrorTagMissingAttribute indicates an expected attribute is missing.
	ErrorTagMissingElement                   // ErrorTagMissingElement indicates an expected element is missing. ErrorInfo's BadElement field will contain the name of the missing element.
	ErrorTagOpFailed                         // ErrorTagOpFailed indicates failure for some reason not covered by other error conditions.
	ErrorTagOpNotSupported                   // ErrorTagOpNotSupported indicates RPC is not supported by the implementation.
	ErrorTagOpPartial                        // ErrorTagOpPartial indicates the RPC failed partially or was aborted, and cleanup was not performed. ErrorInfo's OkElement, ErrElement, and NOPElement identify elements that succeeded, failed, and were aborted respectively.
	ErrorTagResourceDenied                   // ErrorTagResourceDenied indicates insufficient resources.
	ErrorTagRollbackFailed                   // ErrorTagRollbackFailed indicates the rollback was not completed.
	ErrorTagTooBig                           // ErrorTagTooBig indicates the request or response is too large to handle.
	ErrorTagUnknown                          // ErrorTagUnknown probably indicates an internal error, because the error type could not be identified.
	ErrorTagUnknownAttribute                 // ErrorTagUnknownAttribute indicates an unexpected attribute is present. ErrorInfo's BadAttribute and BadElement field will contain more detail.
	ErrorTagUnknownElement                   // ErrorTagUnknownElement indicates an unexpected element. ErrorInfo's BadElement field will contain its name.
	ErrorTagUnknownNamespace                 // ErrorTagUnknownNamespace indicates an unexpected namespace is present. ErrorInfo's BadElement and BadNamespace fields will contain more detail.
)

// errorTagStringArray contains all error tags,
// and is used to translate ErrorTag values to
// and from strings.
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

// String returns a string representation of this
// ErrorTag value.
func (et ErrorTag) String() string {
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

// UnmarshalText sets the ErrorTag receiver to the constant
// represented by the text argument given. If the text argument
// does not represent a known ErrorTag, the ErrorTag is set
// to the ErrorTagUnknown constant, and an UnmarshalTextError
// is returned.
func (et *ErrorTag) UnmarshalText(text []byte) error {

	sText := string(bytes.ToLower(bytes.TrimSpace(text)))
	if i := sort.SearchStrings(errorTagStringArray[:], sText); i != len(errorTagStringArray) &&
		errorTagStringArray[i] == sText {
		*et = ErrorTag(i)
		return nil
	}

	*et = ErrorTagUnknown
	return &UnmarshalTextError{Type: "ErrorTag", Value: string(text)}
}

// ReplyError encapsulates a NETCONF RPC error, and implements the error interface.
type ReplyError struct {
	Type     ErrorType     `xml:"error-type"`     // Type is the conceptual layer that the error occurred.
	Tag      ErrorTag      `xml:"error-tag"`      // Tag identifies the error condition.
	Severity ErrorSeverity `xml:"error-severity"` // Severity is the error severity: either error or warning.
	Info     ErrorInfo     `xml:"error-info"`     // Info contains protocol or data-model-specific error content.
	Path     string        `xml:"error-path"`     // Path is the absolute XPath expression identifying the element path to the node.
	Message  string        `xml:"error-message"`  // Message is a human friendly description of the error.
}

// Error is the implementation of the error interface.
func (e *ReplyError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("%s %s %s", e.Severity, e.Tag, e.Info.BadElement)
}
