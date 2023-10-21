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

func printHelp() {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.Debug)
	fmt.Println("Simple usage example: ./mockapi -f config.yaml")
	fmt.Fprintln(w, "Command \t Purpose \t Example")
	fmt.Fprintln(w, "-f\t Configuration input file location \t ./mockapi -f <filepath>")
	fmt.Fprintln(w, "-v\t Verbose logging flag \t ./mockapi -f <filepath> -v")
	fmt.Fprintln(w, "-l\t Log file path, mutually exclusive with -v. Discard -v if you're using this. \t ./mockapi -f <filepath> -l dir/helloworld.log")
	fmt.Fprintln(w, "-w\t Watch config file(s) provided by -f, re-apply their configuration if they are changed \t ./mockapi -f <filepath> -w")
	w.Flush()
	os.Exit(0)
}

func handleConfigFileRefresh(fileEventChannel chan co.FileChangedEvent, listenerCommandChannel chan ser.ListenerCommandPacket, listenerResponseChannel chan ser.ListenerResponse, filePath string) {
	for l := range fileEventChannel {
		co.LogVerbose(fmt.Sprintf("Config file \"%s\" was changed. Was: %s is: %s", l.FileName, l.FileHashBeforeChange, l.FileHashAfterChange), co.MSGTYPE_WARN)
		ser.ClearAllListeners(listenerCommandChannel)
		handleListenersFromFile(listenerCommandChannel, listenerResponseChannel, filePath)
	}
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
	for _, i := range u.WebListeners {
		go ser.EstablishListener(listenerCommandChannel, listenerResponseChannel, i)
	}

	return nil
}

var banner string = `

                      ███╗   ███╗ ██████╗  ██████╗██╗  ██╗ █████╗ ██████╗ ██╗                      
                      ████╗ ████║██╔═══██╗██╔════╝██║ ██╔╝██╔══██╗██╔══██╗██║                      
█████╗█████╗█████╗    ██╔████╔██║██║   ██║██║     █████╔╝ ███████║██████╔╝██║    █████╗█████╗█████╗
╚════╝╚════╝╚════╝    ██║╚██╔╝██║██║   ██║██║     ██╔═██╗ ██╔══██║██╔═══╝ ██║    ╚════╝╚════╝╚════╝
                      ██║ ╚═╝ ██║╚██████╔╝╚██████╗██║  ██╗██║  ██║██║     ██║                      
                      ╚═╝     ╚═╝ ╚═════╝  ╚═════╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝     ╚═╝                      
                                                                                                                                                                                     
`

func main() {
	fmt.Println(banner)

	// Ensure we have some calling arugments, or something being passed!
	if len(os.Args) <= 0 {
		fmt.Println("Please use the -h or --help switches for help.")
		os.Exit(1)
	}

	// Handle global arguments
	watchConfigFile := co.ArgSliceContains(os.Args, "-w") // Defines if we expect the config file(s) to dynamically update the configuration of the listeners etc.
	if co.ArgSliceContainsInTerms(os.Args, []string{"-h", "-help", "--h", "--help"}) {
		printHelp()
	} // Prints help if we need it :)

	// Handle -f
	m, params := co.ArgSliceSwitchParameters(os.Args, "-f")
	if m {
		fileWatcherChannel := make(chan co.FileChangedEvent)
		listenerCommandChannel := make(chan ser.ListenerCommandPacket)
		listenerResponseChannel := make(chan ser.ListenerResponse)

		// If specified, watch our config file(s), reload them if needed...
		if watchConfigFile {
			go co.WatchFile(params[0], fileWatcherChannel, false)
		}
		go handleConfigFileRefresh(fileWatcherChannel, listenerCommandChannel, listenerResponseChannel, params[0])

		// Takes first member of slice for now... Will change this when adding multiple file support...
		err := handleListenersFromFile(listenerCommandChannel, listenerResponseChannel, params[0])
		if err != nil {
			log.Fatalf("main(): %s", err)
		}

		// And output out listeners channel!
		for listenResponse := range listenerResponseChannel {
			log.Println(listenResponse)
		}
	}
}
