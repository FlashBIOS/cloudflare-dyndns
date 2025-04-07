package cmd

import (
	"bytes"
	"fmt"
	"github.com/TwiN/go-color"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestFatalError(t *testing.T) {
	tests := []struct {
		name       string
		message    interface{}
		shouldExit bool
		expected   string
	}{
		{
			name:       "nil_message",
			message:    nil,
			shouldExit: false,
			expected:   "",
		},
		{
			name:       "string_message",
			message:    "error occurred",
			shouldExit: true,
			expected:   strings.Trim(color.With(color.Red, "Error: error occurred\n"), "\n"),
		},
		{
			name:       "integer_message",
			message:    42,
			shouldExit: true,
			expected:   strings.Trim(color.With(color.Red, "Error: 42\n"), "\n"),
		},
		{
			name:       "error_struct_message",
			message:    struct{ msg string }{"struct error"},
			shouldExit: true,
			expected:   strings.Trim(color.With(color.Red, "Error: {struct error}\n"), "\n"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldExit {
				cmd := exec.Command(os.Args[0], "-test.run=TestHelperProcess")
				cmd.Env = append(os.Environ(), "TEST_FATAL=1", "FATAL_MSG="+fmt.Sprintf("%v", tt.message))
				var stderr bytes.Buffer
				cmd.Stderr = &stderr
				err := cmd.Run()

				if exitError, ok := err.(*exec.ExitError); ok && !exitError.Success() {
					output := strings.Trim(stderr.String(), "\n")
					if !strings.Contains(output, tt.expected) {
						t.Errorf("Expected '%s', got '%s'", tt.expected, output)
					}
				} else {
					t.Errorf("Process did not exit as expected")
				}
				return
			}

			// If shouldExit is false, just call FatalError directly and ensure no exit occurs
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Unexpected panic: %v", r)
				}
			}()
			FatalError(tt.message) // should not exit for `nil_message`
		})
	}
}

// TestHelperProcess is a helper to simulate process exit during FatalError
// It is executed as a sub-process when fatal error termination is expected.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("TEST_FATAL") != "1" {
		return
	}
	msg := os.Getenv("FATAL_MSG")
	FatalError(msg)
}
