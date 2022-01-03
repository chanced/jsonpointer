package jsonpointer_test

import (
	"strings"
	"testing"

	"github.com/chanced/jsonpointer"
	"github.com/stretchr/testify/require"
)

// func TestError(t *testing.T) {
// 	assert := require.New(t)
// 	var err error = &jsonpointer.FormatError{}
// 	err2 := &jsonpointer.FormatError{}
// 	assert.ErrorIs(err, err2)
// }
type strval string

func (str strval) String() string { return string(str) }

func TestNew(t *testing.T) {
	assert := require.New(t)
	tests := []struct {
		pointer        jsonpointer.JSONPointer
		expectedstring string
	}{
		{jsonpointer.New(""), "/"},
		{jsonpointer.New("foo", "bar", "baz"), "/foo/bar/baz"},
		{jsonpointer.New("/foo/bar/baz"), "/~1foo~1bar~1baz"},
		{jsonpointer.New("/"), "/~1"},
		{jsonpointer.New(), ""},
		{jsonpointer.New(strings.Split("#/foo/bar/baz", "/")...), "/foo/bar/baz"},
		{jsonpointer.New("~"), "/~0"},
		{jsonpointer.New("\\#"), "/#"},
	}
	for _, e := range tests {
		assert.Equal(e.expectedstring, e.pointer.String())
	}
}

func TestNewFromStrings(t *testing.T) {
	assert := require.New(t)
	p := jsonpointer.NewFromStrings([]string{"foo", "bar", "baz"})
	assert.Equal("/foo/bar/baz", p.String())
}

func TestJSONPointerNext(t *testing.T) {
	assert := require.New(t)
	tests := []struct {
		pointer         jsonpointer.JSONPointer
		expectedpointer jsonpointer.JSONPointer
		expectedtoken   jsonpointer.Token
		expectedok      bool
	}{
		{"", "", "", false},
		{"/", "", "", true},
		{"/foo", "", "foo", true},
		{"/foo/bar", "/bar", "foo", true},
		{"/~1foo~1bar~1baz", "", "~1foo~1bar~1baz", true},
		{"malformed", "", "", false},
		{"/foo/bar/baz", "/bar/baz", "foo", true},
	}
	for i, e := range tests {
		np, nt, ok := e.pointer.Next()
		assert.Equal(e.expectedpointer, np, "test[%v]: expected pointer to equal %v, got %v", i, e.expectedpointer, np)
		assert.Equal(e.expectedtoken, nt, "test[%v]: expected token to equal %v, got %v", i, e.expectedtoken, nt)
		assert.Equal(e.expectedok, ok, "test[%v]: expected ok to equal %v, got %v", i, e.expectedok, ok)
	}
}

func TestJSONPointerPop(t *testing.T) {
	assert := require.New(t)
	pt := []struct {
		pointer         jsonpointer.JSONPointer
		expectedpointer jsonpointer.JSONPointer
		expectedtoken   jsonpointer.Token
		expectedok      bool
	}{
		{"/foo/bar/baz", "/foo/bar", "baz", true},
		{"/foo/bar", "/foo", "bar", true},
		{"/", "", "", true},
		{"", "", "", false},
		{"malformed", "", "", false},
	}
	for _, e := range pt {
		np, tok, ok := e.pointer.Pop()
		assert.Equal(e.expectedtoken, tok, "expected token to equal %v, got %v", e.expectedtoken, tok)
		assert.Equal(e.expectedpointer, np, "expected pointer to equal %v, got %v", e.expectedpointer, np)
		assert.Equal(e.expectedok, ok, "expected ok to equal %v, got %v", e.expectedok, ok)
	}
}
