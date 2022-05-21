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
	"fmt"
	"strconv"
)

// Decode decodes a JSON Pointer token by replacing each encoded slash ("~1")
// with '/' (%x2F) and each encoded tilde ("~0") with '~' (%x7E).
func Decode(token string) string {
	return decoder.Replace(token)
}

// Encode encodes a string to a token of a JSON Pointer by replacing each '~'
// (%x7E) with "~0" and '/' (%x2F) with "~1".
func Encode(token string) string {
	return encoder.Replace(token)
}

// Token is a segment of a JSON Pointer, divided by '/' (%x2F).
type Token string

// Bytes returns the decoded Bytes of t
func (t Token) Bytes() []byte {
	return []byte(t.String())
}

// String returns the decoded value of t
func (t Token) String() string {
	return decoder.Replace(string(t))
}

// Int64 attempts to parse t as an int64. If t can be parsed as an int64 then
// the value is returned. If t can not be parsed as an int64 then an error is
// returned.
func (t Token) Int64() (int64, error) {
	return strconv.ParseInt(t.String(), 10, 64)
}

// Uint64 attempts to parse t as an uint64. If t can be parsed as an uint64 then
// the value is returned. If t can not be parsed as an uint64 then an error is
// returned.
func (t Token) Uint64() (uint64, error) {
	return strconv.ParseUint(t.String(), 10, 64)
}

// Int attempts to parse t as an int. If t can be parsed as an int then
// the value is returned. If t can not be parsed as an int then an error is
// returned.
func (t Token) Int() (int, error) {
	return strconv.Atoi(t.String())
}

func (t Token) ptr() Pointer {
	return Pointer(t)
}

// Index parses t for an index value. If t can be parsed as an int, is equal to
// or greater than 0 and less than or equal to next then the value is returned.
// If t is equal to "-" then next is returned. If neither condition is true, -1
// and an IndexError is returned.
//
// next must be greater than or equal to 0.
func (t Token) Index(next int) (int, error) {
	if next < 0 {
		return -1, fmt.Errorf("next (%d) must be greater than or equal to 0", next)
	}
	if t == "-" {
		return next, nil
	}
	i, err := t.Int()
	if err != nil {
		return -1, &indexError{
			err:      err,
			maxIndex: next,
			index:    i,
		}
	}
	if i < 0 || i > next {
		return -1, &indexError{
			err:      ErrOutOfRange,
			maxIndex: next,
			index:    i,
		}
	}
	return i, nil
}

// Tokens is a slice of Tokens.
type Tokens []Token

// Strings returns ts as a slice of strings
func (ts Tokens) Strings() []string {
	s := make([]string, len(ts))
	for i, t := range ts {
		s[i] = t.String()
	}
	return s
}

// Stringers returns ts as a slice of fmt.Stringers
func (ts Tokens) Stringers() []fmt.Stringer {
	s := make([]fmt.Stringer, len(ts))
	for i, t := range ts {
		s[i] = t
	}
	return s
}
