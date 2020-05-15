package ffmpeg

import (
	"fmt"
	"math"
	"os/exec"
	"strings"

	"github.com/schollz/logger"
	"github.com/schollz/op1-aiff/src/utils"
)

type AudioSegment struct {
	Filename string
	Start    float64
	End      float64
	Duration float64
}

// SplitOnSilence splits any audio file based on its silence
func SplitOnSilence(fname string, silenceDB int, silenceMinimumSeconds float64) (segments []AudioSegment, err error) {
	out, err := exec.Command("ffmpeg", strings.Fields(fmt.Sprintf("-i %s -af silencedetect=noise=%ddB:d=%2.3f -f null -", fname, silenceDB, silenceMinimumSeconds))...).CombinedOutput()
	if err != nil {
		return
	}
	logger.Debugf("ffmpeg output: %s", out)
	if !strings.Contains(string(out), "silence_end") {
		err = fmt.Errorf("could not find silence")
		return
	}

	var segment AudioSegment
	segment.Start = 0
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, "silence_start") {
			seconds, err := utils.ConvertToSeconds(utils.GetStringInBetween(line+" ", "silence_start: ", " "))
			if err == nil {
				segment.End = seconds
				segment.Filename = fname
				segment.Duration = segment.End - segment.Start
				segments = append(segments, segment)
			}
		} else if strings.Contains(line, "silence_end") {
			seconds, err := utils.ConvertToSeconds(utils.GetStringInBetween(line, "silence_end: ", " "))
			if err == nil {
				segment.Start = seconds
			}
		} else if strings.Contains(line, "time=") {
			seconds, err := utils.ConvertToSeconds(utils.GetStringInBetween(line, "time=", " "))
			if err == nil {
				segment.End = seconds
				segment.Duration = segment.End - segment.Start
				segments = append(segments, segment)
			}
		}
	}

	newSegments := make([]AudioSegment, len(segments))
	i := 0
	for _, segment := range segments {
		if segment.Duration > 0 {
			newSegments[i] = segment
			i++
		}
	}
	if i == 0 {
		err = fmt.Errorf("could not find any segmenets")
		return
	}
	newSegments = newSegments[:i]
	return newSegments, nil
}

// Split will take AudioSegments and split them apart
func Split(segments []AudioSegment, fnamePrefix string) (splitSegments []AudioSegment, err error) {
	splitSegments = make([]AudioSegment, len(segments))
	for i := range segments {
		splitSegments[i] = segments[i]
		splitSegments[i].Filename = fmt.Sprintf("%s%d.wav", fnamePrefix, i)
		var out []byte
		cmd := fmt.Sprintf("-y -i %s -acodec copy -ss %2.8f -to %2.8f %s", segments[i].Filename, segments[i].Start, segments[i].End, splitSegments[i].Filename)
		logger.Debug(cmd)
		out, err = exec.Command("ffmpeg", strings.Fields(cmd)...).CombinedOutput()
		if err != nil {
			logger.Errorf("ffmpeg: %s", out)
			return
		}
	}

	// also generate the audio waveform for each
	colors := []string{"7FFFD4", "F5F5DC"}
	allfnames := make([]string, len(splitSegments))
	for i := range splitSegments {
		allfnames[i] = fmt.Sprintf("%s.png", splitSegments[i].Filename)
		var out []byte
		color := colors[int(math.Mod(float64(i), 2))]
		cmd := fmt.Sprintf("-i %s -o %s.png --background-color ffffff00 --waveform-color %s --amplitude-scale 1 --no-axis-labels --pixels-per-second 100 --height 80 --width %2.0f", splitSegments[i].Filename, splitSegments[i].Filename, color, splitSegments[i].Duration*100)
		logger.Debug(cmd)
		out, err = exec.Command("audiowaveform", strings.Fields(cmd)...).CombinedOutput()
		if err != nil {
			logger.Errorf("audiowaveform: %s", out)
			return
		}
	}
	// generate a merged audio waveform
	cmd := fmt.Sprintf("%s +append %s-merge.png", strings.Join(allfnames, " "), fnamePrefix)
	logger.Debug(cmd)
	out, err := exec.Command("convert", strings.Fields(cmd)...).CombinedOutput()
	if err != nil {
		logger.Errorf("convert: %s", out)
		return
	}

	return
}
