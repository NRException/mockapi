package main

import (
	"fmt"
	"log"
	"os"
	"src/server"
	"src/settings"
	"text/tabwriter"
)

func main() {
	fmt.Println("--- Welcome to MockAPI ---")

	if len(os.Args) <= 1 {
		fmt.Println("Please use the -h or --help switches for help.")
	}

	for i, arg := range os.Args {
		if (arg == "-h") || (arg == "-help") || (arg == "--h") || (arg == "--help") {
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.Debug)
			fmt.Fprintln(w, "Command \t Purpose \t Example")
			fmt.Fprintln(w, "-f\t Configuration input file location \t -f helloworld.yaml")
			fmt.Fprintln(w, "-v\t Verbose logging flag \t -v")
			fmt.Fprintln(w, "-l\t Log file path \t -l dir/helloworld.log")
			w.Flush()
			os.Exit(0)
		}

		if arg == "-f" {
			log.Println("reading settings file...")
			filePath := os.Args[i+1]

			u, err := settings.UnmarshalSettingsFile(filePath)

			if err != nil {
				log.Println(err.Error())
			}

			c := make(chan string) // Common channel for all listener server threads.

			// Stand up web listeners
			for _, i := range u.WebListeners {
				go server.EstablishListener(i, c)
			}

			// Channel listen loop
			for i := range c {
				log.Println(i)
			}

			log.Println(u)
		}
	}

}
