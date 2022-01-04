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
	InterContainer InterContainer         `json:"interface"`
	Nested         *Nested                `json:"nested,omitempty"`
	String         string                 `json:"string,omitempty"`
	Int            int                    `json:"int,omitempty"`
	IntPtr         *int                   `json:"intptr,omitempty"`
	Float          float64                `json:"float,omitempty"`
	FloatPtr       *float64               `json:"floatptr,omitempty"`
	Bool           bool                   `json:"bool"`
	BoolPtr        *bool                  `json:"boolptr,omitempty"`
	Uint64         uint64                 `json:"uint64,omitempty"`
	EntryMap       map[string]*Entry      `json:"entrymap,omitempty"`
	StrMap         map[string]string      `json:"strmap,omitempty"`
	IntMap         map[int]int            `json:"intmap,omitempty"`
	UintMap        map[uint64]uint64      `json:"uintmap,omitempty"`
	CustomMap      map[Key]string         `json:"custommap,omitempty"`
	AnonStruct     struct{ Value string } `json:"anon,omitempty"`
	StrSlice       []string               `json:"strslice,omitempty"`
	IntSlice       []int                  `json:"intslice,omitempty"`
	EntrySlice     []*Entry               `json:"entryslice,omitempty"`
	StrArray       [3]string              `json:"strarray,omitempty"`
	IntArray       [3]int                 `json:"intarray,omitempty"`
	AnonStructPtr  *struct {
		Value string
	} `json:"anonptr"`
}

type InterContainer struct {
	Interface `json:",inline"`
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

func (k Key) MarshalText() ([]byte, error) {
	return []byte(k.key), nil
}

type Entry struct {
	Name  string  `json:"name,omitempty"`
	Value float64 `json:"value,omitempty"`
}

type Inline struct {
	InlineStr string `json:"inline,omitempty"`
}

type Embedded struct {
	Value string `json:"value,omitempty"`
}
