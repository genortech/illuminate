package logger

import (
	"os"

	"github.com/charmbracelet/log"
)

var (
	// Default logger instance
	Default *log.Logger
)

func init() {
	Default = log.New(os.Stderr)
	Default.SetPrefix("illuminate")
	Default.SetLevel(log.InfoLevel)
}

// GetLogger returns the default logger instance
func GetLogger() *log.Logger {
	return Default
}
