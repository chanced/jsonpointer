package jsonpointer

import "reflect"

type Assigner interface {
	AssignByJSONPointer(ptr *JSONPointer, value interface{}) error
}

func Assign(ptr JSONPointer, src interface{}, value interface{}) error {
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
		return Delete(src, ptr)
	}
	sv := reflect.ValueOf(src)
	s := newState(ptr, Assigning)
	defer s.Done()
	if sv.Kind() != reflect.Ptr || sv.IsNil() {
		return &ptrError{
			state: *s,
			err:   ErrNonPointer,
			typ:   sv.Type(),
		}
	}
	// TODO: Handle bytes / reader
	_, err := s.assign(sv, reflect.ValueOf(value))
	return err
}
