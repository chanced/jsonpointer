package jsonpointer

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

type Embedded struct {
	F3 string `json:"f3"`
}

type structentry struct {
	Name      string `json:"name"`
	F2        string `json:"f2"`
	*Embedded `json:",inline"`
}

type mapcontainer struct {
	Map map[string]structentry `json:"map"`
}

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

func TestAssignMap(t *testing.T) {
	assert := require.New(t)
	var m map[string]structentry
	ptr := JSONPointer("/foo")
	s := newState(ptr, Assigning|Resolving)
	val := structentry{Name: "fooval"}
	err := s.assign(reflect.ValueOf(&m), val)
	assert.NoError(err)
	assert.Contains(m, "foo")
	assert.Equal(val, m["foo"])
}

// if r, ok := v.Interface().(Resolver); ok {
// 	result.resolver = r
// 	if s.isOnlyResolving() {
// 		result.value = v
// 		result.typ = v.Type()
// 		return result
// 	}
// }
