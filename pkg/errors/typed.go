package errors

import "fmt"

type TypedError interface {
	error
	ErrorType() string
}

func Type(err error) string {
	var typed TypedError
	if As(err, &typed) {
		return typed.ErrorType()
	}

	return ""
}

func TypeStack(err error) []string {
	var stack []string
	for err != nil {
		typ := Type(err)
		if typ == "" {
			break
		}

		if len(stack) == 0 || stack[len(stack)-1] != typ {
			stack = append(stack, typ)
		}
		err = Unwrap(err)
	}

	return stack
}

func Typed(typ, msg string) TypedError {
	return &typedError{
		Type:    typ,
		Message: msg,
	}
}

func Typedf(typ, format string, args ...interface{}) TypedError {
	return &typedError{
		Type:    typ,
		Message: fmt.Sprintf(format, args...),
	}
}

type typedError struct {
	Type    string
	Message string
}

func (e typedError) Error() string {
	return fmt.Sprintf("type: %s, message: %s", e.Type, e.Message)
}

func (e typedError) ErrorType() string {
	return e.Type
}
