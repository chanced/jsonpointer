package jsonpointer

import (
	"bytes"
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"sync"

	"github.com/sanity-io/litter"
)

var (
	// valinfoPool         sync.Pool
	statePool           sync.Pool
	typeJSON            = reflect.TypeOf(JSON{})
	typeAssigner        = reflect.TypeOf((*Assigner)(nil)).Elem()
	typeByteSlice       = reflect.TypeOf([]byte{})
	typeReader          = reflect.TypeOf((*io.Reader)(nil)).Elem()
	typeWriter          = reflect.TypeOf((*io.Writer)(nil)).Elem()
	typeTextUnmarshaler = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
	typeAny             = reflect.TypeOf((*interface{})(nil)).Elem()
	typeAnySlice        = reflect.TypeOf([]interface{}{})
	typeAnyMap          = reflect.TypeOf(map[string]interface{}{})
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

func (s *state) Release() {
	statePool.Put(s)
}

func (s state) Operation() Operation {
	return s.op
}

func (s *state) resolve(v reflect.Value) (reflect.Value, error) {
	var t Token
	var err error
	var ok bool
	for {
		cur := s.current
		if cur.IsRoot() {
			return v, nil
		}
		typ := v.Type()
		if s.current, t, ok = cur.Next(); !ok {
			return reflect.Value{}, fmt.Errorf("unexpected end of JSON pointer %v", s.current)
		}
		if err != nil {
			return v, newError(err, *s, typ)
		}

		v, err = s.resolveNext(v, t)
		if err == nil {
			switch v.Kind() {
			case reflect.Interface, reflect.Ptr, reflect.Map, reflect.Slice:
				if v.IsNil() && !cur.IsRoot() {
					return v, newError(ErrUnreachable, *s, typ)
				}
			}
		}

		if err == nil && v.Kind() == reflect.Invalid {
			err = newError(ErrNotFound, *s, typ)
		}
		if err != nil {
			s.current = cur.Prepend(t)
			updateErrorState(err, *s)
			return v, err
		}
	}
}

func (s *state) assign(dst reflect.Value, val reflect.Value) (reflect.Value, error) {
	var t Token
	var err error
	var ok bool
	cur := s.current
	if cur.IsRoot() {
		_, err := s.assignValue(dst, val)
		return dst, err
	}

	s.current, t, ok = cur.Next()

	if !ok {
		return reflect.Value{}, fmt.Errorf("unexpected end of JSON pointer %v", cur)
	}
	var ds string
	var de interface{}
	var v interface{}
	fmt.Println("============= TKN =============")
	fmt.Println(t)
	if isByteSlice(dst.Elem()) {
		fmt.Println("============= DST =============")
		buf := &bytes.Buffer{}
		json.Compact(buf, dst.Elem().Bytes())
		ds = buf.String()
		fmt.Println(ds)
	} else {
		fmt.Println("============= DST =============")
		de = dst.Elem().Interface()
		litter.Dump(de)
	}
	fmt.Println("============= VAL =============")
	v = val.Interface()
	litter.Dump(v)
	// new dst
	var rn reflect.Value
	rn, err = s.resolveNext(dst, t)
	if err != nil {
		return rn, err
	}
	fmt.Println("============= REN =============")
	litter.Dump(rn.Elem().Interface())
	fmt.Printf("===============================\n")

	// shouldSet := false
	switch rn.Kind() {
	case reflect.Interface:
		if !rn.IsNil() && rn.Type() == typeAny {
			rn = rn.Elem()
		}
	case reflect.Ptr:
		if rn.IsNil() {
			rn.Set(reflect.New(rn.Type().Elem()))
		}
	case reflect.Map:
		if rn.IsNil() {
			rn.Set(reflect.MakeMap(rn.Type()))
		}
	case reflect.Slice:
		if rn.IsNil() {
			rn.Set(reflect.MakeSlice(rn.Type(), 0, 1))
		}
	case reflect.Invalid:
		switch dst.Type().Elem().Kind() {
		case reflect.Map, reflect.Slice:
			// shouldSet = true
			rn = reflect.Zero(dst.Type().Elem().Elem())
			if rn.Kind() == reflect.Ptr && rn.IsNil() {
				rn = reflect.New(rn.Type().Elem())
				// so this works
			} else if rn.Type() == typeAny && rn.IsNil() {
				nt, ok := s.current.NextToken()
				if !ok {
					rn = reflect.Zero(val.Type())
				} else {
					if _, err = nt.Index(0); err == nil {
						rn = reflect.MakeSlice(typeAnySlice, 0, 1)
					} else {
						rn = reflect.MakeMap(typeAnyMap)
					}
				}
			}
		case reflect.Interface:
			_, nt, ok := s.current.Next()
			if !ok {
				panic("pointer is not ok")
			}
			if rn.Type() == typeAny {
				if s.current.IsRoot() {
					rn = reflect.Zero(val.Type())
				} else {
					if _, err = nt.Index(0); err == nil {
						rn = reflect.MakeSlice(typeAnySlice, 0, 1)
					} else {
						rn = reflect.MakeMap(typeAnyMap)
					}
				}
			}
		default:
			return reflect.Value{}, newError(ErrUnreachable, *s, dst.Type())
		}
	}
func (s *state) setValue(dst reflect.Value, v reflect.Value) error {
	switch dst.Kind() {
	case reflect.Interface:
		if !dst.CanInterface() {
			return newError(ErrNotAssignable, *s, dst.Type())
		}
		return s.setValue(dst.Elem(), v)

	case reflect.Ptr:
		switch {
		case v.Type().AssignableTo(dst.Type().Elem()):
			dst.Elem().Set(v)
			return nil
		case v.Type().AssignableTo(typeReader):
			panic("reader not implemented")
		case v.Type().AssignableTo(typeByteSlice):
			// if src is []byte and dst is not then we unmarshal the json
			return json.Unmarshal(v.Interface().([]byte), dst.Interface())
		}
		// none of the above is true
		return newValueError(ErrNotAssignable, *s, dst.Type(), v.Type())
	default:
		// this should never be reached
		panic("can not assign to non-pointer")
	}
}

func (s state) marshal(v reflect.Value) (reflect.Value, error) {
	b, err := json.Marshal(v.Interface())
	if err != nil {
		return reflect.Value{}, newError(err, s, v.Type())
	}
	bv := reflect.New(typeByteSlice)
	bv.Elem().SetBytes(b)
	return bv, nil
}

func (s state) unmarshal(v reflect.Value) (reflect.Value, error) {
	var i interface{}
	err := json.Unmarshal(v.Bytes(), &i)
	if err != nil {
		return v, newError(err, s, reflect.TypeOf(v))
	}
	return reflect.ValueOf(i), nil
}

func (s *state) resolveNext(v reflect.Value, t Token) (reflect.Value, error) {
	var err error
	switch {
	case v.Type().NumMethod() > 0 && v.CanInterface():
		if resolver, ok := v.Interface().(Resolver); ok {
			rv, err := s.resolveResolver(resolver, v, t)
			if err != nil {
				if !errors.Is(err, YieldOperation) {
					return rv, err
				}
			} else {
				return rv, nil
			}
		}
	case isByteSlice(v):
		v, err = s.unmarshal(v)
		if err != nil {
			return v, err
		}
	case v.Kind() == reflect.Ptr && isByteSlice(v.Elem()):
		v, err = s.unmarshal(v.Elem())
		if err != nil {
			return v, err
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
		return s.resolveSlice(v, t)
	case reflect.Struct:
		return s.resolveStructField(v, t)
	default:
		return v, newError(ErrUnreachable, *s, v.Type())
	}
}

func (s *state) resolveResolver(r Resolver, rv reflect.Value, t Token) (reflect.Value, error) {
	// storing the current pointer in the event the Resolver mutates it and
	// there is an error
	prev := s.current
	cur := prev.Prepend(t)
	s.current = cur
	res, err := r.ResolveJSONPointer(&s.current, s.op)
	if err != nil {
		// if the Resolver returns an error, it can either be a YieldOperation
		// which continues the flow or an actual error.
		if !errors.Is(err, YieldOperation) {
			return rv, newError(err, *s, reflect.TypeOf(rv))
		}
		s.current = prev
		return rv, YieldOperation
	}
	if s.current == cur {
		// if the Resolver didn't mutate the pointer, then we
		// restore it to the original value
		s.current = prev
	}
	return reflect.ValueOf(res), nil
}

func (s state) resolveMapIndex(v reflect.Value, t Token) (reflect.Value, error) {
	kv, err := s.mapKey(v, t)
	if err != nil {
		return kv, err
	}
	return v.MapIndex(kv), nil
}

func (s state) resolveArrayIndex(v reflect.Value, t Token) (reflect.Value, error) {
	i, err := s.arrayIndex(v, t)
	if err != nil {
		return reflect.Value{}, err
	}
	return v.Index(i), nil
}

func (s state) resolveStructField(v reflect.Value, t Token) (reflect.Value, error) {
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
		fv, ok := v.Type().FieldByName(t.String())
		if ok && !fv.IsExported() {
			return reflect.Value{}, newError(ErrUnexportedField, s, v.Type())
		}
		return reflect.Value{}, newError(ErrNotFound, s, v.Type())
	}

	return v.FieldByIndex(f.index), nil
}

func (s state) resolveSlice(v reflect.Value, t Token) (reflect.Value, error) {
	i, err := s.sliceIndex(v, t)
	if err != nil {
		if errors.Is(err, strconv.ErrSyntax) {
			return reflect.Value{}, newError(ErrMalformedIndex, s, v.Type())
		}
		return reflect.Value{}, newError(err, s, v.Type())
	}

	if s.op == Resolving && i >= v.Len() {
		return reflect.Value{}, newError(&indexError{
			err:      ErrOutOfRange,
			maxIndex: v.Len() - 1,
			index:    i,
		}, s, v.Type())
	} else if i >= v.Len() {
		return reflect.Value{}, nil
	}
	return v.Index(i), nil
}

func (s state) deleteMapIndex(v reflect.Value) error {
	panic("not impl") // TODO: impl
}

func (s state) JSONPointer() JSONPointer {
	return s.ptr
}

// CurrentJSONPointer returns the JSONPointer at the time of the error.
func (s state) CurrentJSONPointer() JSONPointer {
	return s.current
}


	if rn.CanAddr() {
		rn = rn.Addr()
	} else {
		pv := reflect.New(rn.Type())
		pv.Elem().Set(rn)
		rn = pv
	}
	var nv reflect.Value
	nv, err = s.assign(rn, val)

	var nve interface{}

	fmt.Println("============= NEW =============")
	nve = nv.Elem().Interface()
	litter.Dump(nve)
	fmt.Printf("===============================\n")
	if err != nil {
		return dst, err
	}

	s.current = s.current.Prepend(t)
	if dst.Type().NumMethod() > 0 && dst.CanInterface() && dst.Type().Implements(typeAssigner) {
		if assigner, ok := dst.Interface().(Assigner); ok {
			err = assigner.AssignByJSONPointer(&cur, nv.Elem().Interface())
			if err != nil {
				if !errors.Is(err, YieldOperation) {
					return dst, newError(err, *s, dst.Elem().Type())
				} else {
					// the Assigner has yielded operation back to jsonpointer
					// resetting current incase the Assigner mutated it
					cur = s.current
				}
			} else {
				return dst, nil
			}
			// updating state to reflect the new token if it was set by assigner
			s.current = cur
		}
	} else if dst.Elem().Type().NumMethod() > 0 && dst.Elem().CanInterface() && dst.Type().Elem().Implements(typeAssigner) {
		if assigner, ok := dst.Elem().Interface().(Assigner); ok {
			err = assigner.AssignByJSONPointer(&cur, nv.Elem().Interface())
			if err != nil {
				if !errors.Is(err, YieldOperation) {
					return dst, newError(err, *s, dst.Elem().Type())
				} else {
					cur = s.current
				}
			} else {
				return dst, nil
			}
			// updating state to reflect the new token if it was set by assigner
			s.current = cur
		}
	}
	rn, err = s.assignValue(rn, nv.Elem())
	if err != nil {
		return rn, err
	}

	switch dst.Elem().Kind() {
	case reflect.Map:
		err = s.setMapIndex(dst.Elem(), t, rn.Elem())
	case reflect.Slice:
		err = s.setSliceIndex(dst, t, rn.Elem())
	}
	if err != nil {
		return reflect.Value{}, newError(err, *s, dst.Elem().Type())
	}
	return dst, nil
}

func (s *state) assignValue(dst reflect.Value, val reflect.Value) (reflect.Value, error) {
	if !dst.Elem().CanSet() {
		return dst, newValueError(ErrNotAssignable, *s, dst.Type(), val.Type())
	}
	switch {
	case val.Type().AssignableTo(dst.Elem().Type()):
		dst.Elem().Set(val)
		return dst, nil
	case isByteSlice(val.Elem()):
		err := json.Unmarshal(val.Elem().Bytes(), dst.Interface())
		if err != nil {
			return dst, newValueError(ErrNotAssignable, *s, dst.Type(), val.Type())
		}
		return dst, nil
	case isByteSlice(dst.Elem()):
		b, err := json.Marshal(val.Interface())
		if err != nil {
			return dst, newValueError(ErrNotAssignable, *s, dst.Type(), val.Type())
		}
		dst.Elem().SetBytes(b)
	}
	return val, newValueError(ErrNotAssignable, *s, dst.Elem().Type(), val.Type())
}

func (s *state) delete(src reflect.Value) error {
	panic("not implemented") // TODO: impl
}

func (s state) sliceIndex(src reflect.Value, t Token) (int, error) {
	i, err := t.Index(src.Len())
	return i, err
}

func (s state) arrayIndex(src reflect.Value, t Token) (int, error) {
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
			return -1, newError(&indexError{
				err:      ErrOutOfRange,
				maxIndex: src.Len() - 1,
				index:    -1,
			}, s, src.Type())
		}
		return z, nil
	}

	z, err := t.Index(src.Type().Len() - 1)
	if err != nil {
		return z, newError(err, s, src.Type())
	}
	return z, nil
}

