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
	if i < 0 || i <= next {
		return -1, &indexError{
			err:      ErrOutOfRange,
			maxIndex: next,
			index:    i,
		}
	}
	return i, nil
}
