package convert

import (
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strings"

	log "github.com/schollz/logger"
	"github.com/schollz/teoperator/src/ffmpeg"
	"github.com/schollz/teoperator/src/op1"
)

func ToSynth(fname string, baseFreq float64) (err error) {
	finalName := strings.TrimSuffix(fname, filepath.Ext(fname)) + "_patch.aif"
	synthPatch := op1.NewSynthSamplePatch(baseFreq)
	err = synthPatch.SaveSample(fname, finalName, false)
	if err == nil {
		fmt.Printf("converted %+v -> %s\n", fname, finalName)
	}
	return
}

func ToDrumSplice(fname string) (err error) {
	finalName := strings.TrimSuffix(fname, filepath.Ext(fname)) + "_patch.aif"
	fname2, err := ffmpeg.ToMono(fname)
	defer os.Remove(fname2)
	if err != nil {
		return
	}
	segments, err := ffmpeg.SplitOnSilence(fname2, -22, 0.2, -0.2)
	if err != nil {
		return
	}
	op1data := op1.NewDrumPatch()
	for i, seg := range segments {
		if i < len(op1data.End)-2 {
			start := int64(math.Floor(math.Round(seg.Start*100) * 441 * 4058))
			end := int64(math.Floor(math.Round(seg.End*100) * 441 * 4058))
			if start > end {
				continue
			}
			if end > op1data.End[len(op1data.End)-1] {
				continue
			}
			op1data.Start[i] = start
			op1data.End[i] = end
		}
	}

	err = op1data.Save(fname2, finalName)
	if err == nil {
		fmt.Printf("converted %+v -> %s\n", fname, finalName)
	}
	return
}

func ToDrum(fnames []string) (err error) {
	if len(fnames) == 0 {
		err = fmt.Errorf("no files!")
		return
	}
	if len(fnames) == 1 {
		return ToDrumSplice(fnames[0])
	}
	_, finalName := filepath.Split(fnames[0])
	finalName = strings.TrimSuffix(finalName, filepath.Ext(fnames[0])) + "_patch.aif"
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
		if sampleEnd[i] > 44100*12 {
			sampleEnd[i] = 44100 * 12
		}
		sampleEnd[i] = sampleEnd[i]
		log.Debugf("%s end: %d", fname, sampleEnd[i])
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
			// 2147483646/(44100*12)
			drumPatch.Start[i] = (sampleEnd[i-1]) * 4058
		}
		drumPatch.End[i] = (sampleEnd[i]) * 4058
	}

	err = drumPatch.Save(fname2, finalName)
	if err == nil {
		fmt.Printf("converted %+v -> %s\n", fnames, finalName)
	}
	return
}
