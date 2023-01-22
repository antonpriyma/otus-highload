package errors

import (
	"fmt"

	"github.com/pkg/errors" // nolint:depguard
)

var (
	New    = errors.New
	Errorf = errors.Errorf
	As     = errors.As

	Unwrap = errors.Unwrap
)

type ErrorWithStackTrace interface {
	error
	StackTrace() errors.StackTrace
}

func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}

	var stackTracedErr ErrorWithStackTrace
	if hasStackTrace := As(err, &stackTracedErr); hasStackTrace {
		return errors.WithMessage(err, message)
	}

	return errors.Wrap(err, message)
}

func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	var stackTracedErr ErrorWithStackTrace
	if hasStackTrace := As(err, &stackTracedErr); hasStackTrace {
		return errors.WithMessagef(err, format, args...)
	}

	return errors.Wrapf(err, format, args...)
}

func Is(err error, targets ...error) bool {
	for _, target := range targets {
		if errors.Is(err, target) {
			return true
		}
	}
	return false
}

var _ error = transformedError{}

type transformedError struct {
	TargetError   error
	OriginalError error
}

func (err transformedError) Error() string {
	return fmt.Sprintf("%s: %s", err.TargetError, err.OriginalError)
}

func (err transformedError) Is(another error) bool {
	return Is(err.TargetError, another)
}

func (err transformedError) As(dst interface{}) bool {
	return As(err.TargetError, dst)
}

func (err transformedError) Unwrap() error {
	return err.OriginalError
}

func Transform(originalError error, targerError error) error {
	if originalError == nil {
		return nil
	}

	return transformedError{
		TargetError:   targerError,
		OriginalError: originalError,
	}
}

var (
	_ FieldedError = fieldedError{}
	_ error        = fieldedError{}
)

type fieldedError struct {
	Key           string
	Value         interface{}
	OriginalError error
}

func (err fieldedError) Error() string {
	return fmt.Sprintf("[%s=(%s)] %s", err.Key, err.Value, err.OriginalError)
}

func (err fieldedError) Unwrap() error {
	return err.OriginalError
}

func (err fieldedError) Field() (string, interface{}) {
	return err.Key, err.Value
}

type FieldedError interface {
	error
	Field() (string, interface{})
}

func WithField(err error, key string, value interface{}) error {
	if err == nil {
		return nil
	}

	return fieldedError{
		Key:           key,
		Value:         value,
		OriginalError: err,
	}
}

func WithStack(err error) error {
	if err == nil {
		return nil
	}

	var stackedErr ErrorWithStackTrace
	if errors.As(err, &stackedErr) {
		return err
	}

	return errors.WithStack(err)
}

func Fields(err error) map[string]interface{} {
	ret := map[string]interface{}{}
	for unwrapped := Unwrap(err); unwrapped != nil; unwrapped = Unwrap(unwrapped) {
		if fieldedErr, ok := unwrapped.(FieldedError); ok {
			k, v := fieldedErr.Field()
			ret[k] = v
		}
	}

	return ret
}

type panicError struct{}

func (panicError) Error() string {
	return "panic"
}

func (panicError) ErrorType() string {
	return "panic"
}

var ErrPanic TypedError = &panicError{}

func RecoverError(r interface{}) error {
	if r == nil {
		return nil
	}

	err, ok := r.(error)
	if !ok {
		err = Errorf("%v", r)
	}

	err = WithStack(err)

	return Transform(err, ErrPanic)
}

func ExtractDeepestStacktracer(err error) (ret ErrorWithStackTrace) {
	for err != nil {
		tracer, ok := err.(ErrorWithStackTrace)
		if ok {
			ret = tracer
		}

		err = errors.Unwrap(err)
	}

	return ret
}

type ErrorWithSource struct {
	error
	source string
}

func WithSource(err error, source string) error {
	if err == nil {
		return nil
	}

	return &ErrorWithSource{
		error:  err,
		source: source,
	}
}

func (err ErrorWithSource) GetSource() string {
	return err.source
}

func (err ErrorWithSource) Error() string {
	return fmt.Sprintf("[%s]: %s", err.source, err.error)
}

func (err ErrorWithSource) Unwrap() error {
	return err.error
}
