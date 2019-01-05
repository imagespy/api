package log

import (
	"bytes"
	"fmt"

	"github.com/pkg/errors"
)

type stackTracer interface {
	StackTrace() errors.StackTrace
}

func FormatError(err error) string {
	stErr, ok := err.(stackTracer)
	if ok {
		b := &bytes.Buffer{}
		fmt.Fprintf(b, "%s\n", stErr)

		for _, f := range stErr.StackTrace() {
			fmt.Fprintf(b, "  %+v\n", f)
		}

		return b.String()
	}

	return fmt.Sprint(err)
}
