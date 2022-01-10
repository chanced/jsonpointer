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
