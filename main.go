package main

import (
	"flag"
	"os"

	log "github.com/schollz/logger"
	"github.com/schollz/teoperator/src/server"
)

func main() {
	var flagDebug, flagServer bool
	var flagPort int
	flag.BoolVar(&flagDebug, "debug", false, "debug mode")
	flag.BoolVar(&flagServer, "serve", false, "make a server")
	flag.IntVar(&flagPort, "port", 8053, "port to use")
	flag.Parse()

	if flagDebug {
		log.SetLevel("debug")
	} else {
		log.SetLevel("info")
	}

	if flagServer {
		err := server.Run(flagPort)
		if err != nil {
			os.Exit(1)
		}
	}
}
