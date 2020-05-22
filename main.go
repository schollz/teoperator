package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/schollz/logger"
	"github.com/schollz/teoperator/src/ffmpeg"
	"github.com/schollz/teoperator/src/op1"
	"github.com/schollz/teoperator/src/server"
)

func main() {
	var flagSynth, flagOut string
	var flagDebug, flagServer bool
	var flagPort int
	flag.BoolVar(&flagDebug, "debug", false, "debug mode")
	flag.BoolVar(&flagServer, "serve", false, "make a server")
	flag.IntVar(&flagPort, "port", 8053, "port to use")
	flag.StringVar(&flagSynth, "synth", "", "build synth patch from file")
	flag.StringVar(&flagOut, "out", "", "name of new patch")
	flag.Parse()

	if flagDebug {
		log.SetLevel("debug")
	} else {
		log.SetLevel("info")
	}

	if !ffmpeg.IsInstalled() {
		fmt.Println("ffmpeg not installed")
		fmt.Println("you can install it here: https://www.ffmpeg.org/download.html")
		os.Exit(1)
	}

	var err error
	if flagServer {
		err = server.Run(flagPort)
	} else if flagSynth != "" {
		_, fname := filepath.Split(flagSynth)
		if flagOut == "" {
			flagOut = strings.Split(fname, ".")[0] + ".op1.aif"
		}
		st := time.Now()
		sp := op1.NewSynthPatch()
		err = sp.SaveSample(flagSynth, flagOut, true)
		if err == nil {
			fmt.Printf("converted '%s' to op-1 synth patch '%s' in %s\n", fname, flagOut, time.Since(st))
		}
	} else {
		flag.PrintDefaults()
	}
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
}
