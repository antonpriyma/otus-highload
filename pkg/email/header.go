package email

import (
	"net/textproto"

	"github.com/antonpriyma/otus-highload/pkg/errors"
	"github.com/emersion/go-message"
)

type Header map[string][]string

// Add adds the key, value pair to the header.
// It appends to any existing values associated with key.
// The key is case insensitive; it is canonicalized by
// CanonicalHeaderKey.
func (h Header) Add(key, value string) {
	textproto.MIMEHeader(h).Add(key, value)
}

// Set sets the header entries associated with key to the
// single element value. It replaces any existing values
// associated with key. The key is case insensitive; it is
// canonicalized by textproto.CanonicalMIMEHeaderKey.
// To use non-canonical keys, assign to the map directly.
func (h Header) Set(key, value string) {
	textproto.MIMEHeader(h).Set(key, value)
}

// Get gets the first value associated with the given key. If
// there are no values associated with the key, Get returns "".
// It is case insensitive; textproto.CanonicalMIMEHeaderKey is
// used to canonicalize the provided key. To use non-canonical keys,
// access the map directly.
func (h Header) Get(key string) string {
	return textproto.MIMEHeader(h).Get(key)
}

func (h Header) MessageHeaders() message.Header {
	var ret message.Header
	h.ExtendMessageHeaders(&ret)
	return ret
}

func (h Header) ExtendMessageHeaders(msgHeaders *message.Header) {
	for field, values := range h {
		for _, value := range values {
			msgHeaders.Add(field, value)
		}
	}
}

type HeaderGetter interface {
	Get(string) string
}

func ValidateHeaders(h HeaderGetter) error {
	contentType := h.Get(HeaderContentType)
	if contentType == "" {
		return errors.Errorf("header Content-Type is needed")
	}

	return nil
}
