// Copyright 2022 Chance Dinkins
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
//
// The License can be found in the LICENSE file.
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	//
	ErrMalformedToken = errors.New(`jsonpointer: fragment is malformed`)

	// ErrMalformedStart is an ErrMalformedToken that is returned when the JSON
	// Pointer is not empty or does not start with a "/".
	ErrMalformedStart = fmt.Errorf(`%w; pointer must be an empty string or start with "/"`, ErrMalformedToken)

	ErrMalformedEncoding = fmt.Errorf("%w; '~' must be encoded as ~0", ErrMalformedToken)

	// ErrNonPointer indicates a non-pointer value was passed to Assign.
	//
	ErrNonPointer = errors.New("jsonpointer: dst must be a pointer")

	// ErrUnexportedField indicates the given path is not reachable due to being
	// an unexported field.
	//
	ErrUnexportedField = errors.New("jsonpointer: unexported field")

	// ErrInvalidKeyType indicates the key type is not supported.
	//
	// Custom key types must implement encoding.TextUnmarshaler
	//
	ErrInvalidKeyType = errors.New("jsonpointer: invalid key type")

	// ErrNotAssignable indicates the type of the value is not assignable to the
	// provided path.
	//
	ErrNotAssignable = errors.New("jsonpointer: invalid value type")

	// ErrNotFound indicates a JSONPointer is not reachable from the root object
	// (e.g. a nil pointer, missing map key).
	//
	ErrNotFound = errors.New(`jsonpointer: value not found`)

	// ErrOutOfRange indicates an index is out of range for an array or slice
	//
	ErrOutOfRange = errors.New("jsonpointer: index out of range")

	// ErrUnreachable indicates a reference is not reachable. This occurs when
	// resolving and a primitive (string, number, or bool) leaf node is reached
	// and the reference is not empty.
	//
	ErrUnreachable = fmt.Errorf("%w due to being unreachable", ErrNotFound)

	// ErrNilInterface is returned when assigning and a nil interface is
	// reached.
	//
	// To solve this, the node containing the interface should implement
	// jsonpoint.Resolver and return a non-nil implemention of the interface.
	//
	ErrNilInterface = errors.New("jsonpointer: can not assign due to nil interface")

	// ErrMalformedIndex indicates a syntax error in the index or a slice or an array.
	ErrMalformedIndex = errors.New("jsonpointer: malformed slice or array index")
)

// Error is a base error type returned from Resolve, Assign, and Delete.
type Error interface {
	error
	JSONPointer() Pointer
	CurrentJSONPointer() Pointer
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

func newError(err error, s state, typ reflect.Type) *ptrError {
	return &ptrError{
		err:   err,
		typ:   typ,
		state: s,
	}
}

type ptrError struct {
	state
	err error
	typ reflect.Type
}

func (e *ptrError) Error() string {
	t, ok := e.Token()
	if ok {
		return fmt.Sprintf(`%v for token "%s" in reference "%v"`, e.err.Error(), t, e.ptr)
	}
	return fmt.Sprintf(`%v for reference "%v"`, e.err.Error(), e.ptr)
}

// Unwrap returns the underlying error.
func (e *ptrError) Unwrap() error {
	return e.err
}

func (e *ptrError) updateState(s state) {
	e.state = s
}

// JSONPointer returns the initial JSONPointer.
func (e *ptrError) JSONPointer() Pointer {
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

// AsKeyError returns err as a ValueError, if possible. It does so by calling
// errors.As, returning a KeyError and true if successful. If unsuccessful, nil
// and false is returned.
func AsKeyError(err error) (KeyError, bool) {
	var e KeyError
	return e, errors.As(err, &e)
}

func newKeyError(err error, s state, typ reflect.Type, keyValue interface{}, keyType reflect.Type) *keyError {
	if e, ok := err.(*keyError); ok {
		return e
	}
	return &keyError{
		ptrError: ptrError{
			state: s,
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

// AsValueError returns err as a ValueError, if possible.
func AsValueError(err error) (ValueError, bool) {
	var e ValueError
	return e, errors.As(err, &e)
}

func newValueError(err error, s state, typ reflect.Type, valType reflect.Type) *valueError {
	return &valueError{
		valuetype: valType,
		ptrError: ptrError{
			state: s,
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
	return fmt.Sprintf("%v (%v) for reference \"%v\"; expected %v", e.ptrError.err, e.valuetype, e.ptr, e.typ)
}

func (e *valueError) ValueType() reflect.Type {
	return e.valuetype
}

func updateErrorState(err error, s state) {
	if e, ok := err.(interface{ updateState(s state) }); ok {
		e.updateState(s)
	}
}

// IndexError indicates an error occurred with regards to an index of a slice or
// array. The error may be wrapped in an Error if it is returned from an operation on a
// JSON Pointer.
//
// err.Index() will return -1 if:
//
// - the source or destination is an array, token is equal to "-",
// and the array does not have a zero value.
//
// - the token can not be parsed as an int
//
type IndexError interface {
	MaxIndex() int
	Index() int
	Error() string
	Unwrap() error
}

type indexError struct {
	err      error
	maxIndex int
	index    int
}

func (e *indexError) MaxIndex() int {
	return e.maxIndex
}

func (e *indexError) Index() int {
	return e.index
}

func (e *indexError) Error() string {
	if errors.Is(e.err, ErrOutOfRange) {
		return fmt.Sprintf("%v; expected index to be equal to or less than next (%d) but is (%d)", ErrOutOfRange, e.maxIndex, e.index)
	}
	return fmt.Sprintf("%v for index %d of %d", e.err.Error(), e.index, e.maxIndex)
}

// AsIndexError returns err as a IndexError, if possible. It does so by calling
// errors.As, returning a IndexError and true if successful. If unsuccessful, nil
// and false is returned.
func AsIndexError(err error) (IndexError, bool) {
	var ie IndexError
	return ie, errors.As(err, &ie)
}

func (e *indexError) Unwrap() error {
	return e.err
}

type nilInterfaceError struct {
	ptrError
}

func newNilInterfaceError(s state, typ reflect.Type) *nilInterfaceError {
	return &nilInterfaceError{
		ptrError: ptrError{
			state: s,
			err:   ErrNilInterface,
			typ:   typ,
		},
	}
}

func (e *nilInterfaceError) Error() string {
	t, _ := e.Token()
	return fmt.Sprintf("jsonpointer: can not assign token \"%s\" of \"%s\" because %v is nil and can not be instantiated.", t, e.ptr, e.typ)
}
