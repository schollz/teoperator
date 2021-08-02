package convert

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	log "github.com/schollz/logger"
	"github.com/schollz/teoperator/src/ffmpeg"
	"github.com/schollz/teoperator/src/op1"
)

func ToSynth(fname string, baseFreq float64) (err error) {
	finalName := strings.TrimSuffix(fname, filepath.Ext(fname)) + "_patch.aif"
	synthPatch := op1.NewSynthPatch()
	synthPatch.BaseFreq = baseFreq
	err = synthPatch.SaveSample(fname, finalName, false)
	if err == nil {
		fmt.Printf("converted %+v -> %s\n", fname, finalName)
	}
	return
}

func ToDrum(fnames []string) (err error) {
	finalName := strings.TrimSuffix(fnames[0], filepath.Ext(fnames[0])) + "_patch.aif"
	log.Debugf("converting %+v", fnames)
	f, err := ioutil.TempFile(".", "concat")
	defer os.Remove(f.Name())
	sampleEnd := make([]int64, len(fnames))
	fnames2 := make([]string, len(fnames))
	for i, fname := range fnames {
		var fname2 string
		fname2, err = ffmpeg.ToMono(fname)
		defer os.Remove(fname2)
		if err != nil {
			return
		}
		_, fnames2[i] = filepath.Split(fname2)
		sampleEnd[i], err = ffmpeg.NumSamples(fname2)
		if err != nil {
			return
		}
		if i > 0 {
			sampleEnd[i] = sampleEnd[i] + sampleEnd[i-1]
		}
	}
	f.Close()

	log.Debug(fnames)
	fname2, err := ffmpeg.Concatenate(fnames2)
	defer os.Remove(fname2)
	if err != nil {
		return
	}

	drumPatch := op1.NewDrumPatch()
	for i, _ := range drumPatch.Start {
		if i == len(sampleEnd) {
			break
		}
		if i == 0 {
			drumPatch.Start[i] = 0
		} else {
			drumPatch.Start[i] = sampleEnd[i-1] * 4096
		}
		drumPatch.End[i] = sampleEnd[i] * 4096
	}

	err = drumPatch.Save(fname2, finalName)
	if err == nil {
		fmt.Printf("converted %+v -> %s\n", fnames, finalName)
	}
	return
}
