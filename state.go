package jsonpointer

import (
	"encoding"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"sync"
)

var (
	// valinfoPool         sync.Pool
	statePool           sync.Pool
	jsonType            = reflect.TypeOf(RawJSON{})
	textUnmarshalerType = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
)

func newState(ptr JSONPointer, op Operation) *state {
	var s *state
	if v := statePool.Get(); v != nil {
		s = v.(*state)
	} else {
		s = &state{}
	}
	s.ptr = ptr
	s.current = ptr
	s.op = op
	return s
}

type state struct {
	op      Operation
	ptr     JSONPointer
	current JSONPointer
}

func (s state) Done() {
	go func() { statePool.Put(&s) }()
}

func (s state) Operation() Operation {
	return s.op
}

func (s *state) setValue(dst reflect.Value, val reflect.Value) error {
	switch dst.Kind() {
	case reflect.Interface:
		if !dst.CanInterface() {
			// TODO: remove this panic. return an error
			panic("can not interface")
		}
		return s.setValue(dst.Elem(), val)
	case reflect.Ptr:
		if val.Type().AssignableTo(dst.Type().Elem()) {
			dst.Elem().Set(val)
			return nil
		} else {
			return newValueError(ErrNotAssignable, s, dst.Type(), val.Type())
		}

	default:
		panic("can not assign to non-pointer")
	}
}

func (s *state) nextToken() (Token, error) {
	var t Token
	var ok bool
	if s.current, t, ok = s.current.Next(); !ok {
		return "", fmt.Errorf("unexpected end of JSON pointer %v", s.current)
	}
	return t, nil
}

func (s *state) resolve(v reflect.Value) (reflect.Value, error) {
	var t Token
	var err error
	for {
		if s.current.IsRoot() {
			return v, nil
		}

		typ := v.Type()
		t, err = s.nextToken()
		if err != nil {
			return v, newError(err, s, typ)
		}
		v, err = s.resolveNext(v, t)
		if err == nil && v.Kind() == reflect.Invalid {
			err = newError(ErrNotFound, s, typ)
		}
		if err != nil {
			s.current = s.current.Prepend(t)
			updateErrorState(err, s)
			return v, err
		}
	}
}

func (s *state) resolveNext(v reflect.Value, t Token) (reflect.Value, error) {
	if v.Type().NumMethod() > 0 && v.CanInterface() {
		if resolver, ok := v.Interface().(Resolver); ok {
			// storing the current pointer in the event the Resolver mutates it and
			// there is an error
			rv, err := s.resolveResolver(resolver)
			// if the Resolver returns an error, it can either be a YieldOperation
			// which continues the flow or an actual error.
			if err != nil {
				// if it isn't a yield operation, return the error
				if !errors.Is(err, YieldOperation) {
					return rv, newError(err, s, reflect.TypeOf(rv))
				}
				// otherwise restore the pointer to the previous state
			}
		}
	}
	switch v.Kind() {
	case reflect.Interface, reflect.Ptr:
		if v.IsNil() {
			return v, nil
		}
		return s.resolveNext(v.Elem(), t)
	case reflect.Map:
		return s.resolveMapIndex(v, t)
	case reflect.Array:
		return s.resolveArrayIndex(v, t)
	case reflect.Slice:
		return s.resolveSlice(v)
	case reflect.Struct:
		return s.resolveStructField(v, t)
	default:
		// TODO: handle other types or return an error
		panic("unexpected type: " + v.Type().String())
	}
}

func (s *state) resolveResolver(r Resolver) (reflect.Value, error) {
	res, err := r.ResolveJSONPointer(&s.current, s.op)
	if err != nil {
		if errors.Is(err, YieldOperation) {
			return reflect.Value{}, YieldOperation
		}
		rv := reflect.ValueOf(r)
		return rv, &ptrError{
			state: *s,
			err:   err,
			typ:   rv.Type(),
		}
	}
	return reflect.ValueOf(res), nil
}

func (s *state) resolveMapIndex(v reflect.Value, t Token) (reflect.Value, error) {
	kv, err := s.mapKey(v, t)
	if err != nil {
		return kv, err
	}
	return v.MapIndex(kv), nil
}

func (s *state) resolveArrayIndex(v reflect.Value, t Token) (reflect.Value, error) {
	panic("not implemented") // TODO: impl
}

