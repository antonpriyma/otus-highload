package utils

import (
	"bytes"
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMustParseURL(t *testing.T) {
	rawURL := "http://mail.ru"
	expected, _ := url.Parse(rawURL) //nolint: errcheck

	got := MustParseURL(rawURL)
	require.Equal(t, expected, got)
}

func TestMustFprint(t *testing.T) {
	s := "keke"

	w := &bytes.Buffer{}
	MustFprint(w, s)
	require.Equal(t, s, w.String())
}

func TestMustFprintf(t *testing.T) {
	format := "keke %s"
	args := []interface{}{"1"}
	expected := fmt.Sprintf(format, args...)

	w := &bytes.Buffer{}
	MustFprintf(w, format, args...)
	require.Equal(t, expected, w.String())
}
