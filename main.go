package main

import (
	"os"

	log "github.com/schollz/logger"
	"github.com/schollz/teoperator/src/convert"
	cli "github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{}
	app.UseShortOptionHandling = true
	app.Commands = []*cli.Command{
		{
			Name:  "drum",
			Usage: "create drum patch from file(s)",
			Flags: []cli.Flag{
				&cli.IntFlag{Name: "slices", Usage: "number of slices", Value: 0},
				&cli.StringSliceFlag{Name: "file", Aliases: []string{"f"}, Usage: "file(s)"},
			},
			Action: func(c *cli.Context) error {
				err := convert.ToDrum(c.StringSlice("file"), c.Int("slices"))
				return err
			},
		},
		{
			Name:  "synth",
			Usage: "create synth patch from file",
			Flags: []cli.Flag{
				&cli.Float64Flag{Name: "freq", Aliases: []string{"s"}, Value: 440, Usage: "base frequency"},
				&cli.StringFlag{Name: "file", Aliases: []string{"f"}, Usage: "file"},
			},
			Action: func(c *cli.Context) error {
				return nil
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
