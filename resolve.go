// Copyright 2022 Chance Dinkins
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
//
// The License can be found in the LICENSE file.
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package jsonpointer

import (
	"encoding/json"
	"reflect"
)

type Resolver interface {
	ResolveJSONPointer(ptr *JSONPointer, op Operation) (interface{}, error)
}

func Resolve(src interface{}, ptr JSONPointer, dst interface{}) error {
	dv := reflect.ValueOf(dst)
	s := newState(ptr, Resolving)
	defer s.Release()
	if err := ptr.Validate(); err != nil {
		return newError(err, *s, dv.Type())
	}
	if dv.Kind() != reflect.Ptr || dv.IsNil() {
		return &ptrError{
			state: *s,
			err:   ErrNonPointer,
			typ:   dv.Type(),
		}
	}
	v, err := s.resolve(reflect.ValueOf(src))
	if err != nil {
		return err
	}
	if isByteSlice(dv.Elem()) {
		b, err := json.Marshal(v.Interface())
		if err != nil {
			return newError(err, *s, dv.Type())
		}
		v = reflect.ValueOf(b)
	}
	return s.setValue(dv, v)
}
