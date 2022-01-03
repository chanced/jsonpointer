package jsonpointer

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

type structtest struct {
	Field1      string `json:"field1"`
	Field2      string `json:"field2"`
	Field3      string
	NotIncluded string `json:"-"`
}

func TestTypeFields(t *testing.T) {
	assert := require.New(t)
	e := structtest{}
	rt := reflect.TypeOf(e)
	fields := typeFields(rt)
	assert.NotEmpty(fields.list)
	assert.Contains(fields.nameIndex, "field1")
	assert.Contains(fields.nameIndex, "field2")
	assert.Equal(0, fields.nameIndex["field1"])
	assert.Equal(1, fields.nameIndex["field2"])
	assert.Equal(2, fields.nameIndex["Field3"])
	assert.Equal("field1", fields.list[0].name)
	assert.Equal("field2", fields.list[1].name)
	assert.Equal("Field3", fields.list[2].name)
	assert.NotContains(fields.nameIndex, "NotIncluded")
}
