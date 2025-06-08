package log

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWrapError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		err           error
		args          []any
		expectedError string
		expectedArgs  []any
	}{
		{
			name: "nil error",
			err:  nil,
			args: []any{"arg1", "val1"},
		},
		{
			name:          "unwrapped error",
			err:           errors.New("error"),
			args:          []any{"arg1", "val1"},
			expectedError: "error",
			expectedArgs:  []any{"arg1", "val1"},
		},
		{
			name:          "CustomError",
			err:           NewError("error", "arg0", "val0"),
			args:          []any{"arg1", "val1"},
			expectedError: "error",
			expectedArgs:  []any{"arg0", "val0", "arg1", "val1"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := WrapError(test.err, test.args...)
			if err == nil {
				assert.Nil(t, test.err)
				return
			}
			require.NotNil(t, err)
			customError, ok := err.(*CustomError)
			require.True(t, ok)
			assert.Equal(t, test.expectedError, customError.Error())
			assert.Equal(t, test.expectedArgs, customError.Args)
			assert.NotNil(t, customError.Stack)
		})
	}
}
