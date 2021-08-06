package audiosegment

import (
	"fmt"
	"image"
	_ "image/png"
	"math"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/schollz/logger"
	"github.com/schollz/teoperator/src/aubio"
	"github.com/schollz/teoperator/src/ffmpeg"
	"github.com/schollz/teoperator/src/models"
	"github.com/schollz/teoperator/src/op1"
	"github.com/schollz/teoperator/src/utils"
)

var imagemagickconvert = "convert"

func init() {
	if runtime.GOOS == "windows" {
		imagemagickconvert = "imconvert"
	}
}

const SECONDSATEND = 0.1

func SplitEqual(fname string, secondsMax float64, secondsOverlap float64, splices int) (allSegments [][]models.AudioSegment, err error) {
	err = Convert(fname, fname+".wav")
	if err != nil {
		return
	}
	fname = fname + ".wav"

	if splices > 0 {
		secondsOverlap = 0
	}

	cmd := fmt.Sprintf("%s", fname)
	logger.Debug(cmd)
	out, err := exec.Command("ffprobe", strings.Fields(cmd)...).CombinedOutput()
	if err != nil {
		logger.Debugf("%s", out)
		return
	}
	secondsDuration, err := utils.ConvertToSeconds(utils.GetStringInBetween(string(out), "Duration: ", ","))
	if err != nil {
		logger.Debugf("%s", out)
		logger.Debug(err)
		return
	}
	secondsStart, _ := strconv.ParseFloat(utils.GetStringInBetween(string(out), "start: ", ","), 64)

	secondStart := []float64{}
	for i := secondsStart; i < secondsDuration+secondsStart; i += secondsMax - secondsOverlap {
		secondStart = append(secondStart, i)
	}

	numJobs := len(secondStart)
	// step 2: specify the job and result
	type job struct {
		start float64
	}
	type result struct {
		segments []models.AudioSegment
		err      error
	}
	jobs := make(chan job, numJobs)
	results := make(chan result, numJobs)
	runtime.GOMAXPROCS(runtime.NumCPU() * 2)
	for i := 0; i < runtime.NumCPU(); i++ {
		go func(jobs <-chan job, results chan<- result) {
			for j := range jobs {
				// step 3: specify the work for the worker
				var r result
				folder, filenameonly := filepath.Split(fname)
				fnameTrunc := path.Join(folder, fmt.Sprintf("%s%03d.wav", filenameonly[:3], int(j.start)))
				fnameTruncOP1 := path.Join(folder, fmt.Sprintf("%s%03d.aif", filenameonly[:3], int(j.start)))
				r.err = Truncate(fname, fnameTrunc, utils.SecondsToString(j.start), utils.SecondsToString(j.start+secondsMax))
				if r.err != nil {
					logger.Error(r.err)
					results <- r
					continue
				}

				if splices == 0 {
					r.segments, r.err = aubio.SplitOnSilence(fnameTrunc, -22, 0.2, -0.2)
					if r.err != nil {
						r.segments, r.err = ffmpeg.SplitOnSilence(fnameTrunc, -22, 0.2, -0.2)
						if r.err != nil {
							logger.Error(r.err)
							results <- r
							continue
						}
					}
				} else {
					r.segments = make([]models.AudioSegment, splices)
					for i, _ := range r.segments {
						r.segments[i].Start = secondsDuration * float64(i) / float64(splices)
						r.segments[i].End = secondsDuration * float64(i+1) / float64(splices)
						r.segments[i].Duration = r.segments[i].End - r.segments[i].Start
						r.segments[i].Filename = fnameTrunc
					}
				}
				r.err = DrawSegments(r.segments)
				if r.err != nil {
					logger.Error(r.err)
					results <- r
					continue
				}

				// generate op-1 stuff
				op1data := op1.NewDrumPatch()
				for i, seg := range r.segments {
					r.segments[i].StartAbs = j.start
					r.segments[i].EndAbs = j.start + secondsMax
					if i < len(op1data.End)-2 {
						start := int64(math.Floor(math.Round(seg.Start*100) * 441 * 4058))
						end := int64(math.Floor(math.Round(seg.End*100) * 441 * 4058))
						if start > end {
							continue
						}
						if end > op1data.End[len(op1data.End)-1] {
							continue
						}
						logger.Debug(seg.Start, start)
						logger.Debug(seg.End, end)
						op1data.Start[i] = start
						op1data.End[i] = end
					}
				}

				// write as op1 data
				r.err = op1data.Save(fnameTrunc, fnameTruncOP1)
				if r.err != nil {
					logger.Error(r.err)
					continue
				}
				results <- r
			}
		}(jobs, results)
	}

	// step 4: send out jobs
	for i := 0; i < numJobs; i++ {
		jobs <- job{secondStart[i]}
	}
	close(jobs)

	// step 5: do something with results
	for i := 0; i < numJobs; i++ {
		r := <-results
		if r.err != nil {
			logger.Error(err)
			continue
		}
		allSegments = append(allSegments, r.segments)
	}

	return
}

