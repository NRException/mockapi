// Package logging provides functionality for logging information.
//
// TODO: Replace this with a logging package (i.e., zerolog, logrus)
package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"slices"

	"github.com/google/uuid"
)

// Logging levels.
const (
	Info LogMessageType = "INFO"
	Warn LogMessageType = "WARNING"
)

const logFilePermissions = 0o644

// LogMessageType is the log level.
type LogMessageType string

// String returns the string representation of LogMessageType.
func (lmt LogMessageType) String() string {
	return string(lmt)
}

// SetLogFileActive configures a log file to write to.
func SetLogFileActive(filePath string) error {
	filePath = filepath.Clean(filePath)

	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, logFilePermissions)
	if err != nil {
		return fmt.Errorf("error opening log file: %w", err)
	}

	log.SetOutput(io.MultiWriter(os.Stdout, f))

	return nil
}

// Log prints a log message.
func Log(msgType LogMessageType, msg string) {
	log.Printf("[%s] - %s\n", msgType, msg)
}

// LogVerbose prints a log message if verbose is enabled.
func LogVerbose(msgType LogMessageType, msg string) {
	if slices.Contains(os.Args, "-v") {
		log.Printf("[%s] - %s\n", msgType, msg)
	}
}

// LogOnThread prints a log message with associated UUID.
func LogOnThread(threadID uuid.UUID, msgType LogMessageType, msg string) {
	Log(msgType, fmt.Sprintf("[%s] - %s", threadID, msg))
}

// LogVerboseOnThread prints a log message with associated UUID if verbose is
// enabled.
func LogVerboseOnThread(threadID uuid.UUID, msgType LogMessageType, msg string) {
	LogVerbose(msgType, fmt.Sprintf("[%s] - %s", threadID, msg))
}
