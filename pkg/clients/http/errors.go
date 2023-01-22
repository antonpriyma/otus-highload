package http

import (
	"fmt"

	"github.com/antonpriyma/otus-highload/pkg/errors"
)

var _ errors.TypedError = TypedError{}

type TypedError struct {
	Code    int
	Message string
}

func (e TypedError) Error() string {
	return fmt.Sprintf("bad http response: code: %d, message: %s", e.Code, e.Message)
}

func (e TypedError) ErrorType() string {
	return fmt.Sprintf("http_%d", e.Code)
}

var _ errors.TypedError = JSONError{}

type JSONError struct {
	Code    int
	Message string
}

func (e JSONError) Error() string {
	return fmt.Sprintf("bad json response: code: %d, message: %s", e.Code, e.Message)
}

func (e JSONError) ErrorType() string {
	return fmt.Sprintf("json_%d", e.Code)
}