// DrawSegments will take a segment and draw it.
// audiowaveform -i creeley-0.000-12.000.wav -o lifeb.png --background-color ffffff00 --waveform-color 000000 --amplitude-scale 2 --no-axis-labels --pixels-per-second 100 --height 160 --width 1200
// convert -size 600x160 canvas:khaki  canvas_khaki.gif
// convert -size 600x160 canvas:green  canvas_green.gif
// convert canvas_khaki.gif canvas_green.gif +append canvas.gif
// composite lifeb.png canvas.gif -compose Dst_In 3.png
// convert 3.png -fuzz 1% -transparent black 4.png
// eog 4.png
func DrawSegments(segments []models.AudioSegment) (err error) {
	if len(segments) == 0 {
		err = fmt.Errorf("no segments")
		return
	}
	wave := utils.TempFileName("wave", ".png")
	defer os.Remove(wave)
	cmd := fmt.Sprintf("-i %s -o %s --background-color ffffff00 --waveform-color 000000 --amplitude-scale 2 --no-axis-labels --pixels-per-second 100 --height 160 --width %2.0f",
		segments[0].Filename, wave, (segments[len(segments)-1].End-segments[0].Start)*100,
	)
	logger.Debug(cmd)
	out, err := exec.Command("audiowaveform", strings.Fields(cmd)...).CombinedOutput()
	if err != nil {
		logger.Errorf("audiowaveform: %s", out)
	}

	colors := []string{"#EEEEEE", "#343434"}
	canvases := []string{}
	for i := range segments {
		canvasName := utils.TempFileName("canvas", ".png")
		defer os.Remove(canvasName)

		canvases = append(canvases, canvasName)
		cmd = fmt.Sprintf("-size %2.0fx160 canvas:%s %s",
			segments[i].Duration*100, colors[int(math.Mod(float64(i), 2))], canvasName,
		)
		logger.Debug(cmd)
		out, err = exec.Command(imagemagickconvert, strings.Fields(cmd)...).CombinedOutput()
		if err != nil {
			logger.Errorf("audiowaveform: %s", out)
		}
	}

	// merge canvases
	finalCanvas := utils.TempFileName("final", ".png")
	defer os.Remove(finalCanvas)
	cmd = fmt.Sprintf("%s +append %s",
		strings.Join(canvases, " "), finalCanvas,
	)
	logger.Debug(cmd)
	out, err = exec.Command(imagemagickconvert, strings.Fields(cmd)...).CombinedOutput()
	if err != nil {
		logger.Errorf("convert: %s", out)
	}

	// crop final canvas (not sure why this is nessecary)
	width, height := getImageDimension(wave)
	finalCanvasResized := utils.TempFileName("finalresize", ".png")
	defer os.Remove(finalCanvasResized)
	cmd = fmt.Sprintf("%s -crop %dx%d+0+0 %s",
		finalCanvas, width, height, finalCanvasResized,
	)
	logger.Debug(cmd)
	out, err = exec.Command(imagemagickconvert, strings.Fields(cmd)...).CombinedOutput()
	if err != nil {
		logger.Errorf("convert: %s", out)
	}

	composite := utils.TempFileName("composite", ".png")
	defer os.Remove(composite)
	cmd = fmt.Sprintf("%s %s -compose Dst_In %s",
		wave, finalCanvasResized, composite,
	)
	logger.Debug(cmd)
	out, err = exec.Command("composite", strings.Fields(cmd)...).CombinedOutput()
	if err != nil {
		logger.Errorf("composite: %s", out)
	}

	final := segments[0].Filename + ".png"
	cmd = fmt.Sprintf("%s -fuzz 1%% -transparent black %s",
		composite, final,
	)
	logger.Debug(cmd)
	out, err = exec.Command(imagemagickconvert, strings.Fields(cmd)...).CombinedOutput()
	if err != nil {
		logger.Errorf("convert: %s", out)
	}

	return
}

