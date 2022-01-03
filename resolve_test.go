package jsonpointer_test

import (
	"testing"

	"github.com/chanced/jsonpointer"
	"github.com/stretchr/testify/require"
)

func TestResolve(t *testing.T) {
	assert := require.New(t)
	floatv := float64(3.4)
	fl := &floatv

	r1 := Root{
		Nested: Nested{
			String:   "string",
			FloatPtr: fl,
		},
	}
	var err error
	var sr string
	err = jsonpointer.Resolve(r1, "/nested/str", &sr)
	assert.NoError(err)
	assert.Equal("str", sr)

	var floatres *float64
	err = jsonpointer.Resolve(r1, "/nested/floatptr", &floatres)
	assert.NoError(err)
	assert.Equal(fl, floatres)

	var floatptrvalue float64
	err = jsonpointer.Resolve(r1, "/nested/floatptr", &floatptrvalue)
	assert.ErrorIs(err, jsonpointer.ErrNotAssignable)
}
