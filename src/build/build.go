package build

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	log "github.com/schollz/logger"
	"github.com/schollz/teoperator/src/audiosegment"
	"github.com/schollz/teoperator/src/op1"
	"github.com/schollz/teoperator/src/utils"
)

// DrumpatchFromAudio takes an audio file and converts to drum patches
func DrumpatchFromAudio(fnameAudio string, startStop []float64) (fnameShort string, numSegments int, err error) {
	var folder string
	folder, fnameShort = filepath.Split(fnameAudio)
	if len(fnameShort) > 6 {
		fnameShort = fnameShort[:6]
	}
	fnameShort = strings.ToLower(fnameShort)

	// first truncate the audio to only the first minute, while transforming it to a 44100 wav
	truncatedWav := path.Join(folder, "truncated.wav")

	// make it at least 11.75
	if startStop[1]-startStop[0] < 11.75 {
		startStop[1] = 11.75
	}
	err = audiosegment.Truncate(fnameAudio, truncatedWav, utils.SecondsToString(startStop[0]), utils.SecondsToString(startStop[1]))
	if err != nil {
		log.Debug(err)
		return
	}

	// generate splits based on silence
	log.Debug("splitting on silence")
	segments, err := audiosegment.SplitOnSilence(truncatedWav, -18, 0.2)
	if err != nil {
		log.Debug(err)
		return
	}
	log.Debugf("segments: %+v", segments)

	// create the splits
	splitSegments, err := audiosegment.Split(segments, path.Join(folder, fnameShort+"-split"), false)
	if err != nil {
		log.Debug(err)
		return
	}

	mergedSegments, err := audiosegment.Merge(splitSegments, path.Join(folder, fnameShort+"-merge"), 11.5)
	if err != nil {
		log.Debug(err)
		return
	}

	// split the new merged segments on silence
	var allSegments [][]audiosegment.AudioSegment
	for i, mergeSegment := range mergedSegments {
		var mergedSegmentsSegments, splitMergedSegments []audiosegment.AudioSegment
		mergedSegmentsSegments, err = audiosegment.SplitOnSilence(mergeSegment.Filename, -18, 0.2)
		if err != nil {
			log.Debug(err)
			continue
		}

		// generate splitmergeX-merge.png (remove everything else)
		splitMergedSegments, err = audiosegment.Split(mergedSegmentsSegments, path.Join(folder, fmt.Sprintf("%s-%d", fnameShort, i)), false)
		if err != nil {
			log.Debug(err)
			continue
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
		err = op1.DrumPatch(allSegments[i][0].Filename, path.Join(folder, fmt.Sprintf("%s-%d.aif", fnameShort, i)), op1data)
		if err != nil {
			return
		}

		err = audiosegment.Convert(allSegments[i][0].Filename, path.Join(folder, fmt.Sprintf("%s-%d.mp3", fnameShort, i)))
		if err != nil {
			return
		}

		// remove wav data
		err = os.Remove(allSegments[i][0].Filename)
		if err != nil {
			return
		}
	}
	numSegments = len(allSegments)

	err = audiosegment.Convert(path.Join(folder, "truncated.wav"), path.Join(folder, fmt.Sprintf("%s.mp3", fnameShort)))
	if err != nil {
		return
	}
	err = os.Remove(path.Join(folder, "truncated.wav"))

	return
}
