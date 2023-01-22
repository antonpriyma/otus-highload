package echoerrors

import (
	"fmt"
	"net/http"

	"github.com/antonpriyma/otus-highload/pkg/errors"
	"github.com/labstack/echo"
)

type ResponseError interface {
	error
	errors.TypedError
	HTTPError() *echo.HTTPError
}

type httpError struct {
	Cause error
	Code  int
	ErrorMessage
}

type ErrorMessage struct {
	Type    string        `json:"type"`
	Explain string        `json:"explain,omitempty"`
	Fields  messageFields `json:"fields,omitempty"`
}

type messageFields map[string]interface{}

func (e httpError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("http %s error: %s", e.Type, e.Cause)
	}

	if e.Explain != "" {
		return fmt.Sprintf("http %s error: %s", e.Type, e.Explain)
	}
	return fmt.Sprintf("http %s error", e.Type)
}

func (e httpError) Unwrap() error {
	return e.Cause
}

func (e httpError) HTTPError() *echo.HTTPError {
	// Echo unwraps HTTPError internal error for some reason.
	// Wrap it to avoid this unpredicted behaviour
	if _, ok := e.Cause.(*echo.HTTPError); ok {
		e.Cause = errors.New(e.Cause.Error())
	}

	return &echo.HTTPError{
		Code:     e.Code,
		Internal: e.Cause,
		Message:  e.ErrorMessage,
	}
}

func (e httpError) ErrorType() string {
	return e.Type
}

func InternalError(
	cause error,
) ResponseError {
	return httpError{
		Cause: cause,
		Code:  http.StatusInternalServerError,
		ErrorMessage: ErrorMessage{
			Type:    "internal",
			Explain: "internal server error",
		},
	}
}

func CanceledError(err error) ResponseError {
	return httpError{
		Cause: err,
		Code:  499,
		ErrorMessage: ErrorMessage{
			Type: "canceled",
		},
	}
}

func ValidationError(
	cause error,
	explain string,
	fields ValidationErrorFields,
) ResponseError {
	msgFields := make(messageFields, len(fields))
	for k, v := range fields {
		msgFields[k] = v
	}

	return httpError{
		Cause: cause,
		Code:  http.StatusBadRequest,
		ErrorMessage: ErrorMessage{
			Type:    "validation",
			Explain: explain,
			Fields:  msgFields,
		},
	}
}

type ValidationErrorFields map[string]FieldError

type FieldError string

const (
	FieldForbidden   FieldError = "forbidden"
	FieldInvalid     FieldError = "invalid"
	FieldNotFound    FieldError = "not_found"
	FieldConflict    FieldError = "conflict"
	FieldRequired    FieldError = "required"
	FieldRequiredAny FieldError = "required_any"
)

func NotFoundError(
	cause error,
	key string,
) ResponseError {
	return httpError{
		Cause: cause,
		Code:  http.StatusNotFound,
		ErrorMessage: ErrorMessage{
			Type:    "not_found",
			Explain: "not found",
			Fields: messageFields{
				key: "not_found",
			},
		},
	}
}

func AlreadyExistsError(
	cause error,
	key string,
) ResponseError {
	return httpError{
		Cause: cause,
		Code:  http.StatusConflict,
		ErrorMessage: ErrorMessage{
			Type:    "already_exists",
			Explain: "already_exists",
			Fields: messageFields{
				key: "already_exists",
			},
		},
	}
}

type ForbiddenReason string

const (
	ReasonCSRFInvalid ForbiddenReason = "csrf_invalid"
	ReasonSpam        ForbiddenReason = "spam"
	ReasonNoAccess    ForbiddenReason = "no_access"
)

func ForbiddenError(
	cause error,
	reason ForbiddenReason,
	explain string,
) ResponseError {
	return httpError{
		Cause: cause,
		Code:  http.StatusForbidden,
		ErrorMessage: ErrorMessage{
			Type:    "forbidden",
			Explain: explain,
			Fields: messageFields{
				"reason": string(reason),
			},
		},
	}
}

type UnauthorizedReason string

const (
	ReasonTokenInvalid           UnauthorizedReason = "token_invalid"
	ReasonServersideTokenInvalid UnauthorizedReason = "serverside_token_invalid"
	ReasonNoServersideToken      UnauthorizedReason = "no_serverside_token"
	ReasonNoSDC                  UnauthorizedReason = "no_sdc"
	ReasonMpopInvalid            UnauthorizedReason = "mpop_invalid"
	ReasonNoCookies              UnauthorizedReason = "no_cookies"
	ReasonAuthError              UnauthorizedReason = "auth_error"
	ReasonAnonInvalid            UnauthorizedReason = "anon_invalid"
)

func UnauthorizedError(
	cause error,
	reason UnauthorizedReason,
	explain string,
) ResponseError {
	return httpError{
		Cause: cause,
		Code:  http.StatusUnauthorized,
		ErrorMessage: ErrorMessage{
			Type:    "unauthorized",
			Explain: explain,
			Fields: messageFields{
				"reason": string(reason),
			},
		},
	}
}

func TooManyRequestsError(
	cause error,
) ResponseError {
	return httpError{
		Cause: cause,
		Code:  http.StatusTooManyRequests,
		ErrorMessage: ErrorMessage{
			Type:    "too_many_requests",
			Explain: "too many requests, try again later",
		},
	}
}

func RetryLater(
	cause error,
) ResponseError {
	return httpError{
		Cause: cause,
		Code:  http.StatusTooManyRequests,
		ErrorMessage: ErrorMessage{
			Type:    "retry_later",
			Explain: "retry again later",
		},
	}
}
