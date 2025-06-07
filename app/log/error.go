package log

import (
	"errors"
)

type CustomError struct {
	error
	Args  []any
	Stack []*CallInfo
}

func (ce *CustomError) Unwrap() error {
	return ce.error
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
		return nil
	} else {
		var existing *CustomError
		if errors.As(err, &existing) {
			newArgs := make([]any, 0, len(existing.Args)+len(args))
			newArgs = append(newArgs, existing.Args...)
			newArgs = append(newArgs, args...)
			newStack := make([]*CallInfo, 0, len(existing.Stack))
			newStack = append(newStack, existing.Stack...)
			return &CustomError{
				error: existing.Unwrap(),
				Args:  newArgs,
				Stack: newStack,
			}
		}
	}
	return &CustomError{
		error: err,
		Args:  args,
		Stack: GetCallStack(2, 0),
	}
}
