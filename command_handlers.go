package main

import (
    "fmt"
    "os"
	"log"
	"strings"
	"text/tabwriter"

	co "github.com/nrexception/mockapi/pkg/common"
	ser "github.com/nrexception/mockapi/pkg/server"
	se "github.com/nrexception/mockapi/pkg/settings"
)

func printHelp() {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)

	_, _ = fmt.Fprintln(w, "Simple usage example: ./mockapi -f config.yaml")
	_, _ = fmt.Fprintln(w, "")
	_, _ = fmt.Fprintln(w, "Command\tPurpose\tExample")
	_, _ = fmt.Fprintln(w, "-f\tConfiguration input file location\t./mockapi -f <filepath>")
	_, _ = fmt.Fprintln(w, "-v\tVerbose logging flag\t./mockapi -f <filepath> -v")
	_, _ = fmt.Fprintln(w, "-w\tWatch config file(s) provided by -f, re-apply their configuration if they are changed\t./mockapi -f <filepath> -w")

	_ = w.Flush()

	os.Exit(0)
}

func handleConfigFileRefresh(fileEventChannel chan co.FileChangedEvent, listenerCommandChannel chan ser.ListenerCommandPacket, listenerResponseChannel chan ser.ListenerResponse, filePath string) error {
	for l := range fileEventChannel {
		co.LogVerbose(fmt.Sprintf("Config file \"%s\" was changed. Was: %s is: %s", l.FileName, l.FileHashBeforeChange, l.FileHashAfterChange), co.MSGTYPE_WARN)

		ser.ClearAllListeners(listenerCommandChannel)

		err := handleListenersFromFile(listenerCommandChannel, listenerResponseChannel, filePath)
		if err != nil {
			return fmt.Errorf("error handling listeners from file: %w", err)
		}
	}

	return nil
}

func handleListenersFromFile(listenerCommandChannel chan ser.ListenerCommandPacket, listenerResponseChannel chan ser.ListenerResponse, filePath string) error {
	co.LogVerbose("Reading settings file", co.MSGTYPE_INFO)

	if len(filePath) == 0 {
		printHelp()
	}
	if !strings.HasSuffix(filePath, ".yaml") {
		printHelp()
	}

	// Attempt to unmarshal our data from our input file
	u, err := se.UnmarshalSettingsFile(filePath)
	if err != nil {
		return fmt.Errorf("handleListenersFromFile: %w", err)
	}

	// Stand up web listeners and listen
	for _, listener := range u.WebListeners {
		listener := listener

		go func() {
			err := ser.EstablishListener(listenerCommandChannel, listenerResponseChannel, listener)
			if err != nil {
				log.Printf("error establishing listener: %s", err)
			}
		}()
	}

	return nil
}
