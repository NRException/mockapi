package server

import "github.com/google/uuid"

type ValidListenerCommand string
const (
	VLC_Close ValidListenerCommand = "close"
	VLC_Pause ValidListenerCommand = "pause"
)

type ListenerCommandPacket struct {
	Identifier uuid.UUID
	Command ValidListenerCommand
}

type ListenerResponse string
func String(re ListenerResponse) string {return string(re)}