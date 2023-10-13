package server

import (
	"fmt"
	"net"
	ce "src/common"
	se "src/settings"

	"github.com/google/uuid"
)

var globalListenerCount int = 0

func EstablishListener(listenerSettings se.UnmarshalledRootSettingWebListener, l chan string) error {
	globalListenerCount += 1
	guid := uuid.New()

	var portStr string = fmt.Sprintf(":%d", listenerSettings.ListenerPort)

	ce.SendChannelEventInfo(l, ce.GuidFormattedMessageInfo{MessageIdentifier: guid, Message: fmt.Sprintf("establishing listener %s...", listenerSettings.ListenerName)})
	listenerObject, err := net.Listen("tcp4", portStr)

	// If we get any errors on setting up the listener, report them and return.
	if err != nil {
		ce.SendChannelEventError(l, ce.GuidFormattedMessageError{MessageIdentifier: guid, Message: err.Error()})
		ce.SendChannelEventError(l, ce.GuidFormattedMessageError{MessageIdentifier: guid, Message: "exiting..."})
		return err
	} else {
		ce.SendChannelEventInfo(l, ce.GuidFormattedMessageInfo{MessageIdentifier: guid, Message: "connections ready to be accepted..."})
	}

	// Accept some net connections. Return content
	exitAcceptCondition := false
	for !exitAcceptCondition {
		connection, err := listenerObject.Accept()
		if err != nil {
			return err
		}
		if connection != nil {
			ce.SendChannelEventInfo(l, ce.GuidFormattedMessageInfo{MessageIdentifier: guid, Message: fmt.Sprintf("accepting connection... remote address: %s", connection.RemoteAddr().String())})

			var responseBuffer []byte
			var response = `
			HTTP/1.1 200 OK
			Date: Wed, 13 Oct 2009 00:02: BST
			Server: Apache/2.2.14 (Win32)
			Last-Modified: Wed, 13 Oct 2009 00:02: BST
			Content-Length: 0
			Content-Type: text/plain
			Connection: Closed
			`

			responseBuffer = []byte(response)

			n, _ := connection.Write(responseBuffer)

			connection.Close()
			ce.SendChannelEventInfo(l, ce.GuidFormattedMessageInfo{MessageIdentifier: guid, Message: fmt.Sprintf("sent content data... response: %d", n)})
		}
	}

	return nil // no error, we're happy :)
}
