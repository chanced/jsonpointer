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
	"reflect"
)

// Assigner is the interface implemented by types which can assign values via
// JSON Pointers. The input can be assumed to be a valid JSON Pointer and the
// value to assign.
//
// AssignByJSONPointer is called after the value has been resolved. If custom
// resolution is needed, the type should also implement Resolver.
//
type Assigner interface {
	AssignByJSONPointer(ptr *JSONPointer, value interface{}) error
}

// Assign performs an assignment of value to the target dst specified by the
// JSON Pointer ptr. Assign traverses dst recursively, resolving the path of
// the JSON Pointer. If a type in the path implements Resolver, it will attempt
// to resolve by invoking ResolveJSONPointer on that value. If ResolveJSONPointer
// returns YieldOperation or if the value does not implement ResolveJSONPointer,
// encoding/json naming conventions are utilized to resolve the path.
//
// If a type in the path implements Assigner, AssignByJSONPointer will be called
// with the updated value pertinent to that path.
//
func Assign(dst interface{}, ptr JSONPointer, value interface{}) error {
	if err := ptr.Validate(); err != nil {
		return err
	}
	if value == nil {
		return Delete(dst, ptr)
	}
	dv := reflect.ValueOf(dst)
	s := newState(ptr, Assigning)
	defer s.Release()
	if dv.Kind() != reflect.Ptr || dv.IsNil() {
		return &ptrError{
			state: *s,
			err:   ErrNonPointer,
			typ:   dv.Type(),
		}
	}
	cpy := dv
	dv = dv.Elem()
	if dv.Kind() == reflect.Ptr && dv.IsNil() {
		dv = reflect.New(dv.Type().Elem())
	}
	dp := reflect.New(dv.Type())
	dp.Elem().Set(dv)
	res, err := s.assign(dp, reflect.ValueOf(value))
	if err != nil {
		return err
	}
	cpy.Elem().Set(res.Elem())
	return nil
}
