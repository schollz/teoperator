package aubio

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/schollz/logger"
	"github.com/schollz/teoperator/src/ffmpeg"
	"github.com/schollz/teoperator/src/models"
)

// SplitOnSilence splits any audio file based on its silence
func SplitOnSilence(fname string, silenceDB int, silenceMinimumSeconds float64, correction float64) (segments []models.AudioSegment, err error) {
	cmd := []string{"-s", "-30", fname}
	logger.Debug(cmd)
	out, err := exec.Command("aubioonset", cmd...).CombinedOutput()
	if err != nil {
		return
	}
	logger.Debugf("aubio output: %s", out)

	numSamples, sampleRate, err := ffmpeg.NumSamples(fname)
	if err != nil {
		return
	}
	duration := float64(numSamples) / float64(sampleRate)

	var segment models.AudioSegment
	segment.Start = 0
	segment.Filename = fname
	for _, line := range strings.Split(string(out), "\n") {
		seconds, err := strconv.ParseFloat(strings.TrimSpace(line), 64)
		if err != nil {
			logger.Error(line, err)
			continue
		}
		if seconds == 0 {
			continue
		}
		segment.End = seconds
		segment.Duration = segment.End - segment.Start
		segments = append(segments, segment)
		segment.Start = segment.End
	}
	if segment.Start < duration {
		segment.End = duration
		segment.Duration = segment.End - segment.Start
		segments = append(segments, segment)
	}
	logger.Debugf("segments: %+v", segments)

	newSegments := make([]models.AudioSegment, len(segments))
	i := 0
	for _, segment := range segments {
		if segment.Duration > 0.1 {
			newSegments[i] = segment
			i++
		}
	}
	if i == 0 {
		err = fmt.Errorf("could not find any segments")
		return
	}
	newSegments = newSegments[:i]
	return newSegments, nil
}
