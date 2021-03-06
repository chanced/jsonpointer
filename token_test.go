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
	"testing"

	"github.com/chanced/jsonpointer"
	"github.com/stretchr/testify/require"
)

func TestTokenIndex(t *testing.T) {
	assert := require.New(t)

	token := jsonpointer.Token("1")

	i, err := token.Int()
	assert.NoError(err)
	assert.Equal(1, i)
	token = jsonpointer.Token("-")

	i, err = token.Index(2)
	assert.NoError(err)
	assert.Equal(2, i)

	token = jsonpointer.Token("3")

	i, err = token.Index(1)
	assert.ErrorIs(err, jsonpointer.ErrOutOfRange)
	assert.Equal(-1, i)

	i, err = token.Index(-1)
	assert.Error(err)
	assert.Equal(-1, i)

	token = jsonpointer.Token("1")

	i, err = token.Index(3)
	assert.NoError(err)
	assert.Equal(i, 1)
}

func TestTokenString(t *testing.T) {
	assert := require.New(t)

	token := jsonpointer.Token("~0~1")
	assert.Equal("~/", token.String())
}

// func TestTokenIsIndexable(t *testing.T) {
// 	assert := require.New(t)
// 	tests := []struct {
// 		token    jsonpointer.Token
// 		expected bool
// 	}{
// 		{"1", true},
// 		{"-1", false},
// 		{"-", true},
// 		{"0", true},
// 		{"c", false},
// 		{"", false},
// 	}

// 	for i, t := range tests {
// 		fmt.Println("=== TestTokenIsIndexable #", i, "token:", t.token)
// 		assert.Equal(t.expected, t.token.IsIndexable(), "test %d", i)
// 		fmt.Println("--- PASS TestTokenIsIndexable #", i)
// 	}
// }
