package json

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/antonpriyma/otus-highload/pkg/errors"
)

var (
	_ json.Marshaler   = Time{}
	_ json.Unmarshaler = &Time{}
)

type Time time.Time

func TimePointer(t time.Time) *Time {
	if t.IsZero() {
		return nil
	}

	v := Time(t)
	return &v
}

func (t *Time) Time() time.Time {
	if t == nil {
		return time.Time{}
	}

	return time.Time(*t)
}

const timeLayout = time.RFC3339

func (t Time) MarshalJSON() ([]byte, error) {
	if t.Time().IsZero() {
		return []byte(`null`), nil
	}

	return []byte(fmt.Sprintf(`"%s"`, t.Time().Format(timeLayout))), nil
}

func (t *Time) UnmarshalJSON(b []byte) error {
	s := string(b)
	if s == `null` {
		return nil
	}

	s = strings.Trim(s, `"`)

	parsed, err := time.Parse(timeLayout, s)
	if err != nil {
		return errors.Wrapf(err, "failed to unmarshal time string (%q) with layout (%q)", s, timeLayout)
	}

	*t = Time(parsed)
	return nil
}
