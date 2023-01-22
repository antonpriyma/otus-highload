package json

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/antonpriyma/otus-highload/pkg/utils"
	"github.com/stretchr/testify/require"
)

func TestTimeMarshalJSON(t *testing.T) {
	cases := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name: "normal",
			input: struct {
				Time Time
			}{
				Time: Time(utils.MustParseTimeNano("2021-05-31T21:50:46+07:00")),
			},
			expected: `{"Time":"2021-05-31T21:50:46+07:00"}`,
		},
		{
			name: "null",
			input: struct {
				EmptyTime Time
			}{},
			expected: `{"EmptyTime":null}`,
		},
		{
			name: "omitempty",
			input: struct {
				NotOmitted string
				Omitted    *Time `json:",omitempty"`
			}{
				NotOmitted: "test",
			},
			expected: `{"NotOmitted":"test"}`,
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			data, err := json.Marshal(c.input)
			require.NoError(t, err)
			require.Equal(t, c.expected, string(data))
		})
	}
}

func TestTimeUnmarshalJSON(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected time.Time
	}{
		{
			name:     "value normal",
			input:    `{"Time":"2021-05-31T21:50:46+07:00"}`,
			expected: utils.MustParseTimeNano("2021-05-31T21:50:46+07:00"),
		},
		{
			name:     "value empty",
			input:    `{}`,
			expected: time.Time{},
		},
		{
			name:     "value null",
			input:    `{"Time":null}`,
			expected: time.Time{},
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			var pointerField struct {
				Time *Time
			}
			err := json.Unmarshal([]byte(c.input), &pointerField)
			require.NoError(t, err)
			require.Equal(t, c.expected, pointerField.Time.Time())

			var valueField struct {
				Time Time
			}
			err = json.Unmarshal([]byte(c.input), &valueField)
			require.NoError(t, err)
			require.Equal(t, c.expected, valueField.Time.Time())
		})
	}
}
