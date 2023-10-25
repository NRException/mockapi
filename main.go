// Package main is the entry point of the program.
package main

import (
	"fmt"
	"log"
	"os"
	"slices"
	"strings"
	"text/tabwriter"

	"github.com/nrexception/mockapi/internal/filewatcher"
	"github.com/nrexception/mockapi/internal/logging"
	"github.com/nrexception/mockapi/internal/server"
	"github.com/nrexception/mockapi/internal/settings"
)

//nolint:lll // banner is necessarily a long line
const banner string = `

                      ███╗   ███╗ ██████╗  ██████╗██╗  ██╗ █████╗ ██████╗ ██╗                      
                      ████╗ ████║██╔═══██╗██╔════╝██║ ██╔╝██╔══██╗██╔══██╗██║                      
█████╗█████╗█████╗    ██╔████╔██║██║   ██║██║     █████╔╝ ███████║██████╔╝██║    █████╗█████╗█████╗
╚════╝╚════╝╚════╝    ██║╚██╔╝██║██║   ██║██║     ██╔═██╗ ██╔══██║██╔═══╝ ██║    ╚════╝╚════╝╚════╝
                      ██║ ╚═╝ ██║╚██████╔╝╚██████╗██║  ██╗██║  ██║██║     ██║                      
                      ╚═╝     ╚═╝ ╚═════╝  ╚═════╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝     ╚═╝                      
                                                                                                                                                                                     
`

func printHelp() {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)

	_, _ = fmt.Fprintln(w, "Simple usage example: ./mockapi -v -w -f config.yaml")
	_, _ = fmt.Fprintln(w, "")
	_, _ = fmt.Fprintln(w, "Command\tDescription")
	_, _ = fmt.Fprintln(w, "-f\tConfiguration file location")
	_, _ = fmt.Fprintln(w, "-v\tVerbose logging")
	_, _ = fmt.Fprintln(w, "-w\tWatch config file(s) provided by -f for changes")

	_ = w.Flush()

	os.Exit(0)
}

func handleConfigFileRefresh(
	fileEventChannel chan filewatcher.FileChangedEvent,
	listenerCommandChannel chan server.ListenerCommandPacket,
	filePath string,
) error {
	for l := range fileEventChannel {
		logging.LogVerbose(logging.Info, fmt.Sprintf("Config file %s changed",
			l.FileName,
		))

		server.ClearAllListeners(listenerCommandChannel)

		err := handleListenersFromFile(listenerCommandChannel, filePath)
		if err != nil {
			return fmt.Errorf("error handling listeners from file: %w", err)
		}
	}

	return nil
}

func handleListenersFromFile(listenerCommandChannel chan server.ListenerCommandPacket, filePath string) error {
	logging.LogVerbose(logging.Info, "Reading settings file")

	if filePath == "" {
		printHelp()
	}

	if !strings.HasSuffix(filePath, ".yaml") {
		printHelp()
	}

	// Attempt to unmarshal our data from our input file
	u, err := settings.UnmarshalSettingsFile(filePath)
	if err != nil {
		return fmt.Errorf("handleListenersFromFile: %w", err)
	}

	// Stand up web listeners and listen
	for _, listener := range u.WebListeners {
		listener := listener

		go server.EstablishListener(listenerCommandChannel, listener)
	}

	return nil
}

func run() error {
	fmt.Print(banner)

	// Ensure we have some calling arguments, or something being passed!
	if len(os.Args) <= 1 {
		return fmt.Errorf("please use the -h or --help switches for help")
	}

	// Handle -h
	helpSwitches := []string{"-h", "-help", "--h", "--help"}

	for _, helpSwitch := range helpSwitches {
		if slices.Contains(os.Args, helpSwitch) {
			printHelp()
		}
	}

	// Handle -l log file location
	index := slices.Index(os.Args, "-l")
	if index != -1 && len(os.Args) > index+1 {
		err := logging.SetLogFileActive(os.Args[index+1])
		if err != nil {
			return fmt.Errorf("error setting log file: %w", err)
		} // TODO: Need to defer log file close here
	}

	// Handle -f file inputs
	index = slices.Index(os.Args, "-f")
	if index == -1 || len(os.Args) <= index+1 {
		return fmt.Errorf("settings file not provided")
	}

	configFile := os.Args[index+1]
	fileWatcherChannel := make(chan filewatcher.FileChangedEvent)
	listenerCommandChannel := make(chan server.ListenerCommandPacket)
	listenerResponseChannel := make(chan server.ListenerResponse)

	// If specified, watch our config file(s), reload them if needed...
	if slices.Contains(os.Args, "-w") {
		go func() {
			err := filewatcher.WatchFile(configFile, fileWatcherChannel, false)
			if err != nil {
				log.Printf("error watching file: %s\n", err)
			}
		}()

		go func() {
			err := handleConfigFileRefresh(fileWatcherChannel, listenerCommandChannel, configFile)
			if err != nil {
				log.Printf("error handling config file refresh: %s\n", err)
			}
		}()
	}

	// Takes first member of slice for now... Will change this when adding multiple file support...
	err := handleListenersFromFile(listenerCommandChannel, configFile)
	if err != nil {
		return fmt.Errorf("error handling listeners from file: %w", err)
	}

	// And output out listeners channel!
	// TODO: This does nothing other than block the main goroutine from exiting
	for listenResponse := range listenerResponseChannel {
		log.Println(listenResponse)
	}

	return nil
}

func main() {
	err := run()
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
}
