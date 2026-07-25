// Package log provides minimal leveled logging to stderr for the provider.
//
// The provider intentionally avoids importing the full devsy SDK: it only ever
// shells out to limactl, so a tiny stderr logger keeps the dependency tree
// small and the binary lightweight.
package log

import (
	"fmt"
	"os"
)

var debug = os.Getenv("DEVSY_DEBUG") == "true"

// Debugf logs a debug message when DEVSY_DEBUG=true.
func Debugf(format string, args ...any) {
	if debug {
		fmt.Fprintf(os.Stderr, "DEBUG "+format+"\n", args...)
	}
}

// Infof logs an informational message to stderr.
//
// Provider commands communicate results over stdout, so all human-facing logs
// go to stderr to avoid corrupting command output (e.g. status).
func Infof(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "INFO  "+format+"\n", args...)
}

// Errorf logs an error message to stderr.
func Errorf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "ERROR "+format+"\n", args...)
}
