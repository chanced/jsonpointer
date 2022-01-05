package jsonpointer_test

import (
	"testing"

	"github.com/chanced/jsonpointer"
	"github.com/stretchr/testify/require"
)

func TestAssign(t *testing.T) {
	assert := require.New(t)

	r := Root{}

	tests := []struct {
		ptr   jsonpointer.JSONPointer
		value interface{}
		err   error
		run   func(v interface{})
	}{
		{"/nested/str", "strval", nil, func(val interface{}) {
			assert.Equal(val, r.Nested.String)
		}},
		{"/nestedptr/str", "x", nil, func(val interface{}) {
			assert.Equal(val, r.NestedPtr.String)
		}},
		{"/nested/entrymap/keyval/name", "entry-name", nil, func(v interface{}) {
			assert.Contains(r.Nested.EntryMap, "keyval")
			assert.Equal("entry-name", r.Nested.EntryMap["keyval"].Name)
		}},
	}

	for i, test := range tests {
		err := jsonpointer.Assign(test.ptr, &r, test.value)
		if test.err != nil {
			assert.ErrorIs(err, test.err, "test %d, pointer %s", i, test.ptr)
		} else {
			assert.NoError(err)
			test.run(test.value)
		}
	}
}
