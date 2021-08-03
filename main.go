package main

import (
	"os"

	log "github.com/schollz/logger"
	"github.com/schollz/teoperator/src/convert"
	"github.com/schollz/teoperator/src/download"
	"github.com/schollz/teoperator/src/server"
	cli "github.com/urfave/cli/v2"
)

func main() {
	drumUsage := `
create a drum patch from one-shot files (spliced at file endpoints):
	
    teoperator drum kick.mp3 snare.wav hihat1.wav hihat2.wav

create a drum patch from one file (spliced at transients):
	
    teoperator drum fullset.wav

create a drum patch from one file, spliced at even intervals:
	
    teoperator drum --slices 16 fullset.wav
`
	synthUsage := `
create a synth patch from a sample:
	
    teoperator synth trumpet.wav

create a synth patch from a sample with known frequency:
	
    teoperator synth --freq 220 trumpet_a2.wav`
	app := &cli.App{
		Usage:     "create patches for the op-1 or op-z",
		UsageText: drumUsage + synthUsage,
	}
	app.UseShortOptionHandling = true
	app.Commands = []*cli.Command{
		{
			Name:      "drum",
			Usage:     "create drum patch from file(s)",
			UsageText: drumUsage,
			Flags: []cli.Flag{
				&cli.IntFlag{Name: "slices", Usage: "number of slices", Value: 0},
			},
			Action: func(c *cli.Context) error {
				fnames := make([]string, c.Args().Len())
				for i, _ := range fnames {
					fnames[i] = c.Args().Get(i)
				}
				return convert.ToDrum(fnames, c.Int("slices"))
			},
		},
		{
			Name:      "synth",
			Usage:     "create synth patch from file",
			UsageText: synthUsage,
			Flags: []cli.Flag{
				&cli.Float64Flag{Name: "freq", Aliases: []string{"s"}, Value: 440, Usage: "base frequency"},
			},
			Action: func(c *cli.Context) error {
				return convert.ToSynth(c.Args().Get(1), c.Float64("freq"))
			},
		},
		{
			Name:      "server",
			Usage:     "run server interface",
			UsageText: "",
			Flags: []cli.Flag{
				&cli.IntFlag{Name: "port", Value: 8053, Usage: "local port"},
				&cli.StringFlag{Name: "name", Value: "http://localhost:8053", Usage: "name of server"},
				&cli.StringFlag{Name: "duct", Value: "", Usage: "duct name for spanning multiple workers"},
				&cli.BoolVar{Name: "worker", Usage: "initiate a worker for the server"},
			},
			Action: func(c *cli.Context) error {
				download.Duct = c.String("duct")
				download.ServerName = c.String("name")
				if c.Bool("worker") {
					return download.Work()
				} else {
					return server.Run(c.Int("port"), c.String("name"))
				}
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Error(err)
	}
}

// func run() {
// 	var flagSynth, flagOut, flagDuct, flagServerName string
// 	var flagDebug, flagServer, flagWorker, flagDrum bool
// 	var flagPort, flagSlices int
// 	var flagFreq float64
// 	flag.BoolVar(&flagDebug, "debug", false, "debug mode")
// 	flag.BoolVar(&flagDrum, "drum", false, "build drum patch")
// 	flag.BoolVar(&flagServer, "serve", false, "make a server")
// 	flag.BoolVar(&flagWorker, "work", false, "start a download worker")
// 	flag.Float64Var(&flagFreq, "freq", 440, "base frequency when generating synth patch")
// 	flag.IntVar(&flagPort, "port", 8053, "port to use")
// 	flag.IntVar(&flagSlices, "slice", 0, "if making drum patch, define number of slices")
// 	flag.StringVar(&flagSynth, "synth", "", "build synth patch from file")
// 	flag.StringVar(&flagOut, "out", "", "name of new patch")
// 	flag.StringVar(&flagDuct, "duct", "", "name of duct")
// 	flag.StringVar(&flagServerName, "server", "http://localhost:8053", "name of external ip")
// 	flag.Parse()

// 	download.Duct = flagDuct
// 	download.ServerName = flagServerName
// 	log.SetLevel("error")

// 	if !ffmpeg.IsInstalled() {
// 		fmt.Println("ffmpeg not installed")
// 		fmt.Println("you can install it here: https://www.ffmpeg.org/download.html")
// 		os.Exit(1)
// 	}

// 	if flagDebug {
// 		log.SetLevel("debug")
// 	} else {
// 		log.SetLevel("info")
// 	}

// 	var err error
// 	if flagServer {
// 		err = server.Run(flagPort, flagServerName)
// 	} else if flagSynth != "" {
// 		err = convert.ToSynth(flagSynth, flagFreq)
// 	} else if flagDrum {
// 		err = convert.ToDrum(flag.Args())
// 	} else if flagWorker {
// 		err = download.Work()
// 	} else {
// 		flag.PrintDefaults()
// 	}
// 	if err != nil {
// 		log.Error(err)
// 		os.Exit(1)
// 	}
// }
