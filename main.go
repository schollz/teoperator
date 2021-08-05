package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	log "github.com/schollz/logger"
	"github.com/schollz/teoperator/src/convert"
	"github.com/schollz/teoperator/src/download"
	"github.com/schollz/teoperator/src/ffmpeg"
	"github.com/schollz/teoperator/src/server"
	cli "github.com/urfave/cli/v2"
)

func main() {
	log.SetLevel("info")
	if !ffmpeg.IsInstalled() {
		fmt.Println("ffmpeg not installed")
		fmt.Println("you can install it here: https://www.ffmpeg.org/download.html")
		os.Exit(1)
	}

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
		Name:      "teoperator",
		Usage:     "create patches for the op-1 or op-z",
		UsageText: drumUsage + synthUsage,
	}
	app.UseShortOptionHandling = true
	app.Flags = []cli.Flag{
		&cli.BoolFlag{Name: "debug"},
	}
	app.Action = func(c *cli.Context) error {
		download.Duct = ""
		download.ServerName = "http://localhost:8053"
		go func() {
			time.Sleep(100 * time.Millisecond)
			openbrowser("http://localhost:8053")
		}()
		return server.Run(8053, download.ServerName)
	}
	app.Commands = []*cli.Command{
		{
			Name:      "drum",
			Usage:     "create drum patch from file(s)",
			UsageText: drumUsage,
			Flags: []cli.Flag{
				&cli.IntFlag{Name: "slices", Usage: "number of slices", Value: 0},
			},
			Action: func(c *cli.Context) error {
				if c.Bool("debug") {
					log.SetLevel("debug")
				}

				fnames := make([]string, c.Args().Len())
				for i, _ := range fnames {
					fnames[i] = c.Args().Get(i)
				}
				if len(fnames) == 0 {
					return fmt.Errorf("need to specify filename")
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
				if c.Bool("debug") {
					log.SetLevel("debug")
				}
				fnames := make([]string, c.Args().Len())
				for i, _ := range fnames {
					fnames[i] = c.Args().Get(i)
				}
				if len(fnames) == 0 {
					return fmt.Errorf("need to specify filename")
				}
				return convert.ToSynth(fnames[0], c.Float64("freq"))
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
				&cli.BoolFlag{Name: "worker", Usage: "initiate a worker for the server"},
			},
			Action: func(c *cli.Context) error {
				if c.Bool("debug") {
					log.SetLevel("debug")
				}

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

func openbrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		fmt.Println(err)
	}

}
