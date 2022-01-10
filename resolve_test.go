package jsonpointer_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/chanced/jsonpointer"
	"github.com/stretchr/testify/require"
)

func TestResolveField(t *testing.T) {
	assert := require.New(t)
	floatv := float64(3.4)
	fp := &floatv
	bv := (true)
	bp := &bv
	anon := struct {
		Value string
	}{
		Value: "anon struct value",
	}
	r := Root{
		Nested: Nested{
			Str:           "strval",
			Float:         34.21,
			FloatPtr:      fp,
			Inline:        Inline{InlineStr: "inline value"},
			Nested:        &Nested{Str: "deeply nested value"},
			Embedded:      Embedded{Value: "embedded value"},
			IntSlice:      []int{},
			Bool:          true,
			BoolPtr:       bp,
			AnonStructPtr: &anon,
		},
	}

	tests := []struct {
		ptr         jsonpointer.JSONPointer
		expectedval interface{}
		expectederr error
	}{
		{"/nested/str", "strval", nil},
		{"/nested/float", 34.21, nil},
		{"/nested/floatptr", fp, nil},
		{"/nested/inline", "inline value", nil},
		{"/nested/nested/str", "deeply nested value", nil},
		{"/nested/bool", true, nil},
		{"/nested/boolptr", bp, nil},
		{"/nested/anonptr/value", "anon struct value", nil},
		{"/nested/private", nil, jsonpointer.ErrUnexportedField},
		{"/nested/invalid", nil, jsonpointer.ErrNotFound},
		{"/nested/empty/str", nil, jsonpointer.ErrUnreachable},
		{"/nested/empty/str", nil, jsonpointer.ErrNotFound},
		{"/nested/intslice/badkey", nil, jsonpointer.ErrMalformedIndex},
	}

	for i, test := range tests {
		fmt.Printf("=== RUN TestResolveField #%d, pointer %s\n", i, test.ptr)
		var val interface{}
		err := jsonpointer.Resolve(r, test.ptr, &val)
		if test.expectederr != nil {
			assert.ErrorIs(err, test.expectederr)
		} else {
			assert.NoError(err)
		}
		assert.Equal(test.expectedval, val)
		fmt.Printf("---PASS\n")
	}
}

func TestResolveMapIndex(t *testing.T) {
	assert := require.New(t)

	r := Root{
		Nested: Nested{
			CustomMap: map[Key]string{
				{"foo"}: "bar",
				{"baz"}: "qux",
			},
			EntryMap: map[string]*Entry{
				"foo": {
					Name:  "bar",
					Value: 34.34,
				},
			},
		},
	}

	tests := []struct {
		ptr         jsonpointer.JSONPointer
		expectedval interface{}
		expectederr error
	}{
		{"/nested/entrymap/foo/name", "bar", nil},
		{"/nested/custommap/foo", "bar", nil},
	}

	for i, test := range tests {
		fmt.Printf("=== RUN TestResolveMapIndex #%d, pointer %s\n", i, test.ptr)
		var val interface{}
		err := jsonpointer.Resolve(r, test.ptr, &val)
		if test.expectederr != nil {
			assert.ErrorIs(err, test.expectederr, "test %d", i)
		} else {
			assert.NoError(err, "test %d", i)
		}
		assert.Equal(test.expectedval, val, "test %d", i)
		fmt.Printf("--- PASS\n")
	}
}

func TestResolveBadMapKey(t *testing.T) {
	assert := require.New(t)
	var s string
	r := Root{
		Nested: Nested{
			EntryMap: map[string]*Entry{},
		},
	}

	if err, ok := jsonpointer.AsError(jsonpointer.Resolve(r, "/nested/entrymap/x/name", &s)); ok {
		t, ok := err.Token()
		assert.True(ok)
		assert.Equal("x", t.String())
	} else {
		assert.Fail("expected jsonpointer.Error")
	}
}

