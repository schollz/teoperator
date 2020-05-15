package audiosegment

import (
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/schollz/logger"
	"github.com/schollz/teoperator/src/utils"
	"github.com/schollz/teoperator/src/waveform"
)

type AudioSegment struct {
	Filename string
	Start    float64
	End      float64
	Duration float64
}

const SECONDSATEND = 0.05

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
			} else {
				logger.Debug(err)
			}
		} else if strings.Contains(line, "silence_end") {
			seconds, err := utils.ConvertToSeconds(utils.GetStringInBetween(line, "silence_end: ", " "))
			if err == nil {
				segment.Start = seconds
			} else {
				logger.Debug(err)
			}
		} else if strings.Contains(line, "time=") {
			seconds, err := utils.ConvertToSeconds(utils.GetStringInBetween(line, "time=", " "))
			if err == nil {
				segment.End = seconds
				segment.Duration = segment.End - segment.Start
				segment.Filename = fname
				segments = append(segments, segment)
			} else {
				logger.Debug(err)
			}
		}
	}

	newSegments := make([]AudioSegment, len(segments))
	i := 0
	for _, segment := range segments {
		if segment.Duration > 0.1 {
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
func Split(segments []AudioSegment, fnamePrefix string, addsilence bool) (splitSegments []AudioSegment, err error) {
	splitSegments = make([]AudioSegment, len(segments))
	for i := range segments {
		splitSegments[i] = segments[i]
		splitSegments[i].Filename = fmt.Sprintf("%s%d.wav", fnamePrefix, i)
		splitSegments[i].Duration += 0.1
		var out []byte
		cmd := fmt.Sprintf("-y -i %s -acodec copy -ss %2.8f -to %2.8f %s.0.wav", segments[i].Filename, segments[i].Start, segments[i].End, splitSegments[i].Filename)
		if !addsilence {
			cmd = fmt.Sprintf("-y -i %s -acodec copy -ss %2.8f -to %2.8f %s", segments[i].Filename, segments[i].Start, segments[i].End, splitSegments[i].Filename)
		}
		logger.Debug(cmd)
		out, err = exec.Command("ffmpeg", strings.Fields(cmd)...).CombinedOutput()
		if err != nil {
			logger.Errorf("ffmpeg: %s", out)
			return
		}
		if addsilence {
			// -af 'apad=pad_dur=0.1' adds SECONDSATEND milliseconds of silence to the end
			cmd = fmt.Sprintf("-y -i %s.0.wav -af apad=pad_dur=%2.3f %s", splitSegments[i].Filename, SECONDSATEND, splitSegments[i].Filename)
			logger.Debug(cmd)
			out, err = exec.Command("ffmpeg", strings.Fields(cmd)...).CombinedOutput()
			if err != nil {
				logger.Errorf("ffmpeg: %s", out)
				return
			}
			os.Remove(fmt.Sprintf("%s.0.wav", splitSegments[i].Filename))
		}
	}

	// also generate the audio waveform image for each
	colors := []string{"7FFFD4", "F5F5DC"}
	allfnames := make([]string, len(splitSegments))
	for i := range splitSegments {
		allfnames[i] = fmt.Sprintf("%s.png", splitSegments[i].Filename)
		color := colors[int(math.Mod(float64(i), 2))]
		err = waveform.Image(splitSegments[i].Filename, color, splitSegments[i].Duration)
		if err != nil {
			return
		}
	}
	// generate a merged audio waveform image
	cmd := fmt.Sprintf("%s +append %s-merge.png", strings.Join(allfnames, " "), fnamePrefix)
	logger.Debug(cmd)
	cmd0 := "convert"
	if runtime.GOOS == "windows" {
		cmd0 = "imconvert"
	}
	out, err := exec.Command(cmd0, strings.Fields(cmd)...).CombinedOutput()
	if err != nil {
		logger.Errorf("convert: %s", out)
		return
	}

	return
}

// Merge takes audio segments and creates merges of at most `secondsInEachMerge` seconds
func Merge(segments []AudioSegment, fnamePrefix string, secondsInEachMerge float64) (mergedSegments []AudioSegment, err error) {
	fnamesToMerge := []string{}
	currentLength := 0.0
	mergeNum := 0
	for _, segment := range segments {
		if segment.Duration+currentLength > secondsInEachMerge {
			var mergeSegment AudioSegment
			mergeSegment, err = MergeAudioFiles(fnamesToMerge, fmt.Sprintf("%s%d.wav", fnamePrefix, mergeNum))
			if err != nil {
				return
			}
			mergedSegments = append(mergedSegments, mergeSegment)
			currentLength = 0
			fnamesToMerge = []string{}
			mergeNum++
		}
		fnamesToMerge = append(fnamesToMerge, segment.Filename)
		currentLength += segment.Duration
	}
	var mergeSegment AudioSegment
	mergeSegment, err = MergeAudioFiles(fnamesToMerge, fmt.Sprintf("%s%d.wav", fnamePrefix, mergeNum))
	if err != nil {
		return
	}
	mergedSegments = append(mergedSegments, mergeSegment)

	return
}

func MergeAudioFiles(fnames []string, outfname string) (segment AudioSegment, err error) {
	f, err := ioutil.TempFile(os.TempDir(), "merge")
	if err != nil {
		return
	}
	if !strings.HasSuffix(outfname, ".wav") {
		err = fmt.Errorf("must have wav")
		return
	}
	// defer os.Remove(f.Name())

	for _, fname := range fnames {
		fname, err = filepath.Abs(fname)
		if err != nil {
			return
		}
		_, err = f.WriteString(fmt.Sprintf("file '%s'\n", fname))
		if err != nil {
			return
		}
	}
	f.Close()

	cmd := fmt.Sprintf("-y -f concat -safe 0 -i %s -c copy %s", f.Name(), outfname)
	logger.Debug(cmd)
	out, err := exec.Command("ffmpeg", strings.Fields(cmd)...).CombinedOutput()
	logger.Debugf("ffmpeg: %s", out)
	if err != nil {
		err = fmt.Errorf("ffmpeg; %s", err.Error())
		return
	}
	seconds, err := utils.ConvertToSeconds(utils.GetStringInBetween(string(out), "time=", " bitrate"))

	segment.Duration = seconds
	segment.End = seconds
	segment.Filename = outfname

	// create audio waveform
	err = waveform.Image(segment.Filename, "ffffff", segment.Duration)
	return
}

// Truncate will truncate a file, while converting it to 44100
func Truncate(fnameIn, fnameOut, from, to string) (err error) {
	cmd := fmt.Sprintf("-y -i %s -c copy -ss %s -to %s -ar 44100 %s", fnameIn, from, to, fnameOut)
	logger.Debug(cmd)
	out, err := exec.Command("ffmpeg", strings.Fields(cmd)...).CombinedOutput()
	logger.Debugf("ffmpeg: %s", out)
	if err != nil {
		err = fmt.Errorf("ffmpeg; %s", err.Error())
		return
	}
	return
}
