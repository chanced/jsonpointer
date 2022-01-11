package main

import (
	"log"

	"github.com/chanced/jsonpointer"
)

type Nested struct {
	Str string
}

type Root struct {
	Nested Nested
}

func main() {
	r := Root{Nested: Nested{Str: "nested str"}}

	// JSONPointer is a string so if you have a properly formatted pointer,
	// simply convert it:
	//      ptr := jsonpointer.JSONPointer(myPointer)

	ptr := jsonpointer.New("nested", "str")
	// Note: jsonpointer.New does not validate and it encodes each token's value.
	// which means "/" is encoded to "~1" and "~" encodes to "~0" in compliance
	// with RFC 6901.

	// Resolve

	var s string

	err := jsonpointer.Resolve(r, ptr, &s)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(s) // outputs "nested str"

	// Assign

	err = jsonpointer.Assign(&r, ptr, "new value")
	if err != nil {
		log.Fatal(err)
	}
	log.Println(r.Nested.Str) // outputs "new value"

	// Delete

	err = jsonpointer.Delete(&r, ptr)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(r.Nested.Str) // outputs ""
}
