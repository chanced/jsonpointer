package jsonpointer_test

import (
	"fmt"
	"testing"

	"github.com/chanced/jsonpointer"
	"github.com/stretchr/testify/require"
)

func TestAssign(t *testing.T) {
	assert := require.New(t)

	var r *Root

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
		{"/nested/strarray/1", "strval", nil, func(v interface{}) {
			assert.Equal(v, r.Nested.StrArray[1])
		}},
		{"/nested/intarray/1", int(1), nil, func(v interface{}) {
			assert.Equal(v, r.Nested.IntArray[1])
		}},
		{"/nested/anon/value", "val", nil, func(v interface{}) {
			assert.Equal(v, r.Nested.AnonStruct.Value)
		}},
		{"/nested/strslice/-", "val", nil, func(v interface{}) {
			assert.Len(r.Nested.StrSlice, 1)
			assert.Equal(v, r.Nested.StrSlice[0])
		}},
		{"/nested/strslice/-", "val2", nil, func(v interface{}) {
			assert.Len(r.Nested.StrSlice, 2)
			assert.Equal(v, r.Nested.StrSlice[1])
		}},
		{"/nested/custommap/key", "val", nil, func(v interface{}) {
			assert.Len(r.Nested.CustomMap, 1)
			assert.Contains(r.Nested.CustomMap, Key{"key"})
		}},
		{"/nested/embedded/value", "embed-val", nil, func(v interface{}) {
			assert.Equal(v, r.Nested.Embedded.Value)
		}},
		{"/nested/yield/value", "yielded value", nil, func(v interface{}) {
			assert.Equal(v, r.Nested.Yield.Value)
		}},
		{"/nested/interface/private/value", uint(3), nil, func(v interface{}) {
			assert.Equal(v, r.Nested.InterContainer.Interface.Value())
		}},
	}

	for i, test := range tests {
		fmt.Printf("=== RUN TestAssign#%d, pointer %s\n", i+1, test.ptr)
		err := jsonpointer.Assign(&r, test.ptr, test.value)
		if test.err != nil {
			assert.ErrorIs(err, test.err)
		} else {
			assert.NoError(err)
			test.run(test.value)
		}
		fmt.Printf("--- PASS TestAssign #%d, pointer %s\n", i, test.ptr)
	}
}
