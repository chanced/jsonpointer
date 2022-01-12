# jsonpointer - an [RFC 6901](https://datatracker.ietf.org/doc/html/rfc6901) implementation for Go

Package jsonpointer provides the ability to resolve, assign, and delete values
of any type, including raw JSON, by [JSON
Pointers](https://datatracker.ietf.org/doc/html/rfc6901).

[![GoDoc](https://img.shields.io/badge/godoc-reference-5272B4.svg?style=flat-square)](https://pkg.go.dev/github.com/chanced/jsonpointer)

## Motivation

jsonpointer was built to support
[github.com/chanced/openapi](https://github.com/chanced/openapi) but it may be
useful for others so it has been released as an independent package.

For the openapi package, I needed a way to resolve and assign JSON Pointers
against concrete types while also maintaining integrity of pointer values. All
existing JSON Pointer implementations for Go operate on `map[string]interface{}`
and `[]interface{}`, raw JSON, or both.

## Install

```bash
go get github.com/chanced/jsonpointer
```

## Usage

### General

```go
package main

import (
    "log"
    "encoding/json"
    "github.com/chanced/jsonpointer"
)

type Nested struct {
    Str string
}

type Root struct {
    Nested Nested
}

func main() {

    r := Root{ Nested: Nested{ Str: "nested str" }}

    // jsonpointer.JSONPointer is a string type so if you have a properly
    // formatted json pointer then you can simply convert it:
    //  ptr := jsonpointer.JSONPointer(myPointer)
    //  err := ptr.Validate()

    // Note: jsonpointer.New encodes each token's value.
    // "/" encodes to "~1" and "~" encodes to "~0" in compliance with RFC 6901.

    ptr := jsonpointer.New("nested", "str")

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


    // jsonpointer can also Resolve, Assign, and Delete JSON in []byte format.
    // This includes field values, such as those of type json.RawMessage.

    r.Nested.Str = "str val"

    b, err := json.Marshal(r)
    if err != nil {
        log.Fatal(err)
    }

    err = jsonpointer.Resolve(b, ptr, &s)
    if err != nil {
        log.Fatal(err)
    }
    log.Println(s) // outputs "str val"
}
```

### Interfaces

Package jsonpointer provides 3 interfaces: `Assigner`, `Resolver`, and
`Deleter`. Regardless of the operation, if `Resolver` is implemented, `ResolveJSONPointer` will be
called. `ResolveJSONPointer` should not have side effects. If resolving for an assignment, utilize the
pointer to infer which type should be assigned.

Both `AssignByJSONPointer` and `DeleteByJSONPointer` are invoked on the way back from the leaf.

All three methods are passed a pointer to the `jsonpointer.JSONPointer` so that
it can be modified. If you do not modify it, jsonpointer will assume the current
token was addressed and continue on.

If you wish to only handle some cases with the interfaces, return `jsonpointer.YieldOperation` to have the jsonpointer package resolve, assign, or delete as if the type did not implement the interface. Note that doing so results in changes to `ptr` being dismissed.

### JSONPointer methods

All methods return new values rather than modifying the pointer itself. If you wish to modify the pointer in one of the interface methods, you will need to reassign it: `*ptr = newPtrVal`

```go
func (mt MyType) ResolveJSONPointer(ptr *jsonpointer.JSONPointer, op Operation) (interface{}, error) {
    next, t, ok := ptr.Next()
    if !ok {
        // this will only occur if the ptr is a root token in this circumstance
        return mt
    }
    if op == jsonpointer.Assigning && t == "someInterface" {
        // maybe you need to know what comes after someInterface to
        // determine what implementation of someInterface to assign
        t, _ = next.NextToken()

        switch t {
        case "someIdentifier":
            // you could modify ptr if you felt so inclined: *ptr = next
            // but it is not needed in this scenario.
            return SomeImplementation{}, nil
        }
    }
    // otherwise hand resolution back over to jsonpointer
    return nil, jsonpointer.YieldOperation
}
```

## Errors

All errors returned from `Resolve`, `Assign`, and `Delete` will implement `Error`. A convenience function `AsError` exists to help extract out the details.

Depending on the cause, the error could also be `KeyError`, `IndexError`, `FieldError` with additional details. All have corresponding `As{Error}` functions.

Finally, all errors have associated Err instances that are wrapped, such as `ErrMalformedToken`, `ErrInvalidKeyType`, and so on.

See [errors.go for further details on errors](https://github.com/chanced/jsonpointer/blob/main/errors.go).

## Contributions & Issues

Contributions are always welcome. If you run into an issue, please open a issue
on github. If you would like to submit a change, feel free to open up a pull
request.

## Note on Performance

This package is reflect heavy. While it employs the same caching mechanics as
`encoding/json` to help alleviate some of the lookup costs, there will always be
a performance hit with reflection.

There are also probably plenty of ways to improve performance of the package.
Improvements or criticisms are always welcome.

With regards to raw JSON, `json.Marshal` and `json.Unmarshal` are utilized.
Ideally, in the future, that will change and the package will incoroprate the
encoding/decoding logic from `encoding/json` directly, thus skipping the need to
run through unneccesary logic.

## Alternative JSON Pointer Packages for Go

-   [github.com/dolmen-go/jsonptr](https://github.com/dolmen-go/jsonptr)
-   [github.com/qri-io/jsonpointer](https://github.com/qri-io/jsonpointer)
-   [github.com/xeipuuv/gojsonpointer](https://github.com/xeipuuv/gojsonpointer)
-   [github.com/go-openapi/jsonpointer](https://github.com/go-openapi/jsonpointer)

## License

[Apache 2.0](https://raw.githubusercontent.com/chanced/jsonpointer/main/LICENSE)
