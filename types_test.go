package jsonpointer_test

import (
	"fmt"

	"github.com/chanced/jsonpointer"
)

type Root struct {
	Nested    Nested  `json:"nested"`
	NestedPtr *Nested `json:"nestedptr"`
}

type Nested struct {
	Inline         `json:",inline"`
	Embedded       `json:"embedded"`
	InterContainer InterContainer         `json:",inline"`
	Nested         *Nested                `json:"nested"`
	String         string                 `json:"string"`
	Int            int                    `json:"int"`
	IntPtr         *int                   `json:"intptr"`
	Float          float64                `json:"float"`
	FloatPtr       *float64               `json:"floatptr"`
	Bool           bool                   `json:"bool"`
	BoolPtr        *bool                  `json:"boolptr"`
	Uint64         uint64                 `json:"uint64"`
	EntryMap       map[string]*Entry      `json:"entrymap"`
	StrMap         map[string]string      `json:"strmap"`
	IntMap         map[int]int            `json:"intmap"`
	UintMap        map[uint64]uint64      `json:"uintmap"`
	CustomMap      map[Key]string         `json:"custommap"`
	AnonStruct     struct{ Value string } `json:"anon"`
	StrSlice       []string               `json:"strslice"`
	IntSlice       []int                  `json:"intslice"`
	EntrySlice     []*Entry               `json:"entryslice"`
	StrArray       [3]string              `json:"strarray"`
	IntArray       [3]int                 `json:"intarray"`
	AnonStructPtr  *struct {
		Value string `json:"value"`
	} `json:"anonPtr"`
}

type InterContainer struct {
	Interface Interface `json:"interface"`
}

type Interface interface {
	jsonpointer.Assigner
	jsonpointer.Deleter
	jsonpointer.Resolver
	Value() *uint64
}

type privateImpl struct {
	nested *struct {
		value *uint64
	}
}

func (pi *privateImpl) MarshalJSON() ([]byte, error) {
	if pi.nested == nil || pi.nested.value == nil {
		return []byte(`{"value": null}`), nil
	}
	return []byte(fmt.Sprintf(`{"value":%d}`, *pi.nested.value)), nil
}

func (pi *privateImpl) AssignByJSONPointer(ptr *jsonpointer.JSONPointer, v interface{}) error {
	t, ok := ptr.NextToken()
	if !ok {
		panic("token not available? pointer: " + ptr.String())
	}
	if t != "value" {
		panic("unexpected pointer: " + ptr.String())
	}
	if v == nil {
		pi.nested = nil
		return nil
	}
	pi.nested = &struct {
		value *uint64
	}{
		value: v.(*uint64),
	}
	return nil
}

func (pi *privateImpl) DeleteByJSONPointer(ptr *jsonpointer.JSONPointer) error {
	t, ok := ptr.NextToken()
	if !ok {
		panic("token not available? pointer: " + ptr.String())
	}
	if t != "value" {
		panic("unexpected ptr: " + ptr.String())
	}
	pi.nested = nil
	return nil
}

func (pi *privateImpl) ResolveJSONPointer(ptr *jsonpointer.JSONPointer, op jsonpointer.Operation) (interface{}, error) {
	t, ok := ptr.NextToken()
	if !ok {
		panic("token not available? pointer: " + ptr.String())
	}
	if t != "value" {
		panic("unexpected ptr: " + ptr.String())
	}
	if pi.nested == nil {
		return nil, nil
	}
	return pi.nested.value, nil
}

func (p *privateImpl) Value() *uint64 {
	if p.nested == nil {
		return nil
	}
	return p.nested.value
}

var _ Interface = (*privateImpl)(nil)

type PublicImpl struct {
	value *uint64
}

func (pi *PublicImpl) AssignByJSONPointer(ptr *jsonpointer.JSONPointer, v interface{}) error {
	t, ok := ptr.NextToken()
	if !ok {
		panic("token not available? pointer: " + ptr.String())
	}
	if t != "value" {
		panic("unexpected pointer: " + ptr.String())
	}

	pi.value = v.(*uint64)
	return nil
}

func (pi *PublicImpl) DeleteByJSONPointer(ptr *jsonpointer.JSONPointer) error {
	t, ok := ptr.NextToken()
	if !ok {
		panic("token not available? pointer: " + ptr.String())
	}
	if t != "value" {
		panic("unexpected ptr: " + ptr.String())
	}
	pi = nil
	return nil
}

func (pi *PublicImpl) ResolveJSONPointer(ptr *jsonpointer.JSONPointer, op jsonpointer.Operation) (interface{}, error) {
	t, ok := ptr.NextToken()
	if !ok {
		panic("token not available? pointer: " + ptr.String())
	}
	if t != "value" {
		panic("unexpected ptr: " + ptr.String())
	}
	if pi == nil {
		return nil, nil
	}
	return pi.value, nil
}

func (p PublicImpl) Value() *uint64 {
	return p.value
}

var _ Interface = (*PublicImpl)(nil)

type Key struct {
	key string
}

func (k *Key) UnmarshalText(data []byte) error {
	k.key = string(data)
	return nil
}

func (k *Key) MarshalText() ([]byte, error) {
	return []byte(k.key), nil
}

type Entry struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

type Inline struct {
	InlineStr string `json:"inlineStr"`
}

type Embedded struct {
	EmbeddedStr string `json:"embeddedStr"`
}
