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

	token = jsonpointer.Token("1")

	i, err = token.Index(3)
	assert.ErrorIs(err, jsonpointer.ErrOutOfRange)
	assert.Equal(-1, i)

	i, err = token.Index(-1)
	assert.Error(err)
	assert.Equal(-1, i)
}

func TestTokenString(t *testing.T) {
	assert := require.New(t)

	token := jsonpointer.Token("~0~1")
	assert.Equal("~/", token.String())
}