func (s *state) resolveStructField(v reflect.Value, t Token) (reflect.Value, error) {
	var fields structFields
	if v.Type().Kind() == reflect.Ptr {
		fields = cachedTypeFields(v.Type().Elem())
	} else {
		fields = cachedTypeFields(v.Type())
	}
	var f *field
	if i, ok := fields.nameIndex[t.String()]; ok {
		f = &fields.list[i]
	} else {
		for i := range fields.list {
			tf := &fields.list[i]
			if tf.equalFold(tf.nameBytes, t.Bytes()) {
				f = tf
				break
			}
		}
	}
	if f == nil {
		return reflect.Value{}, newError(ErrNotFound, s, v.Type())
	}

	return v.FieldByIndex(f.index), nil
}

func (s *state) resolveSlice(v reflect.Value) (reflect.Value, error) {
	panic("not implemented") // TODO: impl
}

func (s *state) deleteMapIndex(v reflect.Value) error {
	panic("not impl") // TODO: impl
}

func (s state) JSONPointer() JSONPointer {
	return s.ptr
}

// CurrentJSONPointer returns the JSONPointer at the time of the error.
func (s state) CurrentJSONPointer() JSONPointer {
	return s.current
}

func (s *state) assign(src reflect.Value, value interface{}) error {
	sv, err := s.resolve(src)
	if err != nil {
		if isError(err) {
			return err
		}
		return newError(err, s, sv.Type())
	}
	v := reflect.ValueOf(value)
	_ = v
	// // if the current value implements Assigner then use it
	// if sv.assigner != nil {
	// 	return sv.assigner.AssignByJSONPointer(&s.current, value)
	// }

	// // if the value cannot be assigned, return an error
	// if !v.Type().AssignableTo(sv.Type()) {
	// 	return newValueError(ErrInvalidValue, s, sv.Type(), v.Type())
	// }
	// src.Set(sv.Value)
	// return nil
	// // getting the next pointer and token
	// p, t, ok := s.current.Next()
	// _ = t
	// if !ok {
	// 	// TODO: return error?
	// 	// this should never happen
	// 	panic("malformed token")
	// }

	if err != nil {
		return err
	}
	panic("assign not impl")

	// setting the current pointer to the new fragment
	// s.current = p
	// recurse into assign
	// if err = s.assign(nv, value); err != nil {
	// 	return err
	// }
	// // putting the fragment back
	// s.current = s.current.Append(t)
	// if !exists {
	// }
	return nil
}

func (s *state) delete(src reflect.Value) error {
	panic("not implemented") // TODO: impl
}

func (s *state) index(src reflect.Value, t Token) (reflect.Value, error) {
	switch src.Kind() {
	case reflect.Interface, reflect.Ptr:
		// return s.index(src.Elem(), t)
	case reflect.Array:
		return s.arrayIndex(src, t)
	case reflect.Slice:
		return s.sliceIndex(src, t)
	}
	panic("index not implemented")
}

func (s *state) sliceIndex(src reflect.Value, t Token) (reflect.Value, error) {
	i, err := t.Index(src.Len())
	_ = i
	_ = err
	panic("sliceIndex is not implemented")
}

func (s *state) arrayIndex(src reflect.Value, t Token) (reflect.Value, error) {
	// if t == "-" then we attempt to get the last non-zero index
	z := -1
	if t == "-" {
		for i := src.Len() - 1; i >= 0; i-- {
			if !src.Index(i).IsZero() {
				break
			}
			z = i
		}
		if z < 0 {
			return reflect.Value{}, newError(&indexError{
				err:      ErrOutOfRange,
				maxIndex: src.Len() - 1,
				index:    -1,
			}, s, src.Type())
		}
		return reflect.ValueOf(z), nil
	}

	z, err := t.Index(src.Type().Len())
	if err != nil {
		return reflect.Value{}, newError(err, s, src.Type())
	}
	return reflect.ValueOf(z), nil
}

func (s *state) mapKey(src reflect.Value, t Token) (reflect.Value, error) {
	kt := src.Type().Key()
	var kv reflect.Value
	// checks to see if the map's key implements encoding.TextUnmarshaler
	// if so, we use that to unmarshal the key
	if reflect.PtrTo(kt).Implements(textUnmarshalerType) {
		kv = reflect.New(kt)
		if err := kv.Interface().(encoding.TextUnmarshaler).UnmarshalText(t.Bytes()); err != nil {
			return kv, &keyError{
				ptrError: ptrError{
					err:   err,
					typ:   src.Type(),
					state: *s,
				},
				keyType:  kt,
				keyValue: kv,
			}
		}
		kv = kv.Elem()
		// otherwise the map's key must be either a string or an integer kind
	} else {
		switch kt.Kind() {
		case reflect.String:
			kv = reflect.ValueOf(t.String())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			i, err := t.Int64()
			if err != nil {
				return kv, &keyError{
					ptrError: ptrError{
						err:   err,
						typ:   src.Type(),
						state: *s,
					},
					keyType:  kt,
					keyValue: kv,
				}
			}
			kv = reflect.ValueOf(i).Convert(src.Type().Key())

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			u, err := t.Uint64()
			if err != nil {
				return kv, &keyError{
					ptrError: ptrError{
						err:   err,
						typ:   src.Type(),
						state: *s,
					},
					keyType:  kt,
					keyValue: kv,
				}
			}

			kv = reflect.ValueOf(u).Convert(src.Type().Key())
		}
	}
	if !kv.Type().AssignableTo(src.Type().Key()) {
		return kv, &keyError{
			ptrError: ptrError{
				err:   ErrInvalidKeyType,
				typ:   src.Type(),
				state: *s,
			},
			keyType:  kt,
			keyValue: kv,
		}
	}
	return kv, nil
}

