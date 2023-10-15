package server

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	co "src/common"
	se "src/settings"

	"github.com/google/uuid"
)

var globalListenerCount int = 0

func getListenerContent(binding se.UnmarshalledRootSettingWebListenerContentBinding) (string, error) {
	switch binding.ResponseBodyType {
	case se.CONST_RESPONSEBODYTYPE_STATIC:
		return binding.ResponseBody, nil
	case se.CONST_RESPONSEBODYTYPE_FILE:
		c, err := readFileContent(binding.ResponseBody)
		if err != nil {return "", fmt.Errorf("getListenerContent: %w", err)}
		return c, nil
	}
	return "", fmt.Errorf("getListenerContent: %w", errors.New("response type does not match known type of static, file or proxy."))
}

func createListenerBinding(binding se.UnmarshalledRootSettingWebListenerContentBinding, sMux *http.ServeMux, threaduuid uuid.UUID, l chan string) {
	co.SendChannelEventInfo(l, co.GuidFormattedMessageInfo{MessageIdentifier: threaduuid, Message: fmt.Sprintf("\tadding web binding %s...", binding.BindingPath)})

	sMux.HandleFunc(binding.BindingPath, func(w http.ResponseWriter, r *http.Request) {
		co.SendChannelEventInfo(l, co.GuidFormattedMessageInfo{MessageIdentifier: threaduuid, Message: fmt.Sprintf("\t binding \"%s\" got valid request from %s on %s. sending response...", binding.BindingPath, r.RemoteAddr, r.RequestURI)})

		// Add headers to response and write, along with response body
		for _, h := range binding.ResponseHeaders {
			w.Header().Add(h.HeaderKey, h.HeaderValue)
		}

		// Write the headers and body out in response
		w.WriteHeader(binding.ResponseCode)
		if(binding.ResponseBodyType == se.CONST_RESPONSEBODYTYPE_STATIC) {
			_, err := io.WriteString(w, binding.ResponseBody)

			if err != nil {
				co.SendChannelEventErrorObj(l, co.GuidFormattedMessageError{MessageIdentifier: threaduuid, Message: "\terror sending response..."}, err)
			}
		} else if(binding.ResponseBodyType == se.CONST_RESPONSEBODYTYPE_FILE) {
			lc, err := getListenerContent(binding)
			if err != nil {
				co.SendChannelEventErrorObj(l, co.GuidFormattedMessageError{MessageIdentifier: threaduuid, Message: "\terror sending response..."}, err)
			}

			_, err = io.WriteString(w, lc)
			if err != nil {
				co.SendChannelEventErrorObj(l, co.GuidFormattedMessageError{MessageIdentifier: threaduuid, Message: "\terror sending response..."}, err)
			}
		}

	})
}

func createListener(webListenerSettings se.UnmarshalledRootSettingWebListener, sMux *http.ServeMux, threaduuid uuid.UUID, l chan string,) error {
	co.SendChannelEventInfo(l, co.GuidFormattedMessageInfo{MessageIdentifier: threaduuid, Message: fmt.Sprintf("configuring %d content bindings for \"%s\"", len(webListenerSettings.ContentBindings), webListenerSettings.ListenerName)})
	
	for _, binding := range webListenerSettings.ContentBindings {
		binding := binding // Solve concurency issues by creating a copy of binding...
		createListenerBinding(binding, sMux, threaduuid, l) // And call our bindings :)
	}

	if webListenerSettings.EnableTLS {
		co.SendChannelEventInfo(l, co.GuidFormattedMessageInfo{MessageIdentifier: threaduuid, Message: "starting tls listener..."})
		err := http.ListenAndServeTLS(fmt.Sprintf("0.0.0.0:%d", webListenerSettings.ListenerPort), webListenerSettings.CertDetails.CertFile, webListenerSettings.CertDetails.KeyFile, sMux)

		if err != nil {
			co.SendChannelEventErrorObj(l, co.GuidFormattedMessageError{MessageIdentifier: threaduuid, Message: "error starting listener"}, err)
			return err
		}
	} else {
		co.SendChannelEventInfo(l, co.GuidFormattedMessageInfo{MessageIdentifier: threaduuid, Message: "starting non-tls listener..."})
		err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", webListenerSettings.ListenerPort), sMux)

		if err != nil {
			co.SendChannelEventErrorObj(l, co.GuidFormattedMessageError{MessageIdentifier: threaduuid, Message: "error starting listener"}, err)
			return err
		}
	}

	return nil
}


func EstablishListener(ls se.UnmarshalledRootSettingWebListener, l chan string) error {
	// Init some values...
	globalListenerCount += 1
	threaduuid := uuid.New()
	sMux := http.NewServeMux()

	// Actually start listening...
	createListener(ls, sMux, threaduuid, l)

	return nil // no error, we're happy :)
}
