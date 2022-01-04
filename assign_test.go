package jsonpointer_test

import (
	"testing"
)

func TestAssignStructField(t *testing.T) {
	// assert := require.New(t)
	// se := structentry{}
	// assign := newAssignState("/f3")
	// err := assign.structField(
	// 	JSONPointer("/f3"),
	// 	reflect.TypeOf(&se),
	// 	reflect.ValueOf(&se),
	// 	reflect.ValueOf("value"),
	// )

	// assert.NoError(err)
	// assert.Equal("value", se.F3)
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
