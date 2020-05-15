package transform

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	log "github.com/schollz/logger"
	"github.com/schollz/op1-aiff/src/audiosegment"
)

// AudioToDrumPatch takes an audio file and converts to drum patches
func AudioToDrumPatch(fnameAudio string) (patches []string, err error) {
	_, fname := filepath.Split(fnameAudio)
	extension := filepath.Ext(fname)
	fname = fname[0 : len(fname)-len(extension)]

	err = os.MkdirAll(fname, os.ModePerm)
	if err != nil {
		log.Debug(err)
		return
	}

	// first truncate the audio to only the first minute, while transforming it to a 44100 wav
	truncatedWav := path.Join(fname, "truncated.wav")
	err = audiosegment.Truncate(fnameAudio, truncatedWav, "00:00:00.00", "00:00:59.00")
	if err != nil {
		log.Debug(err)
		return
	}

	// generate splits based on silence
	segments, err := audiosegment.SplitOnSilence(truncatedWav, -18, 0.05)
	if err != nil {
		log.Debug(err)
		return
	}
	log.Debugf("segments: %+v", segments)

	// create the splits
	splitSegments, err := audiosegment.Split(segments, path.Join(fname, "split"), true)
	if err != nil {
		log.Debug(err)
		return
	}

	mergedSegments, err := audiosegment.Merge(splitSegments, path.Join(fname, "merge"), 6)
	if err != nil {
		log.Debug(err)
		return
	}

	// split the new merged segments on silence
	for i, mergeSegment := range mergedSegments {
		mergedSegmentsSegments, err0 := audiosegment.SplitOnSilence(mergeSegment.Filename, -25, 0.05)
		if err0 != nil {
			err = err0
			log.Debug(err)
			return
		}

		splitMergedSegments, err0 := audiosegment.Split(mergedSegmentsSegments, path.Join(fname, fmt.Sprintf("splitmerge%d", i)), false)
		if err0 != nil {
			err = err0
			log.Debug(err)
			return
		}
		log.Debug(splitMergedSegments)
	}

	return
}
