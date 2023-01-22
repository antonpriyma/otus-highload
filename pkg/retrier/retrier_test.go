package retrier

import (
	"context"
	"testing"
	"time"

	"github.com/antonpriyma/otus-highload/pkg/errors"
	"github.com/antonpriyma/otus-highload/pkg/test"
	"github.com/stretchr/testify/require"
)

const (
	maxRetries = 5
)

var errSimple = errors.New("test")

func makeTestRetryCancelGetter() (context.Context, func() error) {
	ctx, cancel := context.WithCancel(context.Background())

	callbackWithCancel := func() error {
		cancel()
		return ctx.Err()
	}

	return ctx, callbackWithCancel
}

func makeTestRetryNilGetter() (context.Context, func() error) {
	return context.Background(), func() error { return nil }
}

func makeTestRetryErrorGetter() (context.Context, func() error) {
	return context.Background(), func() error { return errSimple }
}

func makeTestRetryFixedGetter() (context.Context, func() error) {
	triesToFix := 1
	return context.Background(), func() error {
		if triesToFix == 0 {
			return nil
		}

		triesToFix--
		return errSimple
	}
}

func TestRetry(t *testing.T) {
	type testCase struct {
		name               string
		retrier            Retrier
		ctxCallGetter      func() (context.Context, func() error)
		expectedCallsCount int
		expectedError      error
	}

	maxLimitedRetrier := Retrier{
		Config: Config{
			Retries: maxRetries,
		},
	}

	testCases := []testCase{
		{
			name:               "successful callback",
			retrier:            maxLimitedRetrier,
			ctxCallGetter:      makeTestRetryNilGetter,
			expectedCallsCount: 1,
			expectedError:      nil,
		},
		{
			name:               "bad callback",
			retrier:            maxLimitedRetrier,
			ctxCallGetter:      makeTestRetryErrorGetter,
			expectedCallsCount: maxRetries + 1,
			expectedError:      errSimple,
		},
		{
			name:               "cancel context",
			retrier:            maxLimitedRetrier,
			ctxCallGetter:      makeTestRetryCancelGetter,
			expectedCallsCount: 1,
			expectedError:      context.Canceled,
		},
		{
			name:               "fixed callback",
			retrier:            maxLimitedRetrier,
			ctxCallGetter:      makeTestRetryFixedGetter,
			expectedCallsCount: 2,
			expectedError:      nil,
		},
		{
			name: "break by timeout",
			retrier: Retrier{
				Config: Config{
					Timeout: time.Second,
					Sleep:   300 * time.Millisecond,
				},
			},
			ctxCallGetter:      makeTestRetryErrorGetter,
			expectedCallsCount: 4,
			expectedError:      errSimple,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			ctx, call := testCase.ctxCallGetter()
			callsCount := 0

			wrappedCall := func() error {
				callsCount++
				return call()
			}

			err := testCase.retrier.Do(ctx, wrappedCall)
			test.CheckError(t, err, testCase.expectedError, false)
			require.Equal(t, testCase.expectedCallsCount, callsCount)
		})
	}
}
