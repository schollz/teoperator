package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-audio/aiff"
)

func main() {

	_, err := getsilences()
	if err != nil {
		panic(err)
	}
}

func run2() (err error) {
	f, err := os.Open("2.aif")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	d := aiff.NewDecoder(f)

	a, err := d.FullPCMBuffer()
	if err != nil {
		return
	}
	_ = a
	fmt.Println(len(a.Data))
	fmt.Println(a.Format)

	// algo for finding points
	// a = ao + 2*rand(size(a))*10000;
	// a = a/10000;
	// s = [];
	// for i=1:window:length(a)-window
	// 	s = [s; std(a(i:i+window))];
	// end
	// s = [0; s];
	// sdiffs = diff(abs(s));
	// plot(s)
	// hold on;
	// plot(sdiffs,'-o')
	// stddiffs = std(sdiffs)

	// hold on;
	// ind = find(sdiffs > stddiffs)
	// plot(ind,ones(size(ind)),'o')

	return
}

type AudioSegment struct {
	Start    float64
	End      float64
	Filename string
	Duration float64
}

func getsilences() (positions []string, err error) {
	s := strings.Fields(`ffmpeg -i 1.aif -af silencedetect=noise=-30dB:d=0.1 -f null -`)

	out, err := exec.Command(s[0], s[1:]...).CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(out))

	var segments []AudioSegment
	var segment AudioSegment
	segment.Start = 0
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, "silence_end") {
			seconds, err := ConvertToSeconds(GetStringInBetween(line, "silence_end: ", " "))
			if err == nil {
				segment.End = seconds
				segment.Duration = segment.End - segment.Start
				segments = append(segments, segment)
				segment.Start = segment.End
			}
		} else if strings.Contains(line, "time=") {
			seconds, err := ConvertToSeconds(GetStringInBetween(line, "time=", " "))
			if err == nil {
				segment.End = seconds
				segment.Duration = segment.End - segment.Start
				segments = append(segments, segment)
				segment.Start = segment.End
			}
		}
	}
	fmt.Println(segments)

	fnamePrefix := TempFileName("split", "-")
	for i := range segments {
		segments[i].Filename = fmt.Sprintf("%s%d.wav", fnamePrefix, i)
		_, err = exec.Command("ffmpeg", strings.Fields(fmt.Sprintf("-i 1.aif -acodec copy -ss %2.8f -to %2.8f %s", segments[i].Start, segments[i].End, segments[i].Filename))...).CombinedOutput()
		if err != nil {
			return
		}
	}
	fmt.Println(segments)

	fnamesToMerge := []string{}
	currentLength := 0.0
	mergeNum := 0
	fnamePrefix = TempFileName("merge", "-")
	mergedFiles := []string{}
	for _, segment := range segments {
		if segment.Duration+currentLength > 11.5 {
			err = MergeAudioFiles(fnamesToMerge, fmt.Sprintf("%s%d.wav", fnamePrefix, mergeNum))
			if err != nil {
				return
			}
			mergedFiles = append(mergedFiles, fmt.Sprintf("%s%d.wav", fnamePrefix, mergeNum))
			currentLength = 0
			fnamesToMerge = []string{}
			mergeNum++
		}
		fnamesToMerge = append(fnamesToMerge, segment.Filename)
		currentLength += segment.Duration
	}
	err = MergeAudioFiles(fnamesToMerge, fmt.Sprintf("%s%d.wav", fnamePrefix, mergeNum))
	if err != nil {
		return
	}
	mergedFiles = append(mergedFiles, fmt.Sprintf("%s%d.wav", fnamePrefix, mergeNum))

	fmt.Println(mergedFiles)
	return
}

func MergeAudioFiles(fnames []string, outfname string) (err error) {
	mergelist := TempFileName("mergelist", ".txt")
	defer os.Remove(mergelist)

	f, err := os.Create(mergelist)
	if err != nil {
		return
	}
	for _, fname := range fnames {
		f.WriteString(fmt.Sprintf("file '%s'\n", fname))
	}
	f.Close()

	_, err = exec.Command("ffmpeg", strings.Fields(fmt.Sprintf("-f concat -safe 0 -i %s -c copy %s", mergelist, outfname))...).CombinedOutput()
	return
}

func Duration(fname string) (seconds float64, err error) {
	out, err := exec.Command("ffprobe", strings.Fields(fmt.Sprintf("-v error -show_entries format=duration -of default=noprint_wrappers=1:nokey=1 %s", fname))...).CombinedOutput()
	if err != nil {
		return
	}
	seconds, err = strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	return
}

// ConvertToSeconds converts a string lik 00:00:11.35 into seconds (11.35)
func ConvertToSeconds(s string) (seconds float64, err error) {
	parts := strings.Split(s, ":")
	multipliers := []float64{60 * 60, 60, 1}
	if len(parts) == 2 {
		multipliers = []float64{60, 1, 1}
	} else if len(parts) == 1 {
		multipliers = []float64{1, 1, 1}
	}
	for i, part := range parts {
		var partf float64
		partf, err = strconv.ParseFloat(part, 64)
		if err != nil {
			return
		}
		seconds += partf * multipliers[i]
	}
	return
}

func run() (err error) {
	b, err := ioutil.ReadFile("1.aif")
	if err != nil {
		return
	}
	start := bytes.Index(b, []byte("op-1"))
	if start < 0 {
		return
	}
	end := bytes.Index(b[start:], []byte("}"))
	if end < 0 {
		return
	}

	fmt.Println(string(b[start+4 : start+end+1]))

	var op1data OP1MetaData
	err = json.Unmarshal(b[start+4:start+end+1], &op1data)
	if err != nil {
		return
	}
	fmt.Printf("\n op1data: %+v\n", op1data)
	fmt.Println(len(op1data.End))

	// modify one end point of the first one
	op1data.End[0] = op1data.End[0] / 2

	bop1, err := json.Marshal(op1data)
	if err != nil {
		return
	}
	b2 := append([]byte{}, b[:start+4]...)
	b2 = append(b2, bop1...)
	b2 = append(b2, b[start+end+1:]...)
	err = ioutil.WriteFile("2.aif", b2, 0644)

	return
}

// OP1MetaData is a list of custom fields sometimes set by OP-1
type OP1MetaData struct {
	DrumVersion int    `json:"drum_version"`
	DynaEnv     []int  `json:"dyna_env"`
	End         []int  `json:"end"`
	FxActive    bool   `json:"fx_active"`
	FxParams    []int  `json:"fx_params"`
	FxType      string `json:"fx_type"`
	LfoActive   bool   `json:"lfo_active"`
	LfoParams   []int  `json:"lfo_params"`
	LfoType     string `json:"lfo_type"`
	Name        string `json:"name"`
	Octave      int    `json:"octave"`
	Pitch       []int  `json:"pitch"`
	Playmode    []int  `json:"playmode"`
	Reverse     []int  `json:"reverse"`
	Start       []int  `json:"start"`
	Type        string `json:"type"`
	Volume      []int  `json:"volume"`
}

// GetStringInBetween returns empty string if no start or end string found
func GetStringInBetween(str string, start string, end string) (result string) {
	s := strings.Index(str, start)
	if s == -1 {
		return
	}
	s += len(start)
	e := strings.Index(str[s:], end)
	if e == -1 {
		return
	}
	return str[s : s+e]
}

// TempFileName generates a temporary filename for use in testing or whatever
func TempFileName(prefix, suffix string) string {
	randBytes := make([]byte, 16)
	rand.Read(randBytes)

	return filepath.Join(".", prefix+hex.EncodeToString(randBytes)+suffix)
}