func (s *state) setMapIndex(src reflect.Value, typ reflect.Type, v reflect.Value, token Token) error {
	panic("not impl")
	// kv, err := s.mapIndex(src, typ)
	// if err != nil {
	// 	return err
	// }
	// v.SetMapIndex(kv, v)
	// return nil
}

func (s *state) assignStruct() error {
	panic("not implemented") // TODO: impl
}

func (s *state) setStructField(t Token, typ reflect.Type, src reflect.Value, v reflect.Value) error {
	var fields structFields
	if typ.Kind() == reflect.Ptr {
		fields = cachedTypeFields(typ.Elem())
	} else {
		fields = cachedTypeFields(typ)
	}
	var f *field
	if i, ok := fields.nameIndex[t.String()]; ok {
		f = &fields.list[i]
	} else {
		for i := range fields.list {
			tf := &fields.list[i]
			if tf.equalFold(tf.nameBytes, t.Bytes()) {
				f = tf
				break
			}
		}
	}
	if f == nil {
		return &ptrError{
			err:   ErrInvalidKeyType,
			typ:   typ,
			state: *s,
		}
	}

	var subsrc reflect.Value
	subsrc = src
	for _, i := range f.index {
		if subsrc.Kind() == reflect.Ptr {
			if subsrc.IsNil() {
				if !subsrc.CanSet() {
					return &ptrError{
						state: *s,
						typ:   typ,
						err: fmt.Errorf(
							"jsonpointer: cannot set embedded pointer to unexported struct: %v",
							subsrc.Type().Elem(),
						),
					}
				}
				subsrc.Set(reflect.New(subsrc.Type().Elem()))
			}
			subsrc = subsrc.Elem()
		}
		subsrc = subsrc.Field(i)
	}

	if !subsrc.CanSet() {
		return &ptrError{
			state: *s,
			typ:   typ,
			err: fmt.Errorf(
				"jsonpointer: cannot set embedded pointer to unexported struct: %v",
				subsrc.Type().Elem(),
			),
		}
	}
	subsrc.Set(v)
	return nil
}

// type encoderFunc func(e *encodeState, v reflect.Value, opts encOpts)

// A field represents a single field found in a struct.
type field struct {
	name      string
	nameBytes []byte                 // []byte(name)
	equalFold func(s, t []byte) bool // bytes.EqualFold or equivalent
	tag       bool
	index     []int
	typ       reflect.Type
}

type structFields struct {
	list      []field
	nameIndex map[string]int
}

// cachedTypeFields is like typeFields but uses a cache to avoid repeated work.
//
// source: encoding/json/encode.go
func cachedTypeFields(t reflect.Type) structFields {
	if f, ok := fieldCache.Load(t); ok {
		return f.(structFields)
	}
	f, _ := fieldCache.LoadOrStore(t, typeFields(t))
	return f.(structFields)
}

var fieldCache sync.Map // map[reflect.Type]structFields

