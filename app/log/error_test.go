package log

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrapError(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		args          []any
		expectedError string
		expectedArgs  []any
	}{
		{
			name:          "nil error",
			err:           nil,
			args:          []any{"arg1", "val1"},
			expectedError: "err is nil",
			expectedArgs:  []any{"arg1", "val1"},
		},
		{
			name:          "simple error",
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
			err, ok := WrapError(test.err, test.args...).(*CustomError)
			assert.True(t, ok)
			assert.Equal(t, test.expectedError, err.Error())
			assert.Equal(t, test.expectedArgs, err.Args)
			assert.NotNil(t, err.Stack)
		})
	}
}
