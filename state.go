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
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"sync"
)

var (
	// valinfoPool         sync.Pool
	statePool           sync.Pool
	typeAssigner        = reflect.TypeOf((*Assigner)(nil)).Elem()
	typeByteSlice       = reflect.TypeOf([]byte{})
	typeReader          = reflect.TypeOf((*io.Reader)(nil)).Elem()
	typeWriter          = reflect.TypeOf((*io.Writer)(nil)).Elem()
	typeTextUnmarshaler = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
	typeAny             = reflect.TypeOf((*interface{})(nil)).Elem()
	typeAnySlice        = reflect.TypeOf([]interface{}{})
	typeAnyMap          = reflect.TypeOf(map[string]interface{}{})
)

func newState(ptr Pointer, op Operation) *state {
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
	ptr     Pointer
	current Pointer
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
	var cpy reflect.Value

	if isByteSlice(dst.Elem()) {
		cpy = dst
		if dst.Elem().Len() > 0 {
			dst, err = s.unmarshal(dst.Elem())
			if err != nil {
				return reflect.Value{}, err
			}
		} else {
			_, nt, ok := s.current.Next()
			if !ok {
				return reflect.Value{}, newError(ErrMalformedToken, *s, dst.Type())
			}
			if s.current.IsRoot() {
				dst = reflect.Zero(val.Type())
			} else {
				if _, err = nt.Index(0); err == nil {
					dst = reflect.MakeSlice(typeAnySlice, 0, 1)
				} else {
					dst = reflect.MakeMap(typeAnyMap)
				}
			}
			dp := reflect.New(dst.Type())
			dp.Elem().Set(dst)
			dst = dp
		}
	}

	// new dst
	var rn reflect.Value

	rn, err = s.resolveNext(dst, t)
	if err != nil {
		return rn, err
	}

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
			rn = reflect.Zero(dst.Type().Elem().Elem())
			if rn.Kind() == reflect.Ptr && rn.IsNil() {
				rn = reflect.New(rn.Type().Elem())
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
				return reflect.Value{}, newError(ErrMalformedToken, *s, dst.Type())
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
	if rn.CanAddr() {
		rn = rn.Addr()
	} else {
		pv := reflect.New(rn.Type())
		pv.Elem().Set(rn)
		rn = pv
	}
	var nv reflect.Value
	nv, err = s.assign(rn, val)

	if err != nil {
		return dst, err
	}

	s.current = s.current.Prepend(t)
	if cpy.IsValid() {
		if assigner, ok := asAssigner(cpy); ok {
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
	if assigner, ok := asAssigner(dst); ok {
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
	if cpy.IsValid() {
		rn, err = s.marshal(dst)
		if err != nil {
			return reflect.Value{}, err
		}
		cpy.Elem().SetBytes(rn.Elem().Bytes())
		return cpy, nil
	}
	return dst, nil
}

func (s *state) delete(dst reflect.Value) (reflect.Value, error) {
	var t Token
	var err error
	var ok bool
	cur := s.current
	if cur.IsRoot() {
		err := s.deleteValue(dst)
		return dst, err
	}

	s.current, t, ok = cur.Next()

	if !ok {
		return reflect.Value{}, fmt.Errorf("unexpected end of JSON pointer %v", cur)
	}
	var cpy reflect.Value

	if isByteSlice(dst.Elem()) {
		cpy = dst
		if dst.Elem().Len() > 0 {
			dst, err = s.unmarshal(dst.Elem())
			if err != nil {
				return reflect.Value{}, err
			}
		} else {
			_, nt, ok := s.current.Next()
			if !ok {
				return reflect.Value{}, newError(ErrMalformedToken, *s, dst.Type())
			}
			if s.current.IsRoot() {
				dst.SetBytes([]byte{})
			} else {
				if _, err = nt.Index(0); err == nil {
					dst = reflect.MakeSlice(typeAnySlice, 0, 1)
				} else {
					dst = reflect.MakeMap(typeAnyMap)
				}
			}
			dp := reflect.New(dst.Type())
			dp.Elem().Set(dst)
			dst = dp
		}
	}

	if cpy.IsValid() {
		if deleter, ok := asDeleter(cpy); ok {
			err = deleter.DeleteByJSONPointer(&cur)
			if err != nil {
				if !errors.Is(err, YieldOperation) {
					return dst, newError(err, *s, dst.Elem().Type())
				} else {
					cur = s.current
				}
			} else {
				return dst, nil
			}
			s.current = cur
		}
	}
	if deleter, ok := asDeleter(dst); ok {
		err = deleter.DeleteByJSONPointer(&cur)
		if err != nil {
			if !errors.Is(err, YieldOperation) {
				return dst, newError(err, *s, dst.Elem().Type())
			} else {
				cur = s.current
			}
		} else {
			return dst, nil
		}
		// updating state to reflect the new token if it was set by deleter
		s.current = cur
	}
	// new dst
	var rn reflect.Value
	rn, err = s.resolveNext(dst, t)
	// cur = s.current
	if err != nil {
		return rn, err
	}

	if !ok {
		return dst, newError(ErrMalformedToken, *s, dst.Type())
	}
	if rn.IsValid() && rn.CanInterface() && rn.Type() == typeAny && !rn.IsNil() {
		rn = rn.Elem()
	}

	switch rn.Kind() {
	case reflect.Interface:
		if !rn.IsNil() && rn.Type() == typeAny {
			rn = rn.Elem()
		}
	case reflect.Ptr:
		if rn.IsNil() {
			s.current = ""
		} else if s.current.IsRoot() {
			s.current = ""
			rn.Set(reflect.ValueOf(nil))
		}
	case reflect.Map:
		switch {
		case rn.IsNil():
			s.current = ""
		case s.current.IsRoot():
			s.current = ""
			err = s.deleteMapIndex(rn, t)
			if err != nil {
				return reflect.Value{}, err
			}
		}
	case reflect.Slice:
		switch {
		case rn.IsNil():
			s.current = ""
		case s.current.IsRoot():
			err = s.deleteSliceIndex(rn, t)
			if err != nil {
				return reflect.Value{}, err
			}
			s.current = ""
		}
	case reflect.Invalid:
		s.current = ""
		return dst, nil
	}
	if rn.CanAddr() {
		rn = rn.Addr()
	} else {
		pv := reflect.New(rn.Type())
		pv.Elem().Set(rn)
		rn = pv
	}
	var nv reflect.Value
	nv, err = s.delete(rn)

	if err != nil {
		return dst, err
	}
	cur = s.current
	s.current = s.current.Prepend(t)

	rn, err = s.assignValue(rn, nv.Elem())
	if err != nil {
		return rn, err
	}
	switch dst.Elem().Kind() {
	case reflect.Map:
		if cur.IsRoot() {
			err = s.deleteMapIndex(dst.Elem(), t)
		} else {
			err = s.setMapIndex(dst.Elem(), t, rn.Elem())
		}
	case reflect.Slice:
		if cur.IsRoot() {
			err = s.deleteSliceIndex(dst, t)
		} else {
			err = s.setSliceIndex(dst, t, rn.Elem())
		}
	}
	if err != nil {
		return reflect.Value{}, newError(err, *s, dst.Elem().Type())
	}
	if cpy.IsValid() {
		rn, err = s.marshal(dst)
		if err != nil {
			return reflect.Value{}, err
		}
		cpy.Elem().SetBytes(rn.Elem().Bytes())
		return cpy, nil
	}
	return dst, nil
}

func (s *state) setValue(dst reflect.Value, v reflect.Value) error {
	switch dst.Kind() {
	case reflect.Interface:
		if !dst.CanInterface() {
			return newError(ErrNotAssignable, *s, dst.Type())
		}
		return s.setValue(dst.Elem(), v)
	case reflect.Ptr:
		if v.Type().AssignableTo(dst.Type().Elem()) {
			dst.Elem().Set(v)
			return nil
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
	if len(v.Bytes()) == 0 {
		return reflect.Value{}, nil
	}
	err := json.Unmarshal(v.Bytes(), &i)
	if err != nil {
		return v, newError(err, s, reflect.TypeOf(v))
	}
	iv := reflect.ValueOf(i)
	ptr := reflect.New(iv.Type())
	ptr.Elem().Set(iv)
	return ptr, nil
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

func (s state) deleteMapIndex(v reflect.Value, t Token) error {
	return s.setMapIndex(v, t, reflect.Value{})
}

func (s state) JSONPointer() Pointer {
	return s.ptr
}

// CurrentJSONPointer returns the JSONPointer at the time of the error.
func (s state) CurrentJSONPointer() Pointer {
	return s.current
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

func (s *state) deleteValue(dst reflect.Value) error {
	if !dst.Elem().CanSet() {
		return newError(ErrNotAssignable, *s, dst.Type())
	}
	z := reflect.Zero(dst.Elem().Type())
	dst.Elem().Set(z)
	return nil
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

func (s *state) deleteSliceIndex(l reflect.Value, token Token) error {
	e := l.Elem()
	i, err := s.sliceIndex(e, token)
	if err != nil {
		return err
	}
	if i >= l.Elem().Len() {
		return nil
	}

	reflect.Copy(e.Slice(i, e.Len()), e.Slice(i+1, e.Len()))

	e.Index(e.Len() - 1).Set(reflect.Zero(e.Type()))
	e.SetLen(e.Len() - 1)
	l.Elem().Set(e)

	return nil
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
	if v.IsValid() && v.Kind() == reflect.Interface && v.Elem().IsValid() && v.Elem().Type().AssignableTo(typeByteSlice) {
		return true
	}
	return v.Type().AssignableTo(typeByteSlice)
}

func asAssigner(v reflect.Value) (Assigner, bool) {
	if v.Type().NumMethod() > 0 && v.CanInterface() {
		as, ok := v.Interface().(Assigner)
		if ok {
			return as, true
		}
	}
	if v.Kind() == reflect.Ptr {
		return asAssigner(v.Elem())
	}
	return nil, false
}

func asDeleter(v reflect.Value) (Deleter, bool) {
	if v.Type().NumMethod() > 0 && v.CanInterface() {
		del, ok := v.Interface().(Deleter)
		if ok {
			return del, true
		}
	}
	if v.Kind() == reflect.Ptr {
		return asDeleter(v.Elem())
	}
	return nil, false
}
