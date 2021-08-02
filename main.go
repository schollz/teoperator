package main

import (
	"flag"
	"fmt"
	"os"

	log "github.com/schollz/logger"
	"github.com/schollz/teoperator/src/convert"
	"github.com/schollz/teoperator/src/download"
	"github.com/schollz/teoperator/src/ffmpeg"
	"github.com/schollz/teoperator/src/server"
)

func main() {
	var flagSynth, flagOut, flagDuct, flagServerName string
	var flagDebug, flagServer, flagWorker, flagDrum bool
	var flagPort int
	var flagFreq float64
	flag.BoolVar(&flagDebug, "debug", false, "debug mode")
	flag.BoolVar(&flagDrum, "drum", false, "build drum patch")
	flag.BoolVar(&flagServer, "serve", false, "make a server")
	flag.BoolVar(&flagWorker, "work", false, "start a download worker")
	flag.IntVar(&flagPort, "freq", 440, "base frequency when generating synth patch")
	flag.Float64Var(&flagFreq, "port", 8053, "port to use")
	flag.StringVar(&flagSynth, "synth", "", "build synth patch from file")
	flag.StringVar(&flagOut, "out", "", "name of new patch")
	flag.StringVar(&flagDuct, "duct", "", "name of duct")
	flag.StringVar(&flagServerName, "server", "http://localhost:8053", "name of external ip")
	flag.Parse()

	download.Duct = flagDuct
	download.ServerName = flagServerName
	log.SetLevel("error")

	if !ffmpeg.IsInstalled() {
		fmt.Println("ffmpeg not installed")
		fmt.Println("you can install it here: https://www.ffmpeg.org/download.html")
		os.Exit(1)
	}

	if flagDebug {
		log.SetLevel("debug")
	} else {
		log.SetLevel("info")
	}

	var err error
	if flagServer {
		err = server.Run(flagPort, flagServerName)
	} else if flagSynth != "" {
		err = convert.ToSynth(flagSynth, flagFreq)
	} else if flagDrum {
		err = convert.ToDrum(flag.Args())
	} else if flagWorker {
		err = download.Work()
	} else {
		flag.PrintDefaults()
	}
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
}