func TestResolveArray(t *testing.T) {
	assert := require.New(t)
	r := Root{
		Nested: Nested{
			StrArray: [3]string{"foo", "bar", ""},
			IntArray: [3]int{30, 31, 0},
		},
	}

	tests := []struct {
		ptr         jsonpointer.JSONPointer
		expectedval interface{}
		expectederr error
	}{
		{"/nested/strarray/0", "foo", nil},
		{"/nested/strarray/1", "bar", nil},
		{"/nested/strarray/2", "", nil},
		{"/nested/strarray/-", "", nil},
		{"/nested/strarray/3", nil, jsonpointer.ErrOutOfRange},
		{"/nested/intarray/0", 30, nil},
		{"/nested/intarray/1", 31, nil},
		{"/nested/intarray/2", 0, nil},
		{"/nested/intarray/-", 0, nil},
		{"/nested/intarray/3", nil, jsonpointer.ErrOutOfRange},
	}

	for i, test := range tests {
		fmt.Printf("=== RUN TestResolveArray #%d, pointer %s\n", i, test.ptr)

		var val interface{}
		err := jsonpointer.Resolve(r, test.ptr, &val)
		assert.ErrorIs(err, test.expectederr, "test %d", i)

		assert.Equal(test.expectedval, val, "test %d", i)
		fmt.Println("--- PASS")
	}
}

func TestResolveArrayOutOfRange(t *testing.T) {
	assert := require.New(t)
	r := Root{
		Nested: Nested{
			IntArray: [3]int{30, 31, 32},
			StrArray: [3]string{"foo", "bar", "baz"},
		},
	}
	var v interface{}

	err := jsonpointer.Resolve(r, "/nested/intarray/-", &v)
	assert.Error(err)

	err = jsonpointer.Resolve(r, "/nested/strarray/3", &v)
	assert.Error(err)
}

func TestResolveSlice(t *testing.T) {
	assert := require.New(t)
	r := Root{
		Nested: Nested{
			StrSlice: []string{"foo", "bar", "baz"},
			IntSlice: []int{30, 31, 0},
		},
	}

	tests := []struct {
		ptr         jsonpointer.JSONPointer
		expectedval interface{}
		expectederr error
	}{
		{"/nested/strslice/0", "foo", nil},
		{"/nested/strslice/1", "bar", nil},
		{"/nested/strslice/2", "baz", nil},
		{"/nested/strslice/-", nil, jsonpointer.ErrOutOfRange},
		{"/nested/strslice/3", nil, jsonpointer.ErrOutOfRange},
		{"/nested/intslice/0", 30, nil},
		{"/nested/intslice/1", 31, nil},
		{"/nested/intslice/2", 0, nil},
		{"/nested/intslice/-", nil, jsonpointer.ErrOutOfRange},
		{"/nested/intslice/3", nil, jsonpointer.ErrOutOfRange},
	}

	for i, test := range tests {
		fmt.Printf("=== RUN TestResolveArray #%d, pointer %s\n", i, test.ptr)
		var val interface{}
		err := jsonpointer.Resolve(r, test.ptr, &val)
		assert.ErrorIs(err, test.expectederr, "test %d", i)
		assert.Equal(test.expectedval, val, "test %d", i)
		fmt.Println("--- PASS")
	}
}

func TestResolveJSON(t *testing.T) {
	assert := require.New(t)

	tests := []struct {
		ptr  jsonpointer.JSONPointer
		json string
		val  interface{}
		err  error
	}{
		{"/nested/str", `{"nested":{"str":"foo"}}`, "foo", nil},
		{"/nested", `{"nested":{"str":"foo"}}`, []byte(`{"str":"foo"}`), nil},
	}

	for i, test := range tests {
		fmt.Printf("=== RUN TestResolveJSON #%d, pointer %s\n", i, test.ptr)
		vt := reflect.TypeOf(test.val)
		rv := reflect.New(vt).Elem()
		v := rv.Interface()
		err := jsonpointer.Resolve([]byte(test.json), test.ptr, &v)
		if test.err != nil {
			assert.Error(err)
		} else {
			assert.NoError(err)
			assert.Equal(test.val, v)
		}
		fmt.Println("--- PASS")
	}
}
