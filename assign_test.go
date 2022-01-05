package jsonpointer_test

import (
	"fmt"
	"testing"

	"github.com/chanced/jsonpointer"
	"github.com/stretchr/testify/require"
)

func TestAssignStructField(t *testing.T) {
	assert := require.New(t)

	r := Root{
		Nested: Nested{
			String: "",
		},
	}

	tests := []struct {
		ptr   jsonpointer.JSONPointer
		value interface{}
		err   error
		run   func(v interface{})
	}{
		{"/nested/str", "strval", nil, func(val interface{}) {
			assert.Equal(val, r.Nested.String)
			fmt.Println(r.Nested.String)
		}},
		{"/nestedptr/str", "x", nil, func(val interface{}) {
			assert.Equal(val, r.NestedPtr.String)
			fmt.Println(r.NestedPtr.String)
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

func TestAssignMapValue(t *testing.T) {
	// assert := require.New(t)
	// m := make(map[string]structentry)
	// ptr := JSONPointer("/foo")
	// assign := newAssignState(ptr)
	// val := structentry{Name: "fooval"}
	// err := assign.setMapIndex(
	// 	ptr,
	// 	reflect.TypeOf(m),
	// 	reflect.ValueOf(m),
	// 	reflect.ValueOf(val),
	// )
	// assert.NoError(err)
	// assert.Contains(m, "foo")
	// assert.Equal(val, m["foo"])
}
