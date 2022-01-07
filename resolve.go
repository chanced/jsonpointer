package jsonpointer

import "reflect"

type Resolver interface {
	ResolveJSONPointer(ptr *JSONPointer, op Operation) (interface{}, error)
}

func Resolve(src interface{}, ptr JSONPointer, dst interface{}) error {
	if err := ptr.Validate(); err != nil {
		return err
	}
	dv := reflect.ValueOf(dst)
	s := newState(ptr, Resolving)
	defer s.Release()
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
	return s.setValue(dv, v)
}
