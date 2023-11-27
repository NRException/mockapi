// Package server provides functionality for setting up HTTP servers.
package server

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"

	"github.com/nrexception/mockapi/internal/logging"
	"github.com/nrexception/mockapi/internal/settings"
)

//nolint:gochecknoglobals // TODO: we should not use global variables
var (
	mux       = http.NewServeMux()
	threadIDs = make([]uuid.UUID, 0)
)

// Listener commands.
const (
	CommandClose ValidListenerCommand = "close"
	CommandPause ValidListenerCommand = "pause"
)

// ValidListenerCommand is a command to trigger an action for Listener objects.
type ValidListenerCommand string

// ListenerCommandPacket is a packet to encapsulate a ValidListenerCommand.
type ListenerCommandPacket struct {
	Identifier uuid.UUID
	Command    ValidListenerCommand
}

// ListenerResponse is the response from a Listener when receiving a
// ValidListenerCommand.
type ListenerResponse string

// String returns the string representation of ListenerResponse.
func (lr ListenerResponse) String() string {
	return string(lr)
}

// EstablishListener establishes a listener for HTTP requests.
func EstablishListener(
	commandChannel chan ListenerCommandPacket,
	listener settings.UnmarshalledRootSettingWebListener,
) {
	threadID := uuid.New()

	createListener(commandChannel, listener, mux, threadID)
}

// ClearAllListeners resets all Listeners.
func ClearAllListeners(commandChannel chan ListenerCommandPacket) {
	logging.LogVerbose(logging.Info, "Closing all listener threads...")

	for _, threadID := range threadIDs {
		logging.LogVerbose(logging.Info, fmt.Sprintf("Closing listener thread %s...", threadID))

		commandChannel <- ListenerCommandPacket{Identifier: threadID, Command: CommandClose}
	}

	logging.LogVerbose(logging.Info, "De-registering listeners...")

	threadIDs = make([]uuid.UUID, 0)

	logging.LogVerbose(logging.Info, "Refreshing server mux...")

	*mux = *http.NewServeMux()
}

func createListener(
	commandChannel chan ListenerCommandPacket,
	webListenerSettings settings.UnmarshalledRootSettingWebListener,
	mux *http.ServeMux,
	threadID uuid.UUID,
) {
	logging.LogVerboseOnThread(
		threadID,
		logging.Info,
		fmt.Sprintf("configuring %d content bindings for %s",
			len(webListenerSettings.ContentBindings),
			webListenerSettings.ListenerName,
		),
	)

	for _, binding := range webListenerSettings.ContentBindings {
		createListenerBinding(binding, mux)
	}

	threadIDs = append(threadIDs, threadID) // Append our thread uuid for later reference if we need to close it...

	if webListenerSettings.EnableTLS {
		go func() {
			logging.LogOnThread(threadID, logging.Info, "starting tls listener...")

			addr := fmt.Sprintf("0.0.0.0:%d", webListenerSettings.ListenerPort)

			// TODO: we don't need to create a new HTTP server, just update mux
			//nolint:gosec // TODO: we should replace this with a *http.Server instance
			err := http.ListenAndServeTLS(
				addr,
				webListenerSettings.CertDetails.CertFile,
				webListenerSettings.CertDetails.KeyFile,
				mux,
			)
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				logging.LogOnThread(threadID, logging.Warn, fmt.Sprintf("listener error: %s", err))
			}
		}()
	} else {
		go func() {
			logging.LogOnThread(threadID, logging.Info, "starting listener...")

			addr := fmt.Sprintf("0.0.0.0:%d", webListenerSettings.ListenerPort)

			// TODO: we don't need to create a new HTTP server, just update mux
			//nolint:gosec // TODO: we should replace this with a *http.Server instance
			err := http.ListenAndServe(addr, mux)
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				logging.LogOnThread(threadID, logging.Warn, fmt.Sprintf("listener error: %s", err))
			}
		}()
	}

	// Hang the go routine unless we close it...
	for c := range commandChannel {
		if threadID == c.Identifier && c.Command == CommandClose {
			return
		}
	}
}

func createListenerBinding(
	binding *settings.ResponseBinding,
	mux *http.ServeMux,
) {
	logging.LogVerbose(logging.Info, fmt.Sprintf("creating binding for %s", binding.Path))

	mux.HandleFunc(binding.Path, func(w http.ResponseWriter, r *http.Request) {
		logging.Log(logging.Info, fmt.Sprintf("binding %s got request from %s to %s",
			binding.Path,
			r.RemoteAddr,
			r.RequestURI,
		))

		// Add headers to response and write, along with response body
		for _, h := range binding.ResponseHeaders {
			w.Header().Add(h.Key, h.Value)
		}

		w.WriteHeader(binding.ResponseCode)

		// Handle and return body type
		switch binding.ResponseBodyType {
		case settings.Inline:
			_, err := io.WriteString(w, binding.ResponseBody)
			if err != nil {
				return
			}
		case settings.File:
			lc, err := getListenerContent(binding)
			if err != nil {
				return
			}

			_, err = io.WriteString(w, lc)
			if err != nil {
				return
			}
		case settings.Proxy:
			// TODO: Add client proxy functionality
		default:
			return
		}
	})
}

func getListenerContent(binding *settings.ResponseBinding) (string, error) {
	switch binding.ResponseBodyType {
	case settings.Inline:
		return binding.ResponseBody, nil
	case settings.File:
		c, err := readFileContent(binding.ResponseBody)
		if err != nil {
			return "", fmt.Errorf("getListenerContent: %w", err)
		}

		return c, nil
	case settings.Proxy:
	}

	return "", fmt.Errorf("getListenerContent(): response type does not match known type of inline, file or proxy")
}

func readFileContent(filePath string) (string, error) {
	b, err := os.ReadFile(filepath.Clean(filePath))
	if err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}

	if len(b) == 0 {
		return "", fmt.Errorf("file has no content")
	}

	return string(b), nil
}
