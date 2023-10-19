package main

import (
	"fmt"
	"log"
	co "mockapi/src/common"
	ser "mockapi/src/server"
	se "mockapi/src/settings"
	"os"
	"strings"
	"text/tabwriter"
)

func printHelp() {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.Debug)
	fmt.Fprintln(w, "Command \t Purpose \t Example")
	fmt.Fprintln(w, "-f\t Configuration input file location \t -f helloworld.yaml")
	fmt.Fprintln(w, "-v\t Verbose logging flag \t -v")
	fmt.Fprintln(w, "-l\t Log file path \t -l dir/helloworld.log")
	w.Flush()
	os.Exit(0)
}

func handleListenersFromFile(filePath string) error {
	// Init...
	co.LogVerbose("Reading settings file(s)...", co.MSGTYPE_INFO)
	listenerChannel := make(chan string)

	// Simple sanity checks...
	if len(filePath) == 0 {
		printHelp()
	}
	if !strings.HasSuffix(filePath, ".yaml") {
		printHelp()
	}

	// Attempt to unmarshal our data from our input file
	u, err := se.UnmarshalSettingsFile(filePath) 
	if err != nil {return fmt.Errorf("handleListenersFromFile: %w", err)}

	// Stand up web listeners and listen
	for _, i := range u.WebListeners {
		go ser.EstablishListener(i, listenerChannel)
	}
	for l := range listenerChannel {
		log.Println(l)
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
	if len(os.Args) <= 1 {
		fmt.Println("Please use the -h or --help switches for help.")
	}

	// Command line handling...
	for i, arg := range os.Args {
		if (arg == "-h") || (arg == "-help") || (arg == "--h") || (arg == "--help") {
			printHelp()
		}

		if arg == "-f" {
			filePath := os.Args[i+1]
			err := handleListenersFromFile(filePath)
			if err != nil {log.Fatalf("main: %s", err)}
		}
	}
}
