package debug

import (
	"fmt"

	"github.com/antonpriyma/otus-highload/pkg/errors"
	"github.com/antonpriyma/otus-highload/pkg/utils"
)

func StackTrace() string {
	err := errors.New("stack trace")
	return ErrorStackTrace(err)
}

const maxStackSize = 20

func ErrorStackTrace(err error) string {
	tracer := errors.ExtractDeepestStacktracer(err)
	if tracer == nil {
		return ""
	}

	trace := tracer.StackTrace()
	printSize := utils.Min(maxStackSize, len(trace))
	return fmt.Sprintf("%+v", trace[:printSize])
}
