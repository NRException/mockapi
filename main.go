package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"text/tabwriter"

	co "github.com/nrexception/mockapi/pkg/common"
	ser "github.com/nrexception/mockapi/pkg/server"
	se "github.com/nrexception/mockapi/pkg/settings"
)

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

	_, _ = fmt.Fprintln(w, "Simple usage example: ./mockapi -f config.yaml")
	_, _ = fmt.Fprintln(w, "")
	_, _ = fmt.Fprintln(w, "Command\tPurpose\tExample")
	_, _ = fmt.Fprintln(w, "-f\tConfiguration input file location\t./mockapi -f <filepath>")
	_, _ = fmt.Fprintln(w, "-v\tVerbose logging flag\t./mockapi -f <filepath> -v")
	_, _ = fmt.Fprintln(w, "-l\tLog file path, mutually exclusive with -v. Discard -v if you're using this.\t./mockapi -f <filepath> -l dir/helloworld.log")
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
	// Init...
	co.LogVerbose("Reading settings file", co.MSGTYPE_INFO)

	// Simple sanity checks...
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

func run() error {
	fmt.Print(banner)

	// Ensure we have some calling arugments, or something being passed!
	if len(os.Args) <= 1 {
		fmt.Println("Please use the -h or --help switches for help")
		os.Exit(1)
	}

	if co.ArgSliceContainsInTerms(os.Args, []string{"-h", "-help", "--h", "--help"}) {
		printHelp()
		return nil
	} // Prints help if we need it :)

	// Handle global arguments
	watchConfigFile := co.ArgSliceContains(os.Args, "-w") // Defines if we expect the config file(s) to dynamically update the configuration of the listeners etc.

	// Handle -f
	m, params := co.ArgSliceSwitchParameters(os.Args, "-f")
	if m {
		fileWatcherChannel := make(chan co.FileChangedEvent)
		listenerCommandChannel := make(chan ser.ListenerCommandPacket)
		listenerResponseChannel := make(chan ser.ListenerResponse)

		// If specified, watch our config file(s), reload them if needed...
		if watchConfigFile {
			go func() {
				err := co.WatchFile(params[0], fileWatcherChannel, false)
				if err != nil {
					log.Printf("error watching file: %s\n", err)
				}
			}()
		}

		go func() {
			err := handleConfigFileRefresh(fileWatcherChannel, listenerCommandChannel, listenerResponseChannel, params[0])
			if err != nil {
				log.Printf("error handling config file refresh: %s\n", err)
			}
		}()

		// Takes first member of slice for now... Will change this when adding multiple file support...
		err := handleListenersFromFile(listenerCommandChannel, listenerResponseChannel, params[0])
		if err != nil {
			return fmt.Errorf("error handling listeners from file: %w", err)
		}

		// And output out listeners channel!
		for listenResponse := range listenerResponseChannel {
			log.Println(listenResponse)
		}
	}

	return nil
}

func main() {
	err := run()
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
}
