package jsonpointer

type Deleter interface {
	DeleteByJSONPointer(ptr *JSONPointer) error
}

func Delete(src interface{}, ptr JSONPointer) error {
	if err := ptr.Validate(); err != nil {
		return err
	}
	s := newState(ptr, Deleting)
	defer s.Release()
	panic("not done with Delete")
}
