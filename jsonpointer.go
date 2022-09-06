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

// Package jsonpointer provides the ability to resolve, assign, and delete
// values of any type, including raw JSON, by [JSON Pointers](https://datatracker.ietf.org/doc/html/rfc6901).
package jsonpointer

import (
	"errors"
	"strings"
)

// YieldOperation returns resolution back to jsonpointer. This error can be
// utilized within methods satisfying Resolver (ResolveJSONPointer), Assigner
// (AssignByJSONPointer), and Deleter (DeleteByJSONPointer) as an escape hatch.
//
// The intent is is that there may only be certain fields that your application
// would like to manually resolve. For the rest, you'd return YieldOperation as
// the error.
var YieldOperation = errors.New("yield resolution to jsonpointer")

var (
	decoder = strings.NewReplacer("~1", "/", "~0", "~")
	encoder = strings.NewReplacer("/", "~1", "~", "~0")
)

const (
	// Root is a top-level JSONPointer, indicated by "/".
	Root Pointer = ""
)

// New encodes and returns the token + tokens into a JSONPointer.
//
// Examples:
//
//	jsonpointer.New("foo", "bar") => "/foo/bar"
//	jsonpointer.New("/foo/bar") => "/~1foo~1bar"
//	jsonpointer.New() => "/"
//	jsonpointer.New("") => "/"
//	jsonpointer.New("/") => "/~1"
//	jsonpointer.New("~") => "/~0"
func New(tokens ...string) Pointer {
	return NewFromStrings(tokens)
}

// From accepts a string, trims any leading '#' and returns str as a Pointer as
// well as any validation errors.
func From(ptr string) (Pointer, error) {
	if len(ptr) == 0 {
		return Root, nil
	}
	if ptr[0] == '#' {
		ptr = ptr[1:]
	}
	j := Pointer(ptr)
	return j, j.Validate()
}

// NewFromStrings encodes and returns the tokens into a JSONPointer.
//
// Examples:
//
//	jsonpointer.NewFromStrings([]string{"foo", "bar", "baz"}) => "/foo/bar/baz"
func NewFromStrings(tokens []string) Pointer {
	b := &strings.Builder{}
	b.Grow(len(tokens))
	if len(tokens) == 0 {
		return ""
	}
	for _, token := range tokens {
		b.WriteRune('/')
		if _, err := encoder.WriteString(b, token); err != nil {
			// this should never happen
			panic(err)
		}
	}
	return Pointer(b.String())
}

// Alias for Pointer
// Deprecated in favor of Pointer
type JSONPointer = Pointer

// A Pointer is a Unicode string containing a sequence of zero or more
// reference tokens, each prefixed by a '/' character.
//
// See [rfc 6901 for more information](https://datatracker.ietf.org/doc/html/rfc6901).
type Pointer string

func (p Pointer) String() string {
	return string(p)
}

// Append appends token to the end of reference p and returns the new JSONPointer.
//
// Note: token is not encoded. Use p.AppendString to encode and append the
// token.
func (p Pointer) Append(token Token) Pointer {
	return p + "/" + token.ptr()
}

// AppendString encodes and appends token to the value of p and returns the new
// JSONPointer.
func (p Pointer) AppendString(token string) Pointer {
	return p.Append(Token(encoder.Replace(token)))
}

// Preppend prepends token to the beginning of the value of p and returns the
// resulting JSONPointer.
//
// Note: token is not encoded. Use p.PrependString to encode and prepend the
// token.
func (p Pointer) Prepend(token Token) Pointer {
	return "/" + token.ptr() + p
}

// PrependString encodes and prepends token to the value of p and returns the new
// JSONPointer.
func (p Pointer) PrependString(token string) Pointer {
	return p.Prepend(Token(encoder.Replace(token)))
}

// Validate performs validation on p. The following checks are performed:
//
// - p must be either empty or start with '/
//
// - p must be properly encoded, meaning that '~' must be immediately followed
// by a '0' or '1'.
func (p Pointer) Validate() (err error) {
	if err = p.validateStart(); err != nil {
		return err
	}
	return p.validateeEncoding()
}

func (p Pointer) validateeEncoding() error {
	if len(p) == 0 {
		return nil
	}
	if p[len(p)-1] == '~' {
		return ErrMalformedEncoding
	}
	for i := len(p) - 1; i >= 0; i-- {
		if p[i] == '~' && (p[i+1] != '0' && p[i+1] != '1') {
			return ErrMalformedEncoding
		}
	}
	return nil
}

func (p Pointer) validateStart() error {
	if len(p) > 0 && !startsWithSlash(p) {
		return ErrMalformedStart
	}
	return nil
}

func (p Pointer) Pop() (Pointer, Token, bool) {
	i := lastSlash(p)
	if i == -1 {
		return "", "", false
	}
	return Pointer(p[:i]), Token(p[i+1:]), true
}

func (p Pointer) LastToken() (Token, bool) {
	_, t, e := p.Pop()
	return t, e
}

// Next splits the JSONPointer at the first slash and returns the token and the
// remaining JSONPointer.
func (p Pointer) Next() (Pointer, Token, bool) {
	if p == "" {
		return "", "", false
	}
	i := nextSlash(p)
	switch i {
	case -1:
		return "", "", false
	case 0:
		return "", Token(p[1:]), true
	default:
		return Pointer(p[i:]), Token(p[1:i]), true
	}
}

// NextToken splits the JSONPointer at the first slash and returns the token.
func (p Pointer) NextToken() (Token, bool) {
	_, t, ok := p.Next()
	return t, ok
}

func (p Pointer) NextPointer() (Pointer, bool) {
	v, _, ok := p.Next()
	return v, ok
}

func (p Pointer) IsRoot() bool {
	return p == Root
}

// Tokens returns the decoded tokens of the JSONPointer.
func (p Pointer) Tokens() []string {
	if p == "" {
		return []string{}
	}
	tokens := strings.Split(string(p), "/")
	for i, token := range tokens {
		tokens[i] = Decode(token)
	}
	return tokens
}

// func (p *JSONPointer) Resolve(value interface{}, target interface{}) error {
// }

// lastSlash(ptr) returns the index of the last slash in the JSONPointer or -1
// if not present
func lastSlash(ptr Pointer) int {
	return strings.LastIndexByte(string(ptr), '/')
}

func startsWithSlash(ptr Pointer) bool {
	if len(ptr) == 0 {
		return false
	}
	return ptr[0] == '/'
}

// nextSlash(ptr) returns the index of the next slash in the JSONPointer.
func nextSlash(ptr Pointer) int {
	if !startsWithSlash(ptr) {
		return -1
	}
	i := strings.IndexByte(string(ptr)[1:], '/')
	if i == -1 {
		return 0
	}
	return i + 1
}
