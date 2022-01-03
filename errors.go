package jsonpointer

import (
	"errors"
	"fmt"
	"reflect"
)

// TODO: need to clean these up and provide better error message\

var (
	// ErrMalformedToken is returned when a JSONPointer token is malformed.
	//
	// This error is returned by JSONPointer.Validate() which is called by
	// Resolve, Assign, and Delete.
	ErrMalformedToken = fmt.Errorf(`jsonpointer: reference must be empty or start with a "/"`)
	// ErrNonPointer indicates a non-pointer value was passed to Assign.
	ErrNonPointer = errors.New("jsonpointer: dst must be a pointer")
	// ErrUnexportedField indicates the given path is not reachable due to being
	// an unexported field.
	ErrUnexportedField = errors.New("jsonpointer: unexported field")
	// ErrInvalidKeyType indicates the key type is not supported.
	//
	// Custom key types must implement encoding.TextUnmarshaler
	ErrInvalidKeyType = errors.New("jsonpointer: invalid key type")
	// ErrNotAssignable indicates the type of the value is not assignable to the
	// provided path.
	ErrNotAssignable = errors.New("jsonpointer: invalid value type")
	// ErrNotFound indicates a JSONPointer is not reachable from the root object
	// (e.g. a nil pointer, missing map key).
	ErrNotFound = errors.New(`jsonpointer: token path not found`)
	// ErrOutOfBounds indicates an index is out of bounds for an array or slice
	ErrOutOfBounds = errors.New("jsonpointer: index out of bounds")
	// ErrInvalidReference indicates a reference is not reachable. This occurs
	// when a primitive leaf node is reached and the reference is not empty.
	ErrInvalidReference = errors.New("jsonpointer: bad reference")
)

// Error is a base error type returned from Resolve, Assign, and Delete.
type Error interface {
	error
	JSONPointer() JSONPointer
	CurrentJSONPointer() JSONPointer
	Token() (Token, bool)
	Operation() Operation
	Unwrap() error
	Type() reflect.Type
}

func isError(err error) bool {
	_, ok := err.(Error)
	return ok
}

func AsError(err error) (Error, bool) {
	var e Error
	return e, errors.As(err, &e)
}

func newError(err error, s *state, typ reflect.Type) *ptrError {
	return &ptrError{
		err:   err,
		typ:   typ,
		state: *s,
	}
}

type ptrError struct {
	state
	err error
	typ reflect.Type
}

func (e *ptrError) Error() string {
	return e.err.Error()
}

// Unwrap returns the underlying error.
func (e *ptrError) Unwrap() error {
	return e.err
}

// JSONPointer returns the initial JSONPointer.
func (e *ptrError) JSONPointer() JSONPointer {
	return e.ptr
}

// Type returns the reflect.Type of the current container object.
func (e *ptrError) Type() reflect.Type {
	return e.typ
}

// Token returns the token of the JSONPointer which encountered the error
func (e *ptrError) Token() (Token, bool) {
	return e.current.NextToken()
}

// KeyError indicates an error occurred with regards to the key of a map or
// slice.
type KeyError interface {
	Error
	KeyType() reflect.Type
	KeyValue() interface{}
}

func IsKeyError(err error) bool {
	if err == nil {
		return false
	}
	_, ok := AsKeyError(err)
	return ok
}

func AsKeyError(err error) (KeyError, bool) {
	var e KeyError
	return e, errors.As(err, &e)
}

func newKeyError(err error, s *state, typ reflect.Type, keyValue interface{}, keyType reflect.Type) *keyError {
	if e, ok := err.(*keyError); ok {
		return e
	}
	return &keyError{
		ptrError: ptrError{
			state: *s,
			err:   err,
			typ:   typ,
		},
		keyValue: keyValue,
		keyType:  keyType,
	}
}

type keyError struct {
	ptrError
	keyType  reflect.Type
	keyValue interface{}
}

func (e *keyError) Error() string {
	if e.typ == nil {
		return e.err.Error() + " (nil)"
	}
	return e.err.Error() + " for " + e.typ.String() + "(" + e.keyType.String() + ")"
}

func (e *keyError) KeyType() reflect.Type {
	return e.keyType
}

func (e *keyError) KeyValue() interface{} {
	return e.keyValue
}

// FieldError indicates an error occurred with regards to a field of a struct.
type FieldError interface {
	Error
	Field() reflect.StructField
}

type fieldError struct {
	ptrError
	field reflect.StructField
}

func (e *fieldError) Error() string {
	switch {
	case errors.Is(e.err, ErrUnexportedField):
		if t, ok := e.Token(); ok {
			return "jsonpointer: unexported field: " + t.String() + " " + e.typ.String() + "." + e.field.Name
		} else {
			return "jsonpointer: unexported field: " + e.typ.String() + "." + e.field.Name
		}
	default:
		return e.ptrError.Error()
	}
}

type ValueError interface {
	Error
	ValueType() reflect.Type
}

func IsValueError(err error) bool {
	if err == nil {
		return false
	}
	_, ok := AsValueError(err)
	return ok
}

func AsValueError(err error) (ValueError, bool) {
	var e ValueError
	return e, errors.As(err, &e)
}

func newValueError(err error, s *state, typ reflect.Type, valType reflect.Type) *valueError {
	return &valueError{
		valuetype: valType,
		ptrError: ptrError{
			state: *s,
			err:   err,
			typ:   typ,
		},
	}
}

type valueError struct {
	ptrError
	valuetype reflect.Type
}

func (e *valueError) Error() string {
	return e.ptrError.Error() + " " + e.valuetype.String() + " for " + e.typ.String()
}

func (e *valueError) ValueType() reflect.Type {
	return e.valuetype
}
