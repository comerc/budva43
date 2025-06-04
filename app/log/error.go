package log

import "errors"

type CustomError struct {
	error
	Args  []any
	Stack []*CallInfo
}

func NewError(text string, args ...any) error {
	return &CustomError{
		error: errors.New(text),
		Args:  args,
		Stack: GetCallStack(2, 0),
	}
}

func WrapError(err error, args ...any) error {
	if err == nil {
		err = errors.New("err is nil")
	} else {
		var customError *CustomError
		if errors.As(err, &customError) {
			customError.Args = append(customError.Args, args...)
			return customError
		}
	}
	return &CustomError{
		error: err,
		Args:  args,
		Stack: GetCallStack(2, 0),
	}
}
