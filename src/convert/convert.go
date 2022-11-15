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

func newName(fname string) (fname2 string) {
	fname2 = strings.TrimSuffix(fname, filepath.Ext(fname)) + "_patch.aif"
	if _, err := os.Stat(fname2); os.IsNotExist(err) {
		// does not exist
		return
	}
	for i := 2; i < 100; i++ {
		fname2 = strings.TrimSuffix(fname, filepath.Ext(fname)) + fmt.Sprintf("_patch%d.aif", i)
		if _, err := os.Stat(fname2); os.IsNotExist(err) {
			// does not exist
			return
		}

	}
	return
}

func ToSynth(fname string, baseFreq float64) (err error) {
	log.Debug(fname)
	finalName := newName(fname)
	synthPatch := op1.NewSynthSamplePatch(baseFreq)
	err = synthPatch.SaveSample(fname, finalName, false)
	if err == nil {
		fmt.Printf("converted %+v -> %s\n", fname, finalName)
	}
	return
}

func ToDrumSplice(fname string, slices int) (err error) {
	finalName := newName(fname)
	fname2, err := ffmpeg.ToMono(fname)
	defer os.Remove(fname2)
	if err != nil {
		return
	}
	op1data := op1.NewDrumPatch()
	if slices == 0 {
		segments, errSplit := ffmpeg.SplitOnSilence(fname2, -22, 0.2, -0.2)
		if errSplit != nil {
			err = errSplit
			return
		}
		for i, seg := range segments {
			if i < len(op1data.End)-2 {
				start := int64(math.Floor(math.Round(seg.Start*100)*441)) * op1.SAMPLECONVERSION
				end := int64(math.Floor(math.Round(seg.End*100)*441)) * op1.SAMPLECONVERSION
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
	} else {
		var totalSamples int64
		totalSamples, _, err = ffmpeg.NumSamples(fname2)
		if err != nil {
			return
		}
		log.Debugf("found %d samples", totalSamples)
		for i := 0; i < slices; i++ {
			op1data.Start[i] = int64(i) * totalSamples / int64(slices) * op1.SAMPLECONVERSION
			op1data.End[i] = int64(i+1) * totalSamples / int64(slices) * op1.SAMPLECONVERSION
		}
	}

	err = op1data.Save(fname2, finalName)
	if err == nil {
		fmt.Printf("converted %+v -> %s\n", fname, finalName)
	}
	return
}

func ToDrum(fnames []string, slices int) (err error) {
	if len(fnames) == 0 {
		err = fmt.Errorf("no files!")
		return
	}
	if len(fnames) == 1 {
		return ToDrumSplice(fnames[0], slices)
	}
	_, finalName := filepath.Split(fnames[0])
	finalName = newName(finalName)
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
		sampleEnd[i], _, err = ffmpeg.NumSamples(fname2)
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
			drumPatch.Start[i] = (sampleEnd[i-1]) * op1.SAMPLECONVERSION
		}
		drumPatch.End[i] = (sampleEnd[i]) * op1.SAMPLECONVERSION
	}

	err = drumPatch.Save(fname2, finalName)
	if err == nil {
		fmt.Printf("converted %+v -> %s\n", fnames, finalName)
	}
	return
}


func ToDrum2(fnames []string, slices int) (finalName string, err error) {
	if len(fnames) == 0 {
		err = fmt.Errorf("no files!")
		return
	}
	if len(fnames) == 1 {
		err = ToDrumSplice(fnames[0], slices)
		return
	}
	_, finalName = filepath.Split(fnames[0])
	finalName = newName(finalName)
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
		sampleEnd[i], _, err = ffmpeg.NumSamples(fname2)
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
			drumPatch.Start[i] = (sampleEnd[i-1]) * op1.SAMPLECONVERSION
		}
		drumPatch.End[i] = (sampleEnd[i]) * op1.SAMPLECONVERSION
	}

	err = drumPatch.Save(fname2, finalName)
	if err == nil {
		fmt.Printf("converted %+v -> %s\n", fnames, finalName)
	}
	return
}