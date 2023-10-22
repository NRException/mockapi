package common

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/google/uuid"
)

type LogMessageType string

func (re LogMessageType) String() string { return string(re) }

const (
	MSGTYPE_INFO LogMessageType = "INFO"
	MSGTYPE_WARN LogMessageType = "WARNING"
)

func SetLogFileActive(filePath string) error {
	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("SetLogFileActive: %w", err)
	}
	defer f.Close()
	wrt := io.MultiWriter(os.Stdout, f) // Copy io streams
	log.SetOutput(wrt)
	return nil
}

func LogVerbose(msg string, msgType LogMessageType) {
	if ArgSliceContains(os.Args, "-v") {
		log.Println("[" + msgType.String() + "] - " + msg)
	}
}

func LogNonVerbose(msg string, msgType LogMessageType) {
	log.Println("[" + msgType.String() + "] - " + msg)
}

func LogVerboseOnThread(uuid uuid.UUID, msgType LogMessageType, msg string) {
	LogVerbose("["+uuid.String()+"] - "+msg, msgType)
}

func LogNonVerboseOnThread(uuid uuid.UUID, msgType LogMessageType, msg string) {
	LogNonVerbose("["+uuid.String()+"] - "+msg, msgType)
}
