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