func getImageDimension(imagePath string) (int, int) {
	file, err := os.Open(imagePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}

	image, _, err := image.DecodeConfig(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", imagePath, err)
	}
	return image.Width, image.Height
}

// // Split will take models.AudioSegments and split them apart
// func Split(segments []models.AudioSegment, fnamePrefix string, addsilence bool) (splitSegments []models.AudioSegment, err error) {
// 	splitSegments = make([]models.AudioSegment, len(segments))
// 	for i := range segments {
// 		splitSegments[i] = segments[i]
// 		splitSegments[i].Filename = fmt.Sprintf("%s%d.wav", fnamePrefix, i)
// 		splitSegments[i].Duration += 0.1
// 		var out []byte
// 		cmd := fmt.Sprintf("-y -i %s -acodec copy -ss %2.8f -to %2.8f %s.0.wav", segments[i].Filename, segments[i].Start, segments[i].End, splitSegments[i].Filename)
// 		if !addsilence {
// 			cmd = fmt.Sprintf("-y -i %s -acodec copy -ss %2.8f -to %2.8f %s", segments[i].Filename, segments[i].Start, segments[i].End, splitSegments[i].Filename)
// 		}
// 		logger.Debug(cmd)
// 		out, err = exec.Command("ffmpeg", strings.Fields(cmd)...).CombinedOutput()
// 		if err != nil {
// 			logger.Errorf("ffmpeg: %s", out)
// 			return
// 		}
// 		if addsilence {
// 			// -af 'apad=pad_dur=0.1' adds SECONDSATEND milliseconds of silence to the end
// 			cmd = fmt.Sprintf("-y -i %s.0.wav -af apad=pad_dur=%2.3f %s", splitSegments[i].Filename, 0.011, splitSegments[i].Filename)
// 			logger.Debug(cmd)
// 			out, err = exec.Command("ffmpeg", strings.Fields(cmd)...).CombinedOutput()
// 			if err != nil {
// 				logger.Errorf("ffmpeg: %s", out)
// 				return
// 			}
// 			os.Remove(fmt.Sprintf("%s.0.wav", splitSegments[i].Filename))
// 		}
// 	}

// 	// also generate the audio waveform image for each
// 	colors := []string{"7FFFD4", "F5F5DC"}
// 	allfnames := make([]string, len(splitSegments))
// 	for i := range splitSegments {
// 		allfnames[i] = fmt.Sprintf("%s.png", splitSegments[i].Filename)
// 		color := colors[int(math.Mod(float64(i), 2))]
// 		err = waveform.Image(splitSegments[i].Filename, color, splitSegments[i].Duration)
// 		if err != nil {
// 			return
// 		}
// 	}
// 	// generate a merged audio waveform image
// 	cmd := fmt.Sprintf("%s +append %s-merge.png", strings.Join(allfnames, " "), fnamePrefix)
// 	logger.Debug(cmd)
// 	cmd0 := imagemagickconvert
// 	if runtime.GOOS == "windows" {
// 		cmd0 = "imconvert"
// 	}
// 	out, err := exec.Command(cmd0, strings.Fields(cmd)...).CombinedOutput()
// 	if err != nil {
// 		logger.Errorf("convert: %s", out)
// 		return
// 	}

