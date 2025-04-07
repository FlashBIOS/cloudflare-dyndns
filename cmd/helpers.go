package cmd

import (
	"fmt"
	"os"

	"github.com/TwiN/go-color"
)

// FatalError checks if the provided message is not nil, prints it as an error, and terminates the application if true.
func FatalError(message interface{}) {
	if message != nil {
		logger.Error().Msg(fmt.Sprintf("%v", message))
		errorMessage := color.With(color.Red, fmt.Sprintf("Error: %v\n", message))
		_, _ = fmt.Fprintf(os.Stderr, "%s", errorMessage)
		os.Exit(1)
	}
}
