package jsonpointer_test

import (
	"testing"

	"github.com/chanced/jsonpointer"
	"github.com/stretchr/testify/require"
)

func TestResolveField(t *testing.T) {
	assert := require.New(t)
	floatv := float64(3.4)
	fl := &floatv

	r1 := Root{
		Nested: Nested{
			String:   "strval",
			FloatPtr: fl,
		},
	}
	var err error
	var sr string
	err = jsonpointer.Resolve(r1, "/nested/string", &sr)
	assert.NoError(err)
	assert.Equal("strval", sr)

	var floatres *float64
	err = jsonpointer.Resolve(r1, "/nested/floatptr", &floatres)
	assert.NoError(err)
	assert.Equal(fl, floatres)

	var floatptrvalue float64
	err = jsonpointer.Resolve(r1, "/nested/floatptr", &floatptrvalue)
	assert.ErrorIs(err, jsonpointer.ErrNotAssignable)
}

func TestResolveMapIndex(t *testing.T) {
	assert := require.New(t)

	r := Root{
		Nested: Nested{
			EntryMap: map[string]*Entry{
				"foo": {
					Name:  "bar",
					Value: 34.34,
				},
			},
		},
	}
	var s string
	err := jsonpointer.Resolve(r, "/nested/entrymap/foo/name", &s)
	assert.Equal("bar", s)
	assert.NoError(err)
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
			StrArray: [3]string{"foo", "bar", "baz"},
			IntArray: [3]int{30, 31, 32},
		},
	}

	var s string
	err := jsonpointer.Resolve(r, "/nested/strarray/1", &s)
	assert.NoError(err)
	assert.Equal("bar", s)

	var i int
	err = jsonpointer.Resolve(r, "/nested/intarray/1", &i)
	assert.ErrorIs(err, jsonpointer.ErrNotAssignable)
}