// typeFields returns a list of fields that JSON should recognize for the given type.
// The algorithm is breadth-first search over the set of structs to include - the top struct
// and then any reachable anonymous structs.
//
// source: encoding/json/encode.go
func typeFields(t reflect.Type) structFields {
	// Anonymous fields to explore at the current level and the next.
	current := []field{}
	next := []field{{typ: t}}

	// Count of queued names for current level and the next.
	var count, nextCount map[reflect.Type]int

	// Types already visited at an earlier level.
	visited := map[reflect.Type]bool{}

	// Fields found.
	var fields []field

	for len(next) > 0 {
		current, next = next, current[:0]
		count, nextCount = nextCount, map[reflect.Type]int{}

		for _, f := range current {
			if visited[f.typ] {
				continue
			}
			visited[f.typ] = true

			// Scan f.typ for fields to include.
			for i := 0; i < f.typ.NumField(); i++ {
				sf := f.typ.Field(i)
				if sf.Anonymous {
					t := sf.Type
					if t.Kind() == reflect.Ptr {
						t = t.Elem()
					}
					if !sf.IsExported() && t.Kind() != reflect.Struct {
						// Ignore embedded fields of unexported non-struct types.
						continue
					}
					// Do not ignore embedded fields of unexported struct types
					// since they may have exported fields.
				} else if !sf.IsExported() {
					// Ignore unexported non-embedded fields.
					continue
				}
				tag := sf.Tag.Get("json")
				if tag == "-" {
					continue
				}

				name, _ := parseTag(tag)

				if !isValidTag(name) {
					name = ""
				}
				index := make([]int, len(f.index)+1)
				copy(index, f.index)
				index[len(f.index)] = i

				ft := sf.Type
				if ft.Name() == "" && ft.Kind() == reflect.Ptr {
					// Follow pointer.
					ft = ft.Elem()
				}
				// Record found field and index sequence.
				if name != "" || !sf.Anonymous || ft.Kind() != reflect.Struct {
					tagged := name != ""
					if name == "" {
						name = sf.Name
					}
					field := field{
						name:  name,
						tag:   tagged,
						index: index,
						typ:   ft,
					}
					field.nameBytes = []byte(field.name)
					field.equalFold = foldFunc(field.nameBytes)

					fields = append(fields, field)
					if count[f.typ] > 1 {
						// If there were multiple instances, add a second,
						// so that the annihilation code will see a duplicate.
						// It only cares about the distinction between 1 or 2,
						// so don't bother generating any more copies.
						fields = append(fields, fields[len(fields)-1])
					}
					continue
				}

				// Record new anonymous struct to explore in next round.
				nextCount[ft]++
				if nextCount[ft] == 1 {
					next = append(next, field{name: ft.Name(), index: index, typ: ft})
				}
			}
		}
	}

	sort.Slice(fields, func(i, j int) bool {
		x := fields
		// sort field by name, breaking ties with depth, then
		// breaking ties with "name came from json tag", then
		// breaking ties with index sequence.
		if x[i].name != x[j].name {
			return x[i].name < x[j].name
		}
		if len(x[i].index) != len(x[j].index) {
			return len(x[i].index) < len(x[j].index)
		}
		if x[i].tag != x[j].tag {
			return x[i].tag
		}
		return byIndex(x).Less(i, j)
	})

	// Delete all fields that are hidden by the Go rules for embedded fields,
	// except that fields with JSON tags are promoted.

	// The fields are sorted in primary order of name, secondary order
	// of field index length. Loop over names; for each name, delete
	// hidden fields by choosing the one dominant field that survives.
	out := fields[:0]
	for advance, i := 0, 0; i < len(fields); i += advance {
		// One iteration per name.
		// Find the sequence of fields with the name of this first field.
		fi := fields[i]
		name := fi.name
		for advance = 1; i+advance < len(fields); advance++ {
			fj := fields[i+advance]
			if fj.name != name {
				break
			}
		}
		if advance == 1 { // Only one field with this name
			out = append(out, fi)
			continue
		}
		dominant, ok := dominantField(fields[i : i+advance])
		if ok {
			out = append(out, dominant)
		}
	}

	fields = out
	sort.Sort(byIndex(fields))

	// for i := range fields {
	// 	f := &fields[i]
	// 	f.encoder = typeEncoder(typeByIndex(t, f.index))
	// }
	nameIndex := make(map[string]int, len(fields))
	for i, field := range fields {
		nameIndex[field.name] = i
	}
	return structFields{fields, nameIndex}
}

// dominantField looks through the fields, all of which are known to
// have the same name, to find the single field that dominates the
// others using Go's embedding rules, modified by the presence of
// JSON tags. If there are multiple top-level fields, the boolean
// will be false: This condition is an error in Go and we skip all
// the fields.
func dominantField(fields []field) (field, bool) {
	// The fields are sorted in increasing index-length order, then by presence of tag.
	// That means that the first field is the dominant one. We need only check
	// for error cases: two fields at top level, either both tagged or neither tagged.
	if len(fields) > 1 && len(fields[0].index) == len(fields[1].index) && fields[0].tag == fields[1].tag {
		return field{}, false
	}
	return fields[0], true
}

// byIndex sorts field by index sequence.
type byIndex []field

func (x byIndex) Len() int { return len(x) }

func (x byIndex) Swap(i, j int) { x[i], x[j] = x[j], x[i] }

func (x byIndex) Less(i, j int) bool {
	for k, xik := range x[i].index {
		if k >= len(x[j].index) {
			return false
		}
		if xik != x[j].index[k] {
			return xik < x[j].index[k]
		}
	}
	return len(x[i].index) < len(x[j].index)
}
