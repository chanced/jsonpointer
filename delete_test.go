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
	"fmt"
	"testing"

	"github.com/chanced/jsonpointer"
	"github.com/stretchr/testify/require"
)

func TestDelete(t *testing.T) {
	assert := require.New(t)

	tests := []struct {
		ptr  jsonpointer.JSONPointer
		root Root
		run  func(r Root, err error)
	}{
		{"/nested/str", Root{Nested: Nested{Str: "str val"}}, func(r Root, err error) {
			assert.NoError(err)
			assert.Equal("", r.Nested.Str)
		}},
	}
	for i, test := range tests {
		fmt.Printf("=== RUN TestDelete #%d, pointer %s\n", i, test.ptr)
		err := jsonpointer.Delete(&test.root, test.ptr)
		test.run(test.root, err)
	}
}
