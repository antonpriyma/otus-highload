package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/antonpriyma/otus-highload/pkg/errors"
	"github.com/antonpriyma/otus-highload/pkg/log"
)

func MustParseURL(raw string) *url.URL {
	u, err := url.Parse(raw)
	if err != nil {
		panic(errors.Wrapf(err, "failed to parse raw url %q", raw))
	}

	return u
}

func MustParseRequestURL(raw string) *url.URL {
	u, err := url.ParseRequestURI(raw)
	if err != nil {
		panic(errors.Wrapf(err, "failed to parse raw url %q", raw))
	}

	return u
}

func MustFprint(w io.Writer, s string) {
	_, err := fmt.Fprint(w, s)
	if err != nil {
		panic(errors.Wrap(err, "failed to write"))
	}
}

func MustFprintf(w io.Writer, format string, args ...interface{}) {
	_, err := fmt.Fprintf(w, format, args...)
	if err != nil {
		panic(errors.Wrap(err, "failed to write"))
	}
}

func MustParseTimeNano(s string) time.Time {
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		panic(err)
	}
	return t
}

func Must(logger log.Logger, err error, msg string) {
	if err != nil {
		logger.WithError(err).Fatal(msg)
	}
}

func MustMarshalJSON(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}
