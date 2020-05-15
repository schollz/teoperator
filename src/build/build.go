package build

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	log "github.com/schollz/logger"
	"github.com/schollz/teoperator/src/audiosegment"
	"github.com/schollz/teoperator/src/op1"
)

// DrumpatchFromAudio takes an audio file and converts to drum patches
func DrumpatchFromAudio(fnameAudio string) (patches []string, err error) {
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
	var allSegments [][]audiosegment.AudioSegment
	for i, mergeSegment := range mergedSegments {
		var mergedSegmentsSegments, splitMergedSegments []audiosegment.AudioSegment
		mergedSegmentsSegments, err = audiosegment.SplitOnSilence(mergeSegment.Filename, -25, 0.05)
		if err != nil {
			log.Debug(err)
			return
		}

		// generate splitmergeX-merge.png (remove everything else)
		splitMergedSegments, err = audiosegment.Split(mergedSegmentsSegments, path.Join(fname, fmt.Sprintf("splitmerge%d", i)), false)
		if err != nil {
			log.Debug(err)
			return
		}
		for _, segment := range splitMergedSegments {
			log.Debug(segment.Filename)
			os.Remove(segment.Filename)
			os.Remove(segment.Filename + ".png")
		}
		allSegments = append(allSegments, mergedSegmentsSegments)
	}

	// remove unnessecary files
	for _, segment := range splitSegments {
		log.Debug(segment.Filename)
		os.Remove(segment.Filename)
		os.Remove(segment.Filename + ".png")
	}
	for _, segment := range mergedSegments {
		log.Debug(segment)
		os.Remove(segment.Filename + ".png")
	}

	for i := range allSegments {
		op1data := op1.Default()
		for j := range allSegments[i] {
			log.Debug(allSegments[i][j])
			if j < len(op1data.End) {
				op1data.Start[j] = int(allSegments[i][j].Start * 44100 * 4096)
				op1data.End[j] = int(allSegments[i][j].End * 44100 * 4096)
			}
		}
		// write as op1 data
		err = op1.DrumPatch(allSegments[i][0].Filename, path.Join(fname, fmt.Sprintf("%s-%d.aif", fname, i)), op1data)
		if err != nil {
			return
		}
		// remove wav data
		// err = os.Remove(allSegments[i][0].Filename)
		// if err != nil {
		// 	return
		// }
	}
	return
}
