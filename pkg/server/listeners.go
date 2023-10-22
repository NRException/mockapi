package server

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"

	co "github.com/nrexception/mockapi/pkg/common"
	se "github.com/nrexception/mockapi/pkg/settings"
)

func getListenerContent(binding se.ResponseBinding) (string, error) {
	switch binding.ResponseBodyType {
	case se.Inline:
		return binding.ResponseBody, nil
	case se.File:
		c, err := readFileContent(binding.ResponseBody)
		if err != nil {
			return "", fmt.Errorf("getListenerContent: %w", err)
		}

		return c, nil
	}

	return "", fmt.Errorf("getListenerContent(): response type does not match known type of inline, file or proxy")
}

func createListenerBinding(commandChannel chan ListenerCommandPacket, responseChannel chan ListenerResponse, binding se.ResponseBinding, sMux *http.ServeMux, threaduuid uuid.UUID) error {
	co.LogVerboseOnThread(threaduuid, co.MSGTYPE_INFO, fmt.Sprintf("creating binding for %s", binding.Path))

	sMux.HandleFunc(binding.Path, func(w http.ResponseWriter, r *http.Request) {
		co.LogNonVerboseOnThread(threaduuid, co.MSGTYPE_INFO, fmt.Sprintf("\t binding \"%s\" got valid request from %s on %s. sending response...", binding.Path, r.RemoteAddr, r.RequestURI))

		// Add headers to response and write, along with response body
		for _, h := range binding.ResponseHeaders {
			w.Header().Add(h.Key, h.Value)
		}

		w.WriteHeader(binding.ResponseCode)

		// Handle and return body type
		switch binding.ResponseBodyType {
		case se.Inline:
			_, err := io.WriteString(w, binding.ResponseBody)
			if err != nil {
				return
			}
		case se.File:
			lc, err := getListenerContent(binding)
			if err != nil {
				return
			}
			_, err = io.WriteString(w, lc)
			if err != nil {
				return
			}
		case se.Proxy:
			// TODO: Add client proxy functionality
		default:
			return
		}
	})

	return nil
}

func createListener(commandChannel chan ListenerCommandPacket, responseChannel chan ListenerResponse, webListenerSettings se.UnmarshalledRootSettingWebListener, sMux *http.ServeMux, threaduuid uuid.UUID) error {
	co.LogVerboseOnThread(threaduuid, co.MSGTYPE_INFO, fmt.Sprintf("configuring %d content bindings for \"%s\"", len(webListenerSettings.ContentBindings), webListenerSettings.ListenerName))

	for _, binding := range webListenerSettings.ContentBindings {
		binding := binding                                                                       // Solve concurency issues by creating a copy of binding...
		err := createListenerBinding(commandChannel, responseChannel, binding, sMux, threaduuid) // And call our bindings :)
		if err != nil {
			return fmt.Errorf("createListener: %w", err)
		}
	}

	listenerRegister = append(listenerRegister, threaduuid) // Append our thread uuid for later reference if we need to close it...

	if webListenerSettings.EnableTLS {
		co.LogNonVerboseOnThread(threaduuid, co.MSGTYPE_INFO, "starting tls listener...")
		err := http.ListenAndServeTLS(fmt.Sprintf("0.0.0.0:%d", webListenerSettings.ListenerPort), webListenerSettings.CertDetails.CertFile, webListenerSettings.CertDetails.KeyFile, sMux)
		if err != nil {
			return fmt.Errorf("createListener: %w", err)
		}
	} else {
		co.LogNonVerboseOnThread(threaduuid, co.MSGTYPE_INFO, "starting non-tls listener...")
		go http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", webListenerSettings.ListenerPort), sMux)
		// if err != nil {
		// 	return fmt.Errorf("createListener: %w", err)
		// }
	}

	// Hang the go routine unless we close it...
	for {
		select {
		case c := <-commandChannel:
			if threaduuid == uuid.UUID(c.Identifier) && c.Command == VLC_Close {
				return nil
			}
		default:
			time.Sleep(5 * time.Second)
		}
	}
}

var sMux = http.NewServeMux()
var listenerRegister []uuid.UUID

func ClearAllListeners(commandChannel chan ListenerCommandPacket) {
	co.LogVerbose("Closing all listener threads...", co.MSGTYPE_WARN)

	for _, thread := range listenerRegister {
		co.LogVerbose(fmt.Sprintf("Closing listener thread %s...", thread), co.MSGTYPE_WARN)
		commandChannel <- ListenerCommandPacket{Identifier: uuid.UUID(thread), Command: VLC_Close}
	}

	co.LogVerbose("De-registering listeners...", co.MSGTYPE_WARN)
	listenerRegister = []uuid.UUID{}

	co.LogVerbose("Refreshing server Mux...", co.MSGTYPE_WARN)
	*sMux = *http.NewServeMux()
}

func EstablishListener(commandChannel chan ListenerCommandPacket, responseChannel chan ListenerResponse, ls se.UnmarshalledRootSettingWebListener) error {
	// Init some values...
	threaduuid := uuid.New()

	// Actually start listening...
	err := createListener(commandChannel, responseChannel, ls, sMux, threaduuid)
	if err != nil {
		return fmt.Errorf("EstablishListener: %w", err)
	}

	return nil // no error, we're happy :)
}
