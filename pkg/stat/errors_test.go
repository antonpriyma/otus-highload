package stat

import (
	"context"
	"net"
	"testing"

	"github.com/antonpriyma/otus-highload/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestTypedErrorLabel(t *testing.T) {
	cases := []struct {
		name     string
		inputErr error
		expected string
	}{
		{
			name:     "no error",
			inputErr: nil,
			expected: "ok",
		},
		{
			name: "sourced typed",
			inputErr: errors.WithSource(
				errors.Typed("internal", "smth returned internal error"),
				"smth",
			),
			expected: "smth.internal",
		},
		{
			name: "wrapped sourced typed",
			inputErr: errors.Wrap(
				errors.WithSource(
					errors.Typed("internal", "smth returned internal error"),
					"smth",
				),
				"wrapped",
			),
			expected: "smth.internal",
		},
		{
			name: "sourced net error",
			inputErr: errors.WithSource(
				&net.AddrError{Err: "unknown network", Addr: "0.0.0.0"},
				"service",
			),
			expected: "service.net_fail",
		},
		{
			name:     "fail label",
			inputErr: errors.WithSource(errors.New("error"), "smth"),
			expected: "smth.fail",
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			label := TypedErrorLabel(context.Background(), c.inputErr)
			require.Equal(t, c.expected, label)
		})
	}
}

func TestTypedErrorLabelContextCanceled(t *testing.T) {
	cases := []struct {
		name     string
		inputErr error
		expected string
	}{
		{
			name:     "context canceled",
			inputErr: errors.New("little error"),
			expected: "context_canceled",
		},
		{
			name: "sourced error and context canceled",
			inputErr: errors.WithSource(
				errors.New("little error"),
				"source",
			),
			expected: "source.context_canceled",
		},
	}

	ctx := context.Background()
	ctx, canceler := context.WithCancel(ctx)
	canceler()

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			label := TypedErrorLabel(ctx, c.inputErr)
			require.Equal(t, c.expected, label)
		})
	}
}
