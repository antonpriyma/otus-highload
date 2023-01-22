package json

import (
	"encoding/json"
	"testing"

	"github.com/antonpriyma/otus-highload/pkg/test"
	"github.com/stretchr/testify/require"
)

func TestInt_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		wantErr  bool
		expected Int
	}{
		{
			name:     "basic int",
			input:    []byte(`{"Int": 32}`),
			expected: 32,
		},
		{
			name:     "basic string",
			input:    []byte(`{"Int": "100500"}`),
			expected: 100500,
		},
		{
			name:    "malformed string string",
			input:   []byte(`{"Int": "fds"}`),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result struct {
				Int Int
			}
			err := json.Unmarshal(tt.input, &result)

			test.CheckError(t, err, nil, tt.wantErr)

			require.Equal(t, tt.expected, result.Int)
		})
	}
}
