package jsonpointer_test

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/chanced/jsonpointer"
)

type Root struct {
	Nested    Nested  `json:"nested"`
	NestedPtr *Nested `json:"nestedptr"`
}

type Nested struct {
	Inline         `json:",inline"`
	Embedded       `json:"embedded"`
	private        string
	InterContainer InterContainer         `json:"interface"`
	Nested         *Nested                `json:"nested,omitempty"`
	Empty          *Nested                `json:"empty,omitempty"`
	Str            string                 `json:"str,omitempty"`
	Int            int                    `json:"int,omitempty"`
	IntPtr         *int                   `json:"intptr,omitempty"`
	Float          float64                `json:"float,omitempty"`
	FloatPtr       *float64               `json:"floatptr,omitempty"`
	Bool           bool                   `json:"bool"`
	BoolPtr        *bool                  `json:"boolptr,omitempty"`
	Uint           uint                   `json:"uint,omitempty"`
	EntryMap       map[string]*Entry      `json:"entrymap,omitempty"`
	StrMap         map[string]string      `json:"strmap,omitempty"`
	IntMap         map[int]int            `json:"intmap,omitempty"`
	UintMap        map[uint]uint          `json:"uintmap,omitempty"`
	CustomMap      map[Key]string         `json:"custommap,omitempty"`
	AnonStruct     struct{ Value string } `json:"anon,omitempty"`
	StrSlice       []string               `json:"strslice,omitempty"`
	IntSlice       []int                  `json:"intslice,omitempty"`
	EntrySlice     []*Entry               `json:"entryslice,omitempty"`
	StrArray       [3]string              `json:"strarray,omitempty"`
	IntArray       [3]int                 `json:"intarray,omitempty"`
	Yield          Yield                  `json:"yield"`
	Raw            json.RawMessage        `json:"raw"`
	AnonStructPtr  *struct {
		Value string
	} `json:"anonptr"`
}

type InterContainer struct {
	Interface Interface `json:",inline"`
}

func (ic *InterContainer) AssignByJSONPointer(ptr *jsonpointer.JSONPointer, v interface{}) error {
	switch typ := v.(type) {
	case Interface:
		ic.Interface = typ
		return nil
	default:
		panic("unexpected type: " + reflect.TypeOf(v).String())
	}
}

func (ic InterContainer) ResolveJSONPointer(ptr *jsonpointer.JSONPointer, op jsonpointer.Operation) (interface{}, error) {
	p, t, ok := ptr.Next()
	if !ok {
		return nil, fmt.Errorf("unexpected root pointer: %s", ptr.String())
	}
	switch t {
	case "private":
		if in, ok := ic.Interface.(*privateImpl); ok {
			*ptr = p
			return in, nil
		}
		if op.IsAssigning() {
			*ptr = p
			x := &privateImpl{private: &struct{ value uint }{value: 5}}
			return x, nil
		}
	case "public":
		if in, ok := ic.Interface.(*PublicImpl); ok {
			*ptr = p
			return in, nil
		}
		if op.IsAssigning() {
			*ptr = p
			return &PublicImpl{}, nil
		}
	}
	panic("unexpected pointer: " + ptr.String())
}

type Interface interface {
	jsonpointer.Assigner
	jsonpointer.Deleter
	jsonpointer.Resolver
	Value() uint
}

type privateImpl struct {
	private *struct {
		value uint
	}
}

func (pi *privateImpl) MarshalJSON() ([]byte, error) {
	if pi.private == nil {
		return []byte(`{"value": null}`), nil
	}
	return []byte(fmt.Sprintf(`{"value":%d}`, pi.private.value)), nil
}

func (pi *privateImpl) AssignByJSONPointer(ptr *jsonpointer.JSONPointer, v interface{}) error {
	if v == nil {
		pi.private = nil
		return nil
	}
	pi.private = &struct {
		value uint
	}{
		value: v.(uint),
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
	pi.private = nil
	return nil
}

func (pi privateImpl) ResolveJSONPointer(ptr *jsonpointer.JSONPointer, op jsonpointer.Operation) (interface{}, error) {
	np, t, ok := ptr.Next()
	if !ok {
		panic("token not available? pointer: " + ptr.String())
	}
	if t != "value" {
		np, t, ok = np.Next()
		if !ok {
			return nil, fmt.Errorf("unexpected root pointer: %s", ptr.String())
		}
		*ptr = np
	}
	if t != "value" {
		panic("unexpected pointer: " + ptr.String())
	}
	if pi.private == nil {
		return uint(0), nil
	}
	return pi.private.value, nil
}

func (p *privateImpl) Value() uint {
	if p.private == nil {
		return uint(0)
	}
	return p.private.value
}

var _ Interface = (*privateImpl)(nil)

type PublicImpl struct {
	value uint
}

func (pi *PublicImpl) AssignByJSONPointer(ptr *jsonpointer.JSONPointer, v interface{}) error {
	t, ok := ptr.NextToken()
	if !ok {
		panic("token not available? pointer: " + ptr.String())
	}
	if t != "value" {
		panic("unexpected pointer: " + ptr.String())
	}

	pi.value = v.(uint)
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

func (p PublicImpl) Value() uint {
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

type Yield struct {
	Value string
}

func (y Yield) ResolveJSONPointer(p *jsonpointer.JSONPointer, op jsonpointer.Operation) (interface{}, error) {
	return nil, jsonpointer.YieldOperation
}
