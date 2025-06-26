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
			t.Parallel()

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

type SomeError struct {
	*CustomError
}

func (e *SomeError) Unwrap() error {
	return e.CustomError
}

func TestEmbeddedCustomError(t *testing.T) {
	var args []any
	args = append(args, "key", "val")

	err := &SomeError{
		CustomError: NewError("err", args...).(*CustomError),
	}

	var someError *SomeError
	assert.True(t, errors.As(err, &someError))

	var customError *CustomError
	assert.True(t, errors.As(someError, &customError))

	assert.Equal(t, "err", customError.Error())
	assert.Equal(t, "key", customError.Args[0])
	assert.Equal(t, "val", customError.Args[1])
	assert.NotNil(t, customError.Stack)
}
