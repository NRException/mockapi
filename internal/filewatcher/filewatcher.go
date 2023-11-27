// Package filewatcher provides functionality for observing file changes and
// updating state.
package filewatcher

import (
	"crypto/md5" //nolint:gosec // file hashing does not need to be cryptographically secure
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/nrexception/mockapi/internal/logging"
)

// FileChangedEvent represents when a watched file has changed.
type FileChangedEvent struct {
	FileName             string
	FileHashBeforeChange string
	FileHashAfterChange  string
}

// WatchFile watches a file for changes and sends a FileChangedEvent when a
// change is observed.
func WatchFile(filePath string, eventChannel chan FileChangedEvent, quitOnDetect bool) error {
	logging.LogVerbose(logging.Info, fmt.Sprintf("Watching file %s...", filePath))
	defer logging.LogVerbose(logging.Info, fmt.Sprintf("Closing file watcher for %s...", filePath))

	stat, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("error checking file to watch: %w", err)
	}

	if stat.Size() <= 0 {
		return fmt.Errorf("file to watch is empty")
	}

	sleepInterval := time.Second

	lastFileHash, err := generateHash(filePath)
	if err != nil {
		return fmt.Errorf("error generating hash: %w", err)
	}

	for {
		currentFileHash, err := generateHash(filePath)
		if err != nil {
			return fmt.Errorf("error generating hash: %w", err)
		}

		if currentFileHash != lastFileHash {
			eventChannel <- FileChangedEvent{FileName: filePath, FileHashBeforeChange: lastFileHash, FileHashAfterChange: currentFileHash}

			lastFileHash = currentFileHash

			if quitOnDetect {
				break
			}
		}

		time.Sleep(sleepInterval)
	}

	return nil
}

func generateHash(filePath string) (string, error) {
	b, err := os.ReadFile(filepath.Clean(filePath))
	if err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}

	if len(b) == 0 {
		return "", fmt.Errorf("file to hash is empty")
	}

	h := md5.New() //nolint:gosec // file hashing does not need to be cryptographically secure

	// hash.Hash from md5 will never error
	_, _ = h.Write(b)

	return hex.EncodeToString(h.Sum(nil)), nil
}
