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

import "reflect"

// Deleter is an interface that is implemented by any type which can delete a
// value by JSON pointer.
type Deleter interface {
	DeleteByJSONPointer(ptr *Pointer) error
}

// Delete deletes the value at the given JSON pointer from src.
//
// If any part of the path is unreachable, the Delete function is
// considered a success as the value is not present to delete.
func Delete(src interface{}, ptr Pointer) error {
	dv := reflect.ValueOf(src)
	s := newState(ptr, Deleting)
	defer s.Release()
	if err := ptr.Validate(); err != nil {
		return newError(err, *s, dv.Type())
	}
	if dv.Kind() != reflect.Ptr || dv.IsNil() {
		return newError(ErrNonPointer, *s, dv.Type())
	}
	cpy := dv
	dv = dv.Elem()
	if dv.Kind() == reflect.Ptr && dv.IsNil() {
		dv = reflect.New(dv.Type().Elem())
	}
	dp := reflect.New(dv.Type())
	dp.Elem().Set(dv)
	res, err := s.delete(dp)
	if err != nil {
		return err
	}
	cpy.Elem().Set(res.Elem())
	return nil
}
