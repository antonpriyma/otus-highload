package utils

import (
	"strings"

	"github.com/antonpriyma/otus-highload/pkg/errors"
)

func Split2(s string, sep string) (string, string) {
	res := strings.SplitN(s, sep, 2)

	switch len(res) {
	case 1:
		return res[0], ""
	case 2:
		return res[0], res[1]
	default:
		panic(errors.Errorf("unexpected length of slice after split: %d", len(res)))
	}
}
