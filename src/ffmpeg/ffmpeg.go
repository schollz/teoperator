package ffmpeg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/schollz/logger"
	"github.com/schollz/teoperator/src/models"
	"github.com/schollz/teoperator/src/utils"
	wav "github.com/youpy/go-wav"
)

// IsInstalled checks whether ffmpeg is installed
func IsInstalled() bool {
	cmd := []string{"--help"}
	logger.Debug(cmd)
	_, err := exec.Command("ffmpeg", cmd...).CombinedOutput()
	if err != nil {
		return false
	}
	return true
}

func NumSamples(fname string) (numSamples int64, sampleRate int64, err error) {
	file, err := os.Open(fname)
	if err != nil {
		return
	}
	reader := wav.NewReader(file)
	format, err := reader.Format()
	if err != nil {
		return
	}
	sampleRate = int64(format.SampleRate)

	defer file.Close()

	for {
		samples, err := reader.ReadSamples()
		if err == io.EOF {
			err = nil
			break
		}

		numSamples += int64(len(samples))
	}
	return
}

func Concatenate(fnames []string) (fname2 string, err error) {
	f, err := ioutil.TempFile(".", "concat")
	for _, fname := range fnames {
		f.WriteString(fmt.Sprintf("file '%s'\n", fname))
	}
	f.Close()
	_, fname2 = filepath.Split(fnames[0])
	fname2 = fname2 + "concat.wav"
	cmd := []string{"-y", "-f", "concat", "-i", f.Name(), fname2}
	logger.Debug(cmd)
	out, err := exec.Command("ffmpeg", cmd...).CombinedOutput()
	if err != nil {
		logger.Errorf("ffmpeg: %s", out)
		return
	} else {
		os.Remove(f.Name())
	}

	return
}

func ToMono(fname string) (fname2 string, err error) {
	_, fname2 = filepath.Split(fname)
	// Create safe filenames to make ffmpeg concat happy
	fname2 = strings.ReplaceAll(fname2, " ", "-") + ".mono.wav"
	cmd := []string{"-y", "-i", fname, "-ss", "0", "-to", "12", "-ar", "44100", "-ac", "1", fname2}
	logger.Debug(cmd)
	out, err := exec.Command("ffmpeg", cmd...).CombinedOutput()
	if err != nil {
		logger.Errorf("ffmpeg: %s", out)
		return
	}
	return
}

type Normalization struct {
	InputI            string `json:"input_i"`
	InputTp           string `json:"input_tp"`
	InputLra          string `json:"input_lra"`
	InputThresh       string `json:"input_thresh"`
	OutputI           string `json:"output_i"`
	OutputTp          string `json:"output_tp"`
	OutputLra         string `json:"output_lra"`
	OutputThresh      string `json:"output_thresh"`
	NormalizationType string `json:"normalization_type"`
	TargetOffset      string `json:"target_offset"`
}

// Normalize will perform double pass ebu R128 normalization
// http://peterforgacs.github.io/2018/05/20/Audio-normalization-with-ffmpeg/
func Normalize(fname string, fnameout string) (err error) {
	cmd := []string{"-i", fname, "-af", "loudnorm=I=-23:LRA=7:tp=-2:print_format=json", "-f", "null", "-"}
	logger.Debug(cmd)
	out, err := exec.Command("ffmpeg", cmd...).CombinedOutput()
	if err != nil {
		logger.Errorf("ffmpeg: %s", out)
		return
	}
	logger.Tracef("ffmpeg output: %s", out)
	index := bytes.LastIndex(out, []byte("{"))
	var n Normalization
	err = json.Unmarshal(out[index:], &n)
	if err != nil {
		return
	}

	logger.Debugf("n: %+v", n)
	if strings.Contains(n.InputI, "inf") {
		logger.Debug("returning, because of inf values")
		os.Rename(fname, fnameout)
		return
	}

	cmd = []string{"-i", fname, "-ar", "44100", "-af",
		fmt.Sprintf("loudnorm=I=-23:LRA=7:tp=-2:measured_I=%s:measured_LRA=%s:measured_tp=%s:measured_thresh=%s:offset=-0.47",
			n.InputI,
			n.InputLra,
			n.InputTp,
			n.InputThresh),
		"-y", fnameout}
	logger.Debug(cmd)
	out, err = exec.Command("ffmpeg", cmd...).CombinedOutput()
	if err != nil {
		return
	}
	logger.Tracef("ffmpeg output: %s", out)

	return

}

// SplitOnSilence splits any audio file based on its silence
func SplitOnSilence(fname string, silenceDB int, silenceMinimumSeconds float64, correction float64) (segments []models.AudioSegment, err error) {
	cmd := []string{"-i", fname, "-af",
		fmt.Sprintf("silencedetect=noise=%ddB:d=%2.3f", silenceDB, silenceMinimumSeconds),
		"-f", "null", "-"}
	logger.Debug(cmd)
	out, err := exec.Command("ffmpeg", cmd...).CombinedOutput()
	if err != nil {
		return
	}
	logger.Tracef("ffmpeg output: %s", out)
	// if !strings.Contains(string(out), "silence_end") {
	// 	err = fmt.Errorf("could not find silence")
	// 	logger.Error(err)
	// 	return
	// }

	var segment models.AudioSegment
	segment.Start = 0
	for _, line := range strings.Split(string(out), "\n") {
		// if strings.Contains(line, "silence_start") {
		// 	seconds, err := utils.ConvertToSeconds(utils.GetStringInBetween(line+" ", "silence_start: ", " "))
		// 	if err == nil {
		// 		segment.End = seconds
		// 		segment.Filename = fname
		// 		segment.Duration = segment.End - segment.Start
		// 		segments = append(segments, segment)
		// 	} else {
		// 		logger.Debug(err)
		// 	}
		// } else if strings.Contains(line, "silence_end") {
		// 	seconds, err := utils.ConvertToSeconds(utils.GetStringInBetween(line, "silence_end: ", " "))
		// 	if err == nil {
		// 		segment.Start = seconds
		// 	} else {
		// 		logger.Debug(err)
		// 	}
		if strings.Contains(line, "silence_end") {
			seconds, err := utils.ConvertToSeconds(utils.GetStringInBetween(line, "silence_end: ", " "))
			if err == nil {
				segment.End = seconds + correction
				segment.Filename = fname
				segment.Duration = segment.End - segment.Start
				if segment.Duration > 0.25 {
					segments = append(segments, segment)
				}
				segment.Start = seconds + correction
			} else {
				logger.Debug(err)
			}
		} else if strings.Contains(line, "time=") {
			seconds, err := utils.ConvertToSeconds(utils.GetStringInBetween(line, "time=", " "))
			if err == nil {
				segment.End = seconds
				segment.Duration = segment.End - segment.Start
				segment.Filename = fname
				if segment.Duration < 0.25 && len(segments) > 0 {
					segments[len(segments)-1].End = seconds
					segments[len(segments)-1].Duration = segments[len(segments)-1].End - segments[len(segments)-1].Start
				} else {
					segments = append(segments, segment)
				}
			} else {
				logger.Debug(err)
			}
		}
	}

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

func RemoveSilence(fnameIn, fnameOut string) (err error) {
	cmd := []string{"-i", fnameIn, "-af", "silenceremove=stop_periods=-1:stop_duration=0.1:stop_threshold=-50dB", "-y", fnameOut}
	logger.Debug(cmd)
	out, err := exec.Command("ffmpeg", cmd...).CombinedOutput()
	if err != nil {
		return
	}
	logger.Tracef("ffmpeg output: %s", out)

	return
}
