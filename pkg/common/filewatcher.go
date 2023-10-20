package common

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"time"
)

type FileChangedEvent struct {
	FileName             string
	FileHashBeforeChange string
	FileHashAfterChange  string
}

func generateHash(filePath string) (hash string, err error) {
	b, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("generateHash(): %w", err)
	}
	if len(b) <= 0 {
		return "", errors.New("generateHash(): file contains 0 bytes %s")
	}

	h := md5.New()
	_, err = h.Write(b)
	if err != nil {
		return "", fmt.Errorf("generateHash(): %w", err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

func WatchFile(filePath string, eventChannel chan FileChangedEvent, quitOnDetect bool) (err error) {
	LogVerbose(fmt.Sprintf("Watching file \"%s\"...", filePath), MSGTYPE_INFO)
	defer func() { LogVerbose(fmt.Sprintf("Closing file watcher for \"%s\"...", filePath), MSGTYPE_INFO) }()

	stat, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("WatchFile(): %w", err)
	}

	if stat.Size() <= 0 {
		return errors.New("WatchFile(): file is 0 bytes")
	}

	sleepInterval := time.Second

	lastFileHash, err := generateHash(filePath)
	if err != nil {
		return fmt.Errorf("WatchFile(): %w", err)
	}

	for {
		currentFileHash, err := generateHash(filePath)
		if err != nil {
			return fmt.Errorf("WatchFile(): %w", err)
		}

		if currentFileHash != lastFileHash {
			eventChannel <- FileChangedEvent{FileName: filePath, FileHashBeforeChange: lastFileHash, FileHashAfterChange: currentFileHash}
			lastFileHash = currentFileHash

			if quitOnDetect == true {
				break
			}
		}

		time.Sleep(sleepInterval)
	}

	return nil
}
