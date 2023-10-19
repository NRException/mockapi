package common

import (
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

func LogVerbose(msg string, msgType LogMessageType) {
	for _, verb := range os.Args {
		if verb == "-v" {
			log.Println("[" + msgType.String() + "] - " + msg)
		}
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
