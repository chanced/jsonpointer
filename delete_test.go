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

package jsonpointer_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/chanced/jsonpointer"
	"github.com/stretchr/testify/require"
)

func TestDelete(t *testing.T) {
	assert := require.New(t)

	tests := []struct {
		ptr  jsonpointer.Pointer
		root Root
		run  func(r Root, err error)
	}{
		{"/nested/str", Root{Nested: Nested{Str: "str val", Int: 5}}, func(r Root, err error) {
			assert.NoError(err)
			assert.Equal("", r.Nested.Str)
			assert.Equal(5, r.Nested.Int)
		}},
		{"/nested", Root{Nested: Nested{Str: "str val", Int: 5}}, func(r Root, err error) {
			assert.NoError(err)
			assert.Equal("", r.Nested.Str)
			assert.Equal(0, r.Nested.Int)
		}},
		{"/nested/deleter/key", Root{Nested: Nested{Deleter: DeleterImpl{Values: map[string]string{"key": "value"}}}}, func(r Root, err error) {
			assert.NotContains(r.Nested.Deleter.Values, "key")
			assert.NoError(err)
		}},
	}

	for i, test := range tests {
		fmt.Printf("=== RUN TestDelete #%d, pointer %s\n", i, test.ptr)
		err := jsonpointer.Delete(&test.root, test.ptr)
		test.run(test.root, err)

	}
}

func TestDeleteJSON(t *testing.T) {
	assert := require.New(t)

	tests := []struct {
		ptr  jsonpointer.Pointer
		root Root
		run  func(r Root, err error)
	}{
		{"/nested/str", Root{Nested: Nested{Str: "string", Int: 5}}, func(r Root, err error) {
			assert.NoError(err)
			assert.Equal("", r.Nested.Str)
			assert.Equal(5, r.Nested.Int)
		}},
		{"/nested", Root{Nested: Nested{Str: "str val", Int: 5}}, func(r Root, err error) {
			assert.NoError(err)
			assert.Equal("", r.Nested.Str)
			assert.Equal(0, r.Nested.Int)
		}},

		{"/nested/strslice/1", Root{Nested: Nested{StrSlice: []string{"0", "1", "2"}}}, func(r Root, err error) {
			assert.NoError(err)
			assert.Len(r.Nested.StrSlice, 2)
			assert.Equal("2", r.Nested.StrSlice[1])
		}},
	}

	for i, test := range tests {
		fmt.Printf("=== RUN TestDelete #%d, pointer %s\n", i, test.ptr)
		b, err := json.Marshal(test.root)
		fmt.Println(string(b))

		assert.NoError(err)
		err = jsonpointer.Delete(&b, test.ptr)
		var r Root
		uerr := json.Unmarshal(b, &r)
		assert.NoError(uerr)
		fmt.Println(string(b))
		test.run(r, err)
	}
}
