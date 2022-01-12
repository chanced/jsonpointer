package jsonpointer_test

import (
	"encoding/json"
	"fmt"

	"github.com/chanced/jsonpointer"
)

func ExampleNew() {
	ptr := jsonpointer.New("foo", "bar") // => "/foo/bar"
	fmt.Println(`"` + ptr + `"`)

	ptr = jsonpointer.New("foo/bar") // => "/foo~1bar"
	fmt.Println(`"` + ptr + `"`)

	ptr = jsonpointer.New() // => ""
	fmt.Println(`"` + ptr + `"`)

	ptr = jsonpointer.New("") // => "/"
	fmt.Println(`"` + ptr + `"`)

	ptr = jsonpointer.New("/") // => "/~1"
	fmt.Println(`"` + ptr + `"`)

	ptr = jsonpointer.New("~") // => "/~0"
	fmt.Println(`"` + ptr + `"`)

	ptr = jsonpointer.New("#/foo/bar") // => "/#~1foo~1bar"
	fmt.Println(`"` + ptr + `"`)

	// Output:
	// "/foo/bar"
	// "/foo~1bar"
	// ""
	// "/"
	// "/~1"
	// "/~0"
	// "/#~1foo~1bar"
}

func ExampleAssign() {
	type Bar struct {
		Baz string `json:"baz"`
	}
	type Foo struct {
		Bar Bar `json:"bar"`
	}
	var foo Foo
	jsonpointer.Assign(&foo, "/bar/baz", "qux")
	fmt.Println(foo.Bar.Baz)

	// Assigning JSON by JSONPointer

	foo.Bar.Baz = "quux"
	b, _ := json.Marshal(foo)
	jsonpointer.Assign(&b, "/bar/baz", "corge")
	fmt.Println(string(b))

	//Output: qux
	//{"bar":{"baz":"corge"}}
}

func ExampleResolve() {
	type Bar struct {
		Baz string `json:"baz"`
	}
	type Foo struct {
		Bar Bar `json:"bar,omitempty"`
	}
	foo := Foo{Bar{Baz: "qux"}}

	var s string
	jsonpointer.Resolve(foo, "/bar/baz", &s)
	fmt.Println(s)

	// Resolving JSON by JSONPointer

	b, _ := json.Marshal(foo)
	jsonpointer.Resolve(b, "/bar/baz", &s)
	fmt.Println(s)

	// Output: qux
	// qux
}

func ExampleDelete() {
	type Bar struct {
		Baz string `json:"baz,omitempty"`
	}
	type Foo struct {
		Bar Bar `json:"bar"`
	}
	foo := Foo{Bar{Baz: "qux"}}

	jsonpointer.Delete(foo, "/bar/baz")
	fmt.Printf("foo.Bar.Baz: %v\n", foo.Bar.Baz)

	// Deleting JSON by JSONPointer
	foo.Bar.Baz = "quux"
	b, _ := json.Marshal(foo)
	jsonpointer.Delete(&b, "/bar/baz")
	fmt.Println(string(b))

	// Output: foo.Bar.Baz: qux
	// {"bar":{}}
}
