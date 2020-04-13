package utils

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestLogTrace(t *testing.T) {
	var buf bytes.Buffer
	l := NewIOLog(
		&buf,
		os.Stdout,
		os.Stdout,
		os.Stderr,
	)

	l.Trace.Println("Tracing")
	bufStr := buf.String()

	if !strings.HasPrefix(bufStr, "\n[TRACE] ") {
		t.Errorf(
			"Invalid buffer contents %s",
			buf.String(),
		)
	}
	if !strings.HasSuffix(bufStr, "log_test.go:20: Tracing\n") {
		t.Errorf(
			"Invalid buffer contents %s",
			buf.String(),
		)
	}
}
