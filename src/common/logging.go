package common

import (
	"log"
	"os"
)

func LogVerbose(v any) {
	for _, arg := range os.Args {
		if arg == "-v" {
			log.Println(v)
		}
	}
}