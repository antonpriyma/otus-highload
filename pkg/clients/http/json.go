package http

import (
	"strconv"
	"time"

	"github.com/antonpriyma/otus-highload/pkg/errors"
)

type TimestampUnix time.Time

func (t *TimestampUnix) UnmarshalJSON(b []byte) error {
	unixTime, err := strconv.ParseInt(string(b), 10, 64)
	if err != nil {
		return errors.Wrapf(err, "failed to unmarshall TimestampUnix: %q", b)
	}
	timestamp := time.Unix(unixTime, 0)

	*t = TimestampUnix(timestamp)
	return nil
}

func (t TimestampUnix) MarshalJSON() ([]byte, error) {
	b := []byte(strconv.FormatInt(time.Time(t).Unix(), 10))
	return b, nil
}

func (t TimestampUnix) String() string {
	return time.Time(t).UTC().Format(time.RFC3339)
}
