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
	FileName string
	FileHashBeforeChange string
	FileHashAfterChange string
	DateTimeChanged time.Time
}

func generateHash (filePath string) (hash string, err error) {
	stat, err := os.Stat(filePath) 
	if err != nil {return "", fmt.Errorf("generateHash(): %w", err)}
	if stat.Size() <= 0 {return "", errors.New("generateHash(): file is 0 bytes")}

	b, err := os.ReadFile(filePath)
	if err != nil {return "", fmt.Errorf("generateHash(): %w", err)}
	if len(b) <= 0 {return "", errors.New("generateHash(): file contains 0 bytes")}

	h := md5.New()
	_, err = h.Write(b)
	if err != nil {return "", fmt.Errorf("generateHash(): %w", err)}

	return hex.EncodeToString(h.Sum(nil)), nil
}

func WatchFile(filePath string, eventChannel chan FileChangedEvent) (err error) {
	stat, err := os.Stat(filePath) 
	
	if err != nil {return fmt.Errorf("WatchFile(): %w", err)}
	if stat.Size() <= 0 {return errors.New("WatchFile(): file is 0 bytes")}

	finishedWatching := false
	sleepInterval := 2 * 1000

	lastFileHash, err := generateHash(filePath)
	if err != nil {return fmt.Errorf("WatchFile(): %w", err)}

	for finishedWatching != true {
		currentFileHash, err := generateHash(filePath)
		if err != nil {return fmt.Errorf("WatchFile(): %w", err)}

		if currentFileHash != lastFileHash {
			eventChannel <- FileChangedEvent{FileName: filePath, FileHashBeforeChange: lastFileHash, FileHashAfterChange: currentFileHash, DateTimeChanged: time.Now()}
		}

		time.Sleep(time.Duration(sleepInterval))

	}
	return nil
}