func (s *state) mapKey(src reflect.Value, t Token) (reflect.Value, error) {
	kt := src.Type().Key()
	var kv reflect.Value
	// checks to see if the map's key implements encoding.TextUnmarshaler
	// if so, we use that to unmarshal the key
	if reflect.PtrTo(kt).Implements(typeTextUnmarshaler) {
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

func (s *state) setSliceIndex(l reflect.Value, token Token, v reflect.Value) error {
	i, err := s.sliceIndex(l.Elem(), token)
	if err != nil {
		return err
	}
	if i >= l.Elem().Len() {
		l.Elem().Set(reflect.Append(l.Elem(), v))
		return nil
	} else {
		nl := l.Elem()
		nl.Index(i).Set(v)
		l.Set(nl)
	}
	return nil
}

func (s *state) setArrayIndex(src reflect.Value, token Token, v reflect.Value) error {
	i, err := s.arrayIndex(src, token)
	if err != nil {
		return err
	}
	src.Index(i).Set(v)
	return nil
}

func (s *state) setMapIndex(m reflect.Value, token Token, v reflect.Value) error {
	kv, err := s.mapKey(m, token)
	if err != nil {
		return err
	}
	m.SetMapIndex(kv, v)
	return nil
}

func isByteSlice(v reflect.Value) bool {
	if v.Kind() == reflect.Interface && v.Elem().Type().AssignableTo(typeByteSlice) {
		return true
	}
	return v.Type().AssignableTo(typeByteSlice)
}
