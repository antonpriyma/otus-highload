package test

import (
	"testing"

	"github.com/antonpriyma/otus-highload/pkg/errors"
	"github.com/stretchr/testify/require"
)

type testErr struct {
	Message string
}

func (e testErr) Error() string {
	return e.Message
}

type anotherTestErr struct {
	Message string
}

func (e anotherTestErr) Error() string {
	return e.Message
}

func TestIsExactError(t *testing.T) {
	pointerErr := error(&testErr{Message: "pointer error"})
	valueErr := error(testErr{Message: "value error"})

	for _, c := range []struct {
		name          string
		exactErr, err error
		expectedEqual bool
	}{
		{
			name:          "positive pointer error",
			exactErr:      pointerErr,
			err:           errors.Wrap(pointerErr, "some wrap"),
			expectedEqual: true,
		},
		{
			name:          "positive pointer error with the same content",
			exactErr:      &testErr{Message: "pointer error"},
			err:           errors.Wrap(pointerErr, "some wrap"),
			expectedEqual: true,
		},
		{
			name:          "negative pointer error with another content",
			exactErr:      &testErr{Message: "another pointer error"},
			err:           errors.Wrap(pointerErr, "some wrap"),
			expectedEqual: false,
		},
		{
			name:          "negative pointer error with another type",
			exactErr:      &anotherTestErr{Message: "pointer error"},
			err:           errors.Wrap(pointerErr, "some wrap"),
			expectedEqual: false,
		},
		{
			name:          "negative same error with value pointer mismatch",
			exactErr:      testErr{Message: "value error"},
			err:           errors.Wrap(pointerErr, "some wrap"),
			expectedEqual: false,
		},
		{
			name:          "positive value error with the same content",
			exactErr:      testErr{Message: "value error"},
			err:           errors.Wrap(valueErr, "some wrap"),
			expectedEqual: true,
		},
		{
			name:          "negative value error with another content",
			exactErr:      testErr{Message: "another value error"},
			err:           errors.Wrap(valueErr, "some wrap"),
			expectedEqual: false,
		},
		{
			name:          "negative value error with another type",
			exactErr:      anotherTestErr{Message: "value error"},
			err:           errors.Wrap(valueErr, "some wrap"),
			expectedEqual: false,
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			result := ExtractExactError(c.err, c.exactErr)
			if c.expectedEqual {
				require.Equal(t, c.exactErr, result)
			} else {
				require.NotEqual(t, c.exactErr, result)
			}
		})
	}
}
