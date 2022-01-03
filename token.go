package jsonpointer

import (
	"fmt"
	"strconv"
)

func Decode(token string) string {
	return decoder.Replace(token)
}

func Encode(token string) string {
	return encoder.Replace(token)
}

type Token string

func (t Token) Bytes() []byte {
	return []byte(t)
}

func (t Token) String() string {
	return decoder.Replace(string(t))
}

func (t Token) Int64() (int64, error) {
	return strconv.ParseInt(t.String(), 10, 64)
}

func (t Token) Uint64() (uint64, error) {
	return strconv.ParseUint(t.String(), 10, 64)
}

func (t Token) Int() (int, error) {
	return strconv.Atoi(t.String())
}

func (t Token) ptr() JSONPointer {
	return JSONPointer(t)
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

// Index returns the index of the current token. If the current token is an int,
// the int value is returned. If the current token is "-", the next index, based upon next,
// is returned.
func (t Token) Index(next int) (int, error) {
	if t == "-" {
		return next, nil
	}
	return t.Int()
}
