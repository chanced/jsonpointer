package jsonpointer_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/chanced/jsonpointer"
	"github.com/sanity-io/litter"
	"github.com/stretchr/testify/require"
)

func TestAssign(t *testing.T) {
	assert := require.New(t)

	var r *Root

	tests := []struct {
		ptr   jsonpointer.JSONPointer
		value interface{}
		run   func(v interface{}, err error)
	}{
		{"/nested/str", "strval", func(val interface{}, err error) {
			assert.NoError(err)
			assert.Equal(val, r.Nested.Str)
		}},
		{"/nestedptr/str", "x", func(val interface{}, err error) {
			assert.NoError(err)
			assert.Equal(val, r.NestedPtr.Str)
		}},
		{"/nested/entrymap/keyval/name", "entry-name", func(v interface{}, err error) {
			assert.NoError(err)
			assert.Contains(r.Nested.EntryMap, "keyval")
			assert.Equal("entry-name", r.Nested.EntryMap["keyval"].Name)
		}},
		{"/nested/strarray/1", "strval", func(v interface{}, err error) {
			assert.NoError(err)
			assert.Equal(v, r.Nested.StrArray[1])
		}},
		{"/nested/intarray/1", 1, func(v interface{}, err error) {
			assert.NoError(err)
			assert.Equal(v, r.Nested.IntArray[1])
		}},
		{"/nested/intarray/2", 3, func(v interface{}, err error) {
			assert.NoError(err)
			assert.Equal(v, r.Nested.IntArray[2])
		}},
		{"/nested/intarray/3", 3, func(v interface{}, err error) {
			assert.Error(err)
			assert.ErrorIs(err, jsonpointer.ErrOutOfRange)
			ie, ok := jsonpointer.AsIndexError(err)
			assert.True(ok, "err is not an IndexError")
			assert.Equal(3, ie.Index())
		}},
		{"/nested/anon/value", "val", func(v interface{}, err error) {
			assert.NoError(err)
			assert.Equal(v, r.Nested.AnonStruct.Value)
		}},
		{"/nested/strslice/-", "val", func(v interface{}, err error) {
			assert.NoError(err)
			assert.Len(r.Nested.StrSlice, 1)
			assert.Equal(v, r.Nested.StrSlice[0])
		}},
		{"/nested/strslice/-", "val2", func(v interface{}, err error) {
			assert.NoError(err)
			assert.Len(r.Nested.StrSlice, 2)
			assert.Equal(v, r.Nested.StrSlice[1])
		}},
		{"/nested/custommap/key", "val", func(v interface{}, err error) {
			assert.NoError(err)
			assert.Len(r.Nested.CustomMap, 1)
			assert.Contains(r.Nested.CustomMap, Key{"key"})
		}},
		{"/nested/embedded/value", "embed-val", func(v interface{}, err error) {
			assert.NoError(err)
			assert.Equal(v, r.Nested.Embedded.Value)
		}},
		{"/nested/yield/value", "yielded value", func(v interface{}, err error) {
			assert.NoError(err)
			assert.Equal(v, r.Nested.Yield.Value)
		}},
		{"/nested/interface/private/value", uint(3), func(v interface{}, err error) {
			assert.NoError(err)
			assert.Equal(v, r.Nested.InterContainer.Interface.Value())
		}},
		{"/nested/json/obj/value", "val", func(v interface{}, err error) {
			assert.NoError(err)
			var jv JSONValue
			err = json.Unmarshal(r.Nested.JSON, &jv)
			assert.NoError(err)
			assert.Equal(v, jv.Obj.Value)
		}},
	}

	for i, test := range tests {
		fmt.Printf("=== RUN TestAssign #%d, pointer %s\n", i+1, test.ptr)
		err := jsonpointer.Assign(&r, test.ptr, test.value)
		test.run(test.value, err)

		fmt.Println("--- PASS")
	}
	// b, _ := json.MarshalIndent(r, "", "  ")
	// fmt.Println(string(b))
}

func TestAssignAny(t *testing.T) {
	assert := require.New(t)

	m := map[string]interface{}{}
	tests := []struct {
		ptr   jsonpointer.JSONPointer
		value interface{}
		err   error
		run   func(v interface{})
	}{
		{"/nested/str", "strval", nil, func(val interface{}) {
			assert.Contains(m, "nested")
			assert.Contains(m["nested"], "str")
			m := m["nested"].(map[string]interface{})
			assert.Equal(val, m["str"])
		}},
		{"/nestedptr/str", "x", nil, func(val interface{}) {
			assert.Contains(m, "nestedptr")
			assert.Contains(m["nestedptr"], "str")
			n := m["nestedptr"].(map[string]interface{})
			assert.Equal(n["str"], "x")
		}},
		{"/nested/array/0/entry/value", "entry value", nil, func(v interface{}) {
			a := m["nested"].(map[string]interface{})["array"].([]interface{})
			assert.Len(a, 1)
			mv := a[0].(map[string]interface{})
			assert.Contains(mv, "entry")
			e := mv["entry"].(map[string]interface{})
			assert.Contains(e, "value")
			assert.Equal(v, e["value"])
		}},
		{"/nested/intarray/0", int(1), nil, func(v interface{}) {
			a := m["nested"].(map[string]interface{})["intarray"].([]interface{})
			assert.Len(a, 1)
			assert.Equal(v, a[0])
		}},
		{"/nested/strslice/-", "val", nil, func(v interface{}) {
			a := m["nested"].(map[string]interface{})["strslice"].([]interface{})
			assert.Len(a, 1)
			assert.Equal(v, a[0])
		}},
		{"/nested/strslice/1", "val2", nil, func(v interface{}) {
			a := m["nested"].(map[string]interface{})["strslice"].([]interface{})
			assert.Len(a, 2)
			assert.Equal(v, a[1])
		}},
	}

	for i, test := range tests {
		fmt.Printf("=== RUN TestAssignAny #%d, pointer %s\n", i, test.ptr)
		err := jsonpointer.Assign(&m, test.ptr, test.value)
		if test.err != nil {
			assert.ErrorIs(err, test.err)
		} else {
			assert.NoError(err)
			test.run(test.value)
		}
		fmt.Println("--- PASS")
	}
	litter.Dump(m)
}

func TestAssignJSON(t *testing.T) {
	assert := require.New(t)
	_ = assert

	tests := []struct {
		ptr   jsonpointer.JSONPointer
		json  string
		value interface{}
		run   func(v Root, err error)
	}{
		{
			"/nested/str",
			`{
				"nested": {
					"str": "old-value"
				}
			}`,
			"new-value",
			func(r Root, err error) {
				assert.NoError(err)
				assert.Equal("new-value", r.Nested.Str)
			},
		},
	}

	for i, test := range tests {
		fmt.Printf("=== RUN TestAssignJSON #%d, pointer %s\n", i, test.ptr)
		b := []byte(test.json)
		err := jsonpointer.Assign(&b, test.ptr, test.value)
		var r Root
		if uerr := json.Unmarshal(b, &r); uerr != nil {
			assert.Failf("unmarshal failed: %v", uerr.Error())
		}
		test.run(r, err)
		fmt.Println("--- PASS")
	}
}
