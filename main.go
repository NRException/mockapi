package main

import (
	"fmt"
	"log"
	"os"

	co "github.com/nrexception/mockapi/pkg/common"
	ser "github.com/nrexception/mockapi/pkg/server"
)

const banner string = `

                      ███╗   ███╗ ██████╗  ██████╗██╗  ██╗ █████╗ ██████╗ ██╗                      
                      ████╗ ████║██╔═══██╗██╔════╝██║ ██╔╝██╔══██╗██╔══██╗██║                      
█████╗█████╗█████╗    ██╔████╔██║██║   ██║██║     █████╔╝ ███████║██████╔╝██║    █████╗█████╗█████╗
╚════╝╚════╝╚════╝    ██║╚██╔╝██║██║   ██║██║     ██╔═██╗ ██╔══██║██╔═══╝ ██║    ╚════╝╚════╝╚════╝
                      ██║ ╚═╝ ██║╚██████╔╝╚██████╗██║  ██╗██║  ██║██║     ██║                      
                      ╚═╝     ╚═╝ ╚═════╝  ╚═════╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝     ╚═╝                      
                                                                                                                                                                                     
`

func run() error {
	fmt.Print(banner)

	// Ensure we have some calling arugments, or something being passed!
	if len(os.Args) <= 1 {
		fmt.Println("Please use the -h or --help switches for help")
		os.Exit(1)
	}

	// Handle help
	if co.ArgSliceContainsInTerms(os.Args, []string{"-h", "-help", "--h", "--help"}) {
		printHelp()
		return nil
	}

    // Handle file watcher
	watchConfigFile := co.ArgSliceContains(os.Args, "-w") // Defines if we expect the config file(s) to dynamically update the configuration of the listeners etc.

	// Handle input file
    m, params := co.ArgSliceSwitchParameters(os.Args, "-f")
	if m {
		fileWatcherChannel := make(chan co.FileChangedEvent)
		listenerCommandChannel := make(chan ser.ListenerCommandPacket)
		listenerResponseChannel := make(chan ser.ListenerResponse)

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

		err := handleListenersFromFile(listenerCommandChannel, listenerResponseChannel, params[0])
		if err != nil {
			return fmt.Errorf("error handling listeners from file: %w", err)
		}

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
