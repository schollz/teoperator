package main

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-audio/aiff"
)

func main() {

	err := run()
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
	b, err := ioutil.ReadFile("test1.aif")
	if err != nil {
		return
	}
	start := bytes.Index(b, []byte("APPL"))
	if start < 0 {
		return
	}
	end := bytes.Index(b[start:], []byte("SSND"))
	if end < 0 {
		return
	}

	fmt.Println(b[start : start+end+4])
	fmt.Println(string(b[start+12 : start+end-2]))

	var mySlice = []byte{0, 0, 4, 198}
	data := binary.BigEndian.Uint32(mySlice)
	fmt.Println(data)

	bs := make([]byte, 4)
	binary.BigEndian.PutUint32(bs, data)
	fmt.Println(bs)

	startBytes := []byte{65, 80, 80, 76, 0, 0, 4, 198, 111, 112, 45, 49}
	endBytes := []byte{10, 32}
	fmt.Println(bytes.Equal(startBytes, b[start:start+12]))
	fmt.Println(bytes.Equal(endBytes, b[start+end-2:start+end]))
	fmt.Printf("starter %+v", b[start:start+12])
	fmt.Printf("end %+v", b[start+end-2:start+end])
	fmt.Println(len(b[start+12 : start+end-2]))
	var op1data OP1MetaData
	err = json.Unmarshal(b[start+12:start+end-2], &op1data)
	if err != nil {
		return
	}
	fmt.Printf("\n op1data: %+v\n", op1data)
	fmt.Println(len(op1data.End))

	// modify one end point of the first one
	op1data.Start[0] = int(0.2 * 44100 * 4096)
	op1data.End[0] = int(0.4 * 44100 * 4096)
	op1data.Start[1] = int(0.4 * 44100 * 4096)

	bop1, err := json.Marshal(op1data)
	if err != nil {
		return
	}
	fmt.Println(string(bop1))
	fmt.Println(len(bop1))

	filler := []byte{10}

	for {
		b2 := append([]byte{}, b[:start+4]...)
		// new size
		bsSize := make([]byte, 4)
		binary.BigEndian.PutUint32(bsSize, uint32(4+len(filler)+len(bop1)))
		fmt.Println(bsSize)
		b2 = append(b2, bsSize...)
		b2 = append(b2, []byte{111, 112, 45, 49}...)
		b2 = append(b2, bop1...)
		fmt.Println(len(b2))
		b2 = append(b2, filler...)
		b2 = append(b2, b[start+end:]...)

		totalsize := len(b2) - 8
		fmt.Println(totalsize)
		bsTotalSize := make([]byte, 4)
		binary.BigEndian.PutUint32(bsTotalSize, uint32(totalsize))
		b3 := append([]byte{}, b2[:4]...)
		b3 = append(b3, bsTotalSize...)
		b3 = append(b3, b2[8:]...)
		if math.Mod(float64(totalsize), 4.0) == 0 {
			err = ioutil.WriteFile("test2.aif", b3, 0644)
			break
		} else {
			filler = append(filler, []byte{30}...)
			fmt.Println(filler)
		}
	}

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

func DefaultOP1() OP1MetaData {
	b := []byte(`{"drum_version":2,"dyna_env":[0,8192,0,8192,0,0,0,0],"end":[97643143,165163892,211907777,282025634,313446583,372916297,413167412,454132733,478541489,492549640,582028126,642634075,642634075,642634075,642634075,642634075,642634075,642634075,642634075,642634075,642634075,642634075,642634075,2032606256],"fx_active":false,"fx_params":[8000,8000,8000,8000,8000,8000,8000,8000],"fx_type":"delay","lfo_active":false,"lfo_params":[16000,16000,16000,16000,0,0,0,0],"lfo_type":"tremolo","name":"boombap1","octave":0,"pitch":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],"playmode":[8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192],"reverse":[8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192],"start":[0,97647201,165167950,211911835,282029692,313450641,372920355,413171470,454136790,478545547,492553698,582032184,582032184,582032184,582032184,582032184,582032184,582032184,582032184,582032184,582032184,582032184,582032184,642638133],"type":"drum","volume":[8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192]}`)
	var op1data OP1MetaData
	json.Unmarshal(b, &op1data)
	return op1data
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