// 	return
// }

// // Merge takes audio segments and creates merges of at most `secondsInEachMerge` seconds
// func Merge(segments []models.AudioSegment, fnamePrefix string, secondsInEachMerge float64) (mergedSegments []models.AudioSegment, err error) {
// 	fnamesToMerge := []string{}
// 	currentLength := 0.0
// 	mergeNum := 0
// 	for _, segment := range segments {
// 		if segment.Duration+currentLength > secondsInEachMerge {
// 			var mergeSegment models.AudioSegment
// 			mergeSegment, err = MergeAudioFiles(fnamesToMerge, fmt.Sprintf("%s%d.wav", fnamePrefix, mergeNum))
// 			if err != nil {
// 				return
// 			}
// 			mergedSegments = append(mergedSegments, mergeSegment)
// 			currentLength = 0
// 			fnamesToMerge = []string{}
// 			mergeNum++
// 		}
// 		fnamesToMerge = append(fnamesToMerge, segment.Filename)
// 		currentLength += segment.Duration
// 	}
// 	var mergeSegment models.AudioSegment
// 	mergeSegment, err = MergeAudioFiles(fnamesToMerge, fmt.Sprintf("%s%d.wav", fnamePrefix, mergeNum))
// 	if err != nil {
// 		return
// 	}
// 	mergedSegments = append(mergedSegments, mergeSegment)

// 	return
// }

// func MergeAudioFiles(fnames []string, outfname string) (segment models.AudioSegment, err error) {
// 	f, err := ioutil.TempFile(os.TempDir(), "merge")
// 	if err != nil {
// 		return
// 	}
// 	if !strings.HasSuffix(outfname, ".wav") {
// 		err = fmt.Errorf("must have wav")
// 		return
// 	}
// 	defer os.Remove(f.Name())

// 	for _, fname := range fnames {
// 		fname, err = filepath.Abs(fname)
// 		if err != nil {
// 			return
// 		}
// 		_, err = f.WriteString(fmt.Sprintf("file '%s'\n", fname))
// 		if err != nil {
// 			return
// 		}
// 	}
// 	f.Close()

// 	cmd := fmt.Sprintf("-y -f concat -safe 0 -i %s -c copy %s", f.Name(), outfname)
// 	logger.Debug(cmd)
// 	out, err := exec.Command("ffmpeg", strings.Fields(cmd)...).CombinedOutput()
// 	logger.Debugf("ffmpeg: %s", out)
// 	if err != nil {
// 		err = fmt.Errorf("ffmpeg; %s", err.Error())
// 		return
// 	}
// 	seconds, err := utils.ConvertToSeconds(utils.GetStringInBetween(string(out), "time=", " bitrate"))

// 	segment.Duration = seconds
// 	segment.End = seconds
// 	segment.Filename = outfname

// 	// create audio waveform
// 	err = waveform.Image(segment.Filename, "ffffff", segment.Duration)
// 	return
// }

// Truncate will truncate a file, while converting it to 44100
func Truncate(fnameIn, fnameOut, from, to string) (err error) {
	cmd := fmt.Sprintf("-y -i %s -c copy -ss %s -to %s %s", fnameIn, from, to, fnameOut)
	logger.Debug(cmd)
	out, err := exec.Command("ffmpeg", strings.Fields(cmd)...).CombinedOutput()
	logger.Debugf("ffmpeg: %s", out)
	if err != nil {
		err = fmt.Errorf("ffmpeg; %s", err.Error())
		return
	}
	return
}

// Convert will convert a file to
func Convert(fnameIn, fnameOut string) (err error) {
	cmd := fmt.Sprintf("-y -i %s -ar 44100 %s", fnameIn, fnameOut)
	logger.Debug(cmd)
	out, err := exec.Command("ffmpeg", strings.Fields(cmd)...).CombinedOutput()
	logger.Debugf("ffmpeg: %s", out)
	if err != nil {
		err = fmt.Errorf("ffmpeg; %s", err.Error())
		return
	}
	return
}
