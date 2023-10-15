package common

import (
	"strings"

	"github.com/google/uuid"
)

const (
	Message_Type_Information string = "INFO"
	Message_Type_Warning     string = "WARNING"
	Message_Type_Error       string = "ERROR"
)

type GuidFormattedMessageError struct {
	MessageIdentifier uuid.UUID
	Message           string
}

type GuidFormattedMessageInfo struct {
	MessageIdentifier uuid.UUID
	Message           string
}

func SendChannelEventInfo(channel chan string, msg GuidFormattedMessageInfo) {
	channel <- msg.MessageIdentifier.String() + " [" + Message_Type_Information + "] - " + strings.ToLower(msg.Message)
}
func SendChannelEventError(channel chan string, msg GuidFormattedMessageError) {
	channel <- msg.MessageIdentifier.String() + " [" + Message_Type_Error + "] - " + strings.ToLower(msg.Message)
}
func SendChannelEventErrorObj(channel chan string, msg GuidFormattedMessageError, err error) {
	channel <- msg.MessageIdentifier.String() + " [" + Message_Type_Error + "] - " + strings.ToLower(msg.Message) + ":" + strings.ToLower(err.Error())
}