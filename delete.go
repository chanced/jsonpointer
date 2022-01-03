package jsonpointer

type Deleter interface {
	DeleteByJSONPointer(ptr *JSONPointer) error
}

func Delete(src interface{}, ptr JSONPointer) error {
	panic("not impl")
}
