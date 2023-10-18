package server

import (
	"errors"
	"fmt"
	"io"
	co "mockapi/src/common"
	se "mockapi/src/settings"
	"net/http"

	"github.com/google/uuid"
)

var GlobalListenerCount int = 0

func getListenerContent(binding se.UnmarshalledRootSettingWebListenerContentBinding) (string, error) {
	switch binding.ResponseBodyType {
	case se.CONST_RESPONSEBODYTYPE_INLINE:
		return binding.ResponseBody, nil
	case se.CONST_RESPONSEBODYTYPE_FILE:
		c, err := readFileContent(binding.ResponseBody)
		if err != nil {return "", fmt.Errorf("getListenerContent: %w", err)}
		return c, nil
	}
	return "", fmt.Errorf("getListenerContent: %w", errors.New("response type does not match known type of inline, file or proxy."))
}

func createListenerBinding(binding se.UnmarshalledRootSettingWebListenerContentBinding, sMux *http.ServeMux, threaduuid uuid.UUID, l chan string) error {
	co.LogVerboseOnThread(threaduuid, co.MSGTYPE_INFO, fmt.Sprintf("creating binding for %s", binding.BindingPath))
	sMux.HandleFunc(binding.BindingPath, func(w http.ResponseWriter, r *http.Request) {
		co.LogNonVerboseOnThread(threaduuid, co.MSGTYPE_INFO, fmt.Sprintf("\t binding \"%s\" got valid request from %s on %s. sending response...", binding.BindingPath, r.RemoteAddr, r.RequestURI))

		// Add headers to response and write, along with response body
		for _, h := range binding.ResponseHeaders {
			w.Header().Add(h.HeaderKey, h.HeaderValue)
		}

		w.WriteHeader(binding.ResponseCode)

		// Handle and return body type
		switch binding.ResponseBodyType {
		case se.CONST_RESPONSEBODYTYPE_INLINE:
			_, err := io.WriteString(w, binding.ResponseBody)
			if err != nil {return}
		case se.CONST_RESPONSEBODYTYPE_FILE:
			lc, err := getListenerContent(binding)
			if err != nil {return}
			_, err = io.WriteString(w, lc)
			if err != nil {return}
		case se.CONST_RESPONSEBODYTYPE_PROXY:
			// TODO: Add client proxy functionality
		default:
			return
		}
	})
	return nil
}

func createListener(webListenerSettings se.UnmarshalledRootSettingWebListener, sMux *http.ServeMux, threaduuid uuid.UUID, l chan string,) error {
	co.LogVerboseOnThread(threaduuid, co.MSGTYPE_INFO, fmt.Sprintf("configuring %d content bindings for \"%s\"", len(webListenerSettings.ContentBindings), webListenerSettings.ListenerName))
	
	for _, binding := range webListenerSettings.ContentBindings {
		binding := binding // Solve concurency issues by creating a copy of binding...
		err := createListenerBinding(binding, sMux, threaduuid, l) // And call our bindings :)
		if err != nil {return fmt.Errorf("createListener: %w", err)}
	}

	if webListenerSettings.EnableTLS {
		co.LogNonVerboseOnThread(threaduuid, co.MSGTYPE_INFO, "starting tls listener...")
		err := http.ListenAndServeTLS(fmt.Sprintf("0.0.0.0:%d", webListenerSettings.ListenerPort), webListenerSettings.CertDetails.CertFile, webListenerSettings.CertDetails.KeyFile, sMux)
		if err != nil {return fmt.Errorf("createListener: %w", err)}
	} else {
		co.LogNonVerboseOnThread(threaduuid, co.MSGTYPE_INFO, "starting non-tls listener...")
		err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", webListenerSettings.ListenerPort), sMux)
		if err != nil {return fmt.Errorf("createListener: %w", err)}
	}

	return nil
}


func EstablishListener(ls se.UnmarshalledRootSettingWebListener, l chan string) error {
	// Init some values...
	GlobalListenerCount += 1
	threaduuid := uuid.New()
	sMux := http.NewServeMux()

	// Actually start listening...
	err := createListener(ls, sMux, threaduuid, l)
	if err != nil {return fmt.Errorf("EstablishListener: %w", err)}

	return nil // no error, we're happy :)
}
