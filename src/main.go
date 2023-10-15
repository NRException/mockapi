package main

import (
	"fmt"
	"log"
	"os"
	"src/common"
	"src/server"
	"src/settings"
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

func handleListenersFromFile(filePath string) {
	// Init...
	common.LogVerbose("reading settings file...")
	c := make(chan string)

	// Simple sanity checks...
	if len(filePath) == 0 {
		printHelp()
	}
	if !strings.HasSuffix(filePath, ".yaml") {
		printHelp()
	}

	// Attempt to unmarshal our data from our input file
	u, err := settings.UnmarshalSettingsFile(filePath) 
	common.LogVerbose(u.WebListeners)
	if err != nil {
		log.Println(err.Error())
	}

	// Stand up web listeners and listen
	for _, i := range u.WebListeners {
		go server.EstablishListener(i, c)
	}
	for l := range c {
		log.Println(l)
	}
}

func main() {
	fmt.Println("--- Welcome to MockAPI ---")

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
			handleListenersFromFile(filePath)
		}
	}
}
