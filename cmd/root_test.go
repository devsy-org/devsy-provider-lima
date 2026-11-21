package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/devsy-org/devsy/pkg/log"
)

const testErrorLevel = "error"

func TestLoggerLevelsAndRedaction(t *testing.T) {
	for _, tc := range []struct {
		level string
		want  string
	}{
		{"", "info,warn,error"},
		{testErrorLevel, testErrorLevel},
		{"warn", "warn,error"},
		{"info", "info,warn,error"},
		{"debug", "debug,info,warn,error"},
		{"trace", "trace,debug,info,warn,error"},
		{"TrAcE", "trace,debug,info,warn,error"},
	} {
		t.Run(tc.level, func(t *testing.T) {
			t.Setenv("DEVSY_LOG_LEVEL", tc.level)
			t.Setenv("DEVSY_LOG_FORMAT", "auto")
			t.Setenv("LIMA_TEST_SECRET", "secret-canary-123")
			if err := initLogger(); err != nil {
				t.Fatal(err)
			}
			var output bytes.Buffer
			remove := log.AddSink(&output)
			defer remove()
			log.Trace("trace secret-canary-123")
			log.Debug("debug secret-canary-123")
			log.Info("info secret-canary-123")
			log.Warn("warn secret-canary-123")
			log.Error("error secret-canary-123")
			if err := log.Sync(); err != nil {
				t.Fatal(err)
			}
			levels := assertRedactedDiagnostics(t, output.Bytes())
			if got := strings.Join(levels, ","); got != tc.want {
				t.Fatalf("levels = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestLoggerFormats(t *testing.T) {
	for _, format := range []string{"", "auto", "AUTO", "json", "JSON", "text", "plain", "PLAIN", "logfmt", "LOGFMT"} {
		t.Run(format, func(t *testing.T) {
			t.Setenv("DEVSY_LOG_LEVEL", "info")
			t.Setenv("DEVSY_LOG_FORMAT", format)
			if err := initLogger(); err != nil {
				t.Fatal(err)
			}
			var output bytes.Buffer
			remove := log.AddSink(&output)
			defer remove()
			log.Info("format canary")
			if err := log.Sync(); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(output.String(), "format canary") {
				t.Fatalf("missing message in %q", output.String())
			}
			wantJSON := format == "" || strings.EqualFold(format, "auto") ||
				strings.EqualFold(format, "json")
			if json.Valid(bytes.TrimSpace(output.Bytes())) != wantJSON {
				t.Fatalf("unexpected format for %s: %q", format, output.String())
			}
		})
	}
}

func TestLoggerRejectsWhitespace(t *testing.T) {
	for _, key := range []string{"DEVSY_LOG_LEVEL", "DEVSY_LOG_FORMAT"} {
		t.Run(key, func(t *testing.T) {
			t.Setenv("DEVSY_LOG_LEVEL", "")
			t.Setenv("DEVSY_LOG_FORMAT", "")
			t.Setenv(key, " info ")
			if err := initLogger(); err == nil {
				t.Fatal("whitespace-padded configuration was accepted")
			}
		})
	}
}

func TestInvalidLoggerConfigStopsCommand(t *testing.T) {
	if os.Getenv("LIMA_TEST_INVALID_CONFIG") == "1" {
		Execute()
		return
	}
	for _, key := range []string{"DEVSY_LOG_LEVEL", "DEVSY_LOG_FORMAT"} {
		t.Run(key, func(t *testing.T) {
			command := invalidLoggerFixture(t)
			command.Env = append(
				os.Environ(),
				"LIMA_TEST_INVALID_CONFIG=1",
				"DEVSY_LOG_LEVEL=",
				"DEVSY_LOG_FORMAT=",
				key+"=secret-invalid-value",
			)
			var stdout, stderr bytes.Buffer
			command.Stdout = &stdout
			command.Stderr = &stderr
			if err := command.Run(); err == nil {
				t.Fatal("invalid configuration succeeded")
			}
			if stdout.Len() != 0 || strings.Contains(stderr.String(), "secret-invalid-value") {
				t.Fatalf("unexpected output: stdout=%q stderr=%q", stdout.String(), stderr.String())
			}
			assertInvalidConfig(t, stderr.Bytes(), key)
		})
	}
}

func assertRedactedDiagnostics(t *testing.T, output []byte) []string {
	t.Helper()
	var levels []string
	for line := range bytes.SplitSeq(bytes.TrimSpace(output), []byte("\n")) {
		var record struct {
			Timestamp string `json:"ts"`
			Level     string `json:"level"`
			Message   string `json:"msg"`
		}
		if err := json.Unmarshal(line, &record); err != nil {
			t.Fatalf("invalid JSON log: %v", err)
		}
		if record.Timestamp == "" || !strings.Contains(record.Message, "***") ||
			strings.Contains(record.Message, "secret-canary-123") {
			t.Fatalf("invalid or unredacted record: %+v", record)
		}
		levels = append(levels, record.Level)
	}
	return levels
}

func assertInvalidConfig(t *testing.T, stderr []byte, key string) {
	t.Helper()
	var record struct {
		Level   string `json:"level"`
		Message string `json:"msg"`
	}
	if err := json.Unmarshal(bytes.TrimSpace(stderr), &record); err != nil {
		t.Fatalf("invalid error record: %v: %q", err, stderr)
	}
	if record.Level != testErrorLevel || !strings.Contains(record.Message, key) {
		t.Fatalf("unexpected error record: %+v", record)
	}
}

func invalidLoggerFixture(t *testing.T) *exec.Cmd {
	t.Helper()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	return &exec.Cmd{
		Path: executable,
		Args: []string{
			executable,
			"-test.run=^TestInvalidLoggerConfigStopsCommand$",
			"--",
			"create",
		},
	}
}
