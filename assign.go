package jsonpointer

import (
	"reflect"
)

type Assigner interface {
	AssignByJSONPointer(ptr *JSONPointer, value interface{}) error
}

func Assign(dst interface{}, ptr JSONPointer, value interface{}) error {
	if err := ptr.Validate(); err != nil {
		return err
	}

	//  not sure whether or not to delete this. Leaving it out for now.
	//
	// if ptr == TopLevel {
	// 	return &ptrError{
	// 		err: ErrEmptyJSONPointer,
	// 		typ: reflect.TypeOf(src),
	// 	}
	// }

	if value == nil {
		return Delete(dst, ptr)
	}
	dv := reflect.ValueOf(dst)
	s := newState(ptr, Assigning)
	defer s.Done()
	if dv.Kind() != reflect.Ptr || dv.IsNil() {
		return &ptrError{
			state: *s,
			err:   ErrNonPointer,
			typ:   dv.Type(),
		}
	}

	var tmp reflect.Value
	if dv.Type().Elem().Kind() == reflect.Ptr {
		tmp = dv
		if dv.Elem().IsNil() {
			dv = reflect.New(dv.Type().Elem().Elem())
		} else {
			dv = dv.Elem()
		}
	}
	res, err := s.assign(dv, reflect.ValueOf(value))
	tmp.Elem().Set(res)
	return err
}
