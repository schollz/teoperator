package op1

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/schollz/logger"
	"github.com/schollz/teoperator/src/ffmpeg"
	"github.com/schollz/teoperator/src/models"
	"github.com/speps/go-hashids"
)

var defaultSynthPatch SynthPatch
var defaultSynthPatchSampler SynthPatch

func init() {
	b := []byte(`{"adsr":[64,64,0,64,14336,64,4000,4000],"fx_active":true,"fx_params":[64,-14337,4515,7232,0,0,0,0],"fx_type":"nitro","knobs":[3072,0,512,3,0,0,0,0],"lfo_active":false,"lfo_params":[4608,32767,8448,15360,0,0,0,0],"lfo_type":"value","name":"default","octave":0,"synth_version":2,"type":"cluster"}`)
	err := json.Unmarshal(b, &defaultSynthPatch)
	if err != nil {
		panic(err)
	}

	b = []byte(`{"adsr":[64,10746,32767,10000,4000,64,4000,4000],"base_freq":440.0,"fx_active":false,"fx_params":[8000,8000,8000,8000,8000,8000,8000,8000],"fx_type":"delay","knobs":[0,19361,27626,32767,12000,0,0,8192],"lfo_active":false,"lfo_params":[16000,0,0,16000,0,0,0,0],"lfo_type":"tremolo","name":"20911115_1948","octave":0,"synth_version":1,"type":"sampler"}`)
	b = []byte(`{
"adsr":[512,10746,32767,10000,4000,64,4000,4000],
"base_freq":261.6253662109375,
"fx_active":false,
"fx_params":[4480,15544,10788,12104,0,0,0,0],
"fx_type":"delay",
"knobs":[0,0,32767,32767,12000,0,0,8304],
"lfo_active":false,
"lfo_params":[11840,18431,1024,15144,0,0,0,0],
"lfo_type":"random",
"name":"20150108_0251",
"octave":0,
"synth_version":1,
"type":"sampler"
}
`)
	err = json.Unmarshal(b, &defaultSynthPatchSampler)
	if err != nil {
		panic(err)
	}
	rand.Seed(time.Now().Unix())
}

type SynthPatch struct {
	Adsr         [8]int  `json:"adsr"`
	FxActive     bool    `json:"fx_active"`
	FxParams     [8]int  `json:"fx_params"`
	FxType       string  `json:"fx_type"`
	Knobs        [8]int  `json:"knobs"`
	LfoActive    bool    `json:"lfo_active"`
	LfoParams    [8]int  `json:"lfo_params"`
	LfoType      string  `json:"lfo_type"`
	Name         string  `json:"name"`
	Octave       int     `json:"octave"`
	SynthVersion int     `json:"synth_version"`
	Type         string  `json:"type"`
	BaseFreq     float64 `json:"base_freq,omitempty"`
}

// ADSR parameters
const (
	Attack = iota
	Decay
	Sustain
	Release
	Playmode
	Portamendo
)

type Setting struct {
	Name       string
	Parameters [][]int
}

// specified allowed values for parameters
var (
	AllowedADSR = [][]int{
		Range(64, 16320, 512),           // Attack
		Range(64, 16320, 512),           // Decay
		Range(0, 32767, 512),            // Sustain
		Range(64, 16320, 512),           // Release
		[]int{2048, 5120, 11264, 14336}, // Playmode  (poly, mono, legato, unison)
		[]int{64, 192, 6140},            // Portamendo 64 = off, 192 = 1, 6140 = 127 TODO: need more info
	}

	AllowedOctave = []int{-2, 1, 0, 1, 2} // Octave ranges from -2 to +2

	AllowedEngine = []Setting{
		Setting{
			Name: "cluster",
			Parameters: [][]int{
				Range(3072, 17408, 128),
				Range(0, 32767, 128),
				Range(512, 24064, 128),
				Range(3, 1638, 128),
			},
		},
		Setting{
			Name: "digital",
			Parameters: [][]int{
				Range(0, 32767, 128),
				Range(2048, 26624, 128),
				Range(-32768, 32767, 128),
				Range(0, 32767, 128),
			},
		},
		Setting{
			Name: "dna",
			Parameters: [][]int{
				Range(-29491, 32767, 128),
				Range(4608, 12800, 128),
				Range(0, 32767, 128),
				Range(0, 32767, 128),
			},
		},
		Setting{
			Name: "drwave",
			Parameters: [][]int{
				Range(0, 32767, 128),
				Range(0, 32767, 128),
				Range(0, 32767, 128),
				Range(0, 32767, 128),
				Range(32000, 32000, 128),
			},
		},
	}

	AllowedEffects = []Setting{
		Setting{
			Name: "nitro",
			Parameters: [][]int{
				Range(64, 16448, 128),
				Range(-32768, 32768, 512),
				Range(0, 20643, 128),
				Range(64, 16448, 128),
			},
		},
		Setting{
			Name: "cwo",
			Parameters: [][]int{
				Range(0, 32767, 128),
				Range(0, 32767, 128),
				Range(0, 32767, 128),
				Range(0, 32767, 128),
			},
		},
		Setting{
			Name: "delay",
			Parameters: [][]int{
				Range(1024, 11264, 128),
				Range(3276, 32767, 128),
				Range(0, 16384, 128),
				Range(0, 32767, 128),
			},
		},
		Setting{
			Name: "grid",
			Parameters: [][]int{
				Range(1344, 16704, 128),
				Range(1344, 16704, 128),
				Range(0, 32767, 128),
				Range(0, 32767, 128),
				Range(8000, 8000, 128),
				Range(8000, 8000, 128),
				Range(8000, 8000, 128),
				Range(8000, 8000, 128),
			},
		},
	}

	AllowedLFO = []Setting{
		Setting{
			Name: "element",
			Parameters: [][]int{
				[]int{7168, 5056, 5280, 2000, 2144},         // sum, adsr, g, mic
				Range(-32767, 32767, 512),                   // speed
				[]int{1024, 2000, 2448, 5056, 7168},         // wave, adsr, fx, sound
				[]int{1024, 2000, 5056, 5824, 10526, 15360}, // blue, green, white, red
			},
		},
		Setting{
			Name: "tremelo",
			Parameters: [][]int{
				Range(16400, 32440, 512),  // speed
				Range(-32767, 32767, 512), // pitch flucuation
				Range(-32767, 32767, 512), // volume flucuation
				// Range(0, 32767, 512),      // slope
				// Range(0, 0, 512),          // n/a
				// Range(0, 0, 512),          // n/a
				// Range(0, 0, 512),          // n/a
				// []int{0, 9216},            // wave, triangle TODO: get rest of these parameters
			},
		},
	}
)

func NewSynthPatch() (sd SynthPatch) {
	return defaultSynthPatch
}

func NewSynthSamplePatch(freq ...float64) (sd SynthPatch) {
	sd = defaultSynthPatchSampler
	if len(freq) > 0 {
		sd.BaseFreq = freq[0]
	}
	return
}

func RandomSynthPatch(seed ...int64) (sd SynthPatch) {
	sd = NewSynthPatch()

	if len(seed) > 0 {
		rand.Seed(seed[0])
	}
	for i := 0; i < len(AllowedADSR); i++ {
		sd.Adsr[i] = AllowedADSR[i][rand.Intn(len(AllowedADSR[i]))]
	}

	setting := AllowedEngine[rand.Intn(len(AllowedEngine))]
	sd.Type = setting.Name
	for i := 0; i < len(setting.Parameters); i++ {
		sd.Knobs[i] = setting.Parameters[i][rand.Intn(len(setting.Parameters[i]))]
	}

	setting = AllowedEffects[rand.Intn(len(AllowedEffects))]
	sd.FxType = setting.Name
	for i := 0; i < len(setting.Parameters); i++ {
		sd.FxParams[i] = setting.Parameters[i][rand.Intn(len(setting.Parameters[i]))]
	}
	sd.FxActive = true

	setting = AllowedLFO[rand.Intn(len(AllowedLFO))]
	sd.LfoType = setting.Name
	for i := 0; i < len(setting.Parameters); i++ {
		sd.LfoParams[i] = setting.Parameters[i][rand.Intn(len(setting.Parameters[i]))]
	}
	sd.LfoActive = true

	sd.Name = strings.Split(sd.Encode(), "-")[0]
	return
}

func (s SynthPatch) Encode() (encoded string) {
	encoded = s.Type

	encoded += Hashid(s.Knobs[:4])
	encoded += "-" + Hashid(s.Adsr[:4])

	if s.FxActive {
		encoded += "-" + s.FxType + Hashid(s.FxParams[:4])
	} else {
		encoded += "-"
	}

	if s.LfoActive {
		encoded += "-" + s.LfoType + Hashid(s.LfoParams[:4])
	}
	// 	effect := s.FxType[:2]
	// 	lfo := s.LfoType[:2]
	// 	if !s.FxActive {
	// 		effect = "__"
	// 	}
	// 	if !s.LfoActive {
	// 		lfo = "__"
	// 	}
	// fname := fmt.Sprintf("%s%s%s")

	mdhash := fmt.Sprintf("%x", md5.Sum([]byte(encoded)))
	encoded = "p" + mdhash[:6] + "-" + encoded
	return
}

// Check will return an error if any of the values are out of range
func (s SynthPatch) Check() (err error) {
	// check octave
	if s.Octave < -2 || s.Octave > 2 {
		err = fmt.Errorf("octave out of range")
		return
	}

	// check adsr
	for i := 0; i < len(AllowedADSR); i++ {
		min := AllowedADSR[i][0]
		max := AllowedADSR[i][len(AllowedADSR[i])-1]
		if s.Adsr[i] < min || s.Adsr[i] > max {
			err = fmt.Errorf("adsr %d is out of bounds", i)
			return
		}
	}

	// check engine knobs
	for _, setting := range AllowedEngine {
		if setting.Name != s.Type {
			continue
		}
		for i := range setting.Parameters {
			if !Has(setting.Parameters[i], s.Knobs[i]) {
				err = fmt.Errorf("engine %d value %d is not in bounds for %s", i, s.Knobs[i], s.Type)
				return
			}
		}
	}

	for _, setting := range AllowedEffects {
		if setting.Name != s.FxType {
			continue
		}
		for i := range setting.Parameters {
			if !Has(setting.Parameters[i], s.FxParams[i]) {
				err = fmt.Errorf("fx %d value %d is not in bounds for %s", i, s.FxParams[i], s.FxType)
				return
			}
		}
	}

	for _, setting := range AllowedLFO {
		if setting.Name != s.LfoType {
			continue
		}
		for i := range setting.Parameters {
			if !Has(setting.Parameters[i], s.LfoParams[i]) {
				err = fmt.Errorf("lfo %d value %d is not in bounds for %s", i, s.LfoParams[i], s.LfoType)
				return
			}
		}
	}

	return
}

func ReadSynthPatch(fname string) (sp SynthPatch, err error) {
	b, err := ioutil.ReadFile(fname)
	if err != nil {
		return
	}

	index1 := bytes.Index(b, []byte("op-1"))
	if index1 < 0 {
		err = fmt.Errorf("could not find header in '%s'", fname)
		return
	}
	index2 := bytes.Index(b[index1:], []byte("}"))
	if index2 < 0 {
		err = fmt.Errorf("could not find JSON end in '%s'", fname)
		return
	}

	fmt.Println(string(b[index1+4 : index2+index1+1]))
	err = json.Unmarshal(b[index1+4:index2+index1+1], &sp)

	return
}

// Build a synth patch from a file
func (s SynthPatch) SaveSample(fname string, fnameout string, trimSilence bool) (err error) {
	startClip := 0.0
	if trimSilence {
		var silenceSegments []models.AudioSegment
		silenceSegments, err = ffmpeg.SplitOnSilence(fname, -30, .02, 0)
		if err != nil {
			return
		}
		logger.Debugf("silenceSegments: %+v", silenceSegments)
		if silenceSegments[0].End < 2 {
			startClip = silenceSegments[0].End
		} else if silenceSegments[0].Start < 2 {
			startClip = silenceSegments[0].Start
		}
	}

	// generate a truncated, merged audio waveform, downsampled to 1 channel
	fnameDownsampled := fname + ".down.aif"
	defer os.Remove(fnameDownsampled)
	cmd := fmt.Sprintf("-y -i %s -ss %2.4f -to %2.4f -ar 44100  -ac 1 %s", fname, startClip, startClip+5.75, fnameDownsampled)
	logger.Debug(cmd)
	out, err := exec.Command("ffmpeg", strings.Fields(cmd)...).CombinedOutput()
	if err != nil {
		logger.Errorf("ffmpeg: %s", out)
		return
	}

	// normalize
	fnameNormalized := fname + ".norm.aif"
	defer os.Remove(fnameNormalized)
	err = ffmpeg.Normalize(fnameDownsampled, fnameNormalized)
	if err != nil {
		return
	}

	err = s.SaveSynth(fnameout, fnameNormalized)
	return
}

func (s SynthPatch) SaveSynth(fnameOut string, fnamein ...string) (err error) {
	if !strings.HasSuffix(fnameOut, ".aif") {
		err = fmt.Errorf("%s does not have .aif", fnameOut)
		return
	}

	// validate the patch
	if err = s.Check(); err != nil {
		return
	}

	// clip out the current data
	var b []byte
	if len(fnamein) > 0 {
		b, err = ioutil.ReadFile(fnamein[0])
		if err != nil {
			return
		}
	} else {
		// use the default robot op-1 patch
		b = append([]byte{}, defaultSynthAif...)
		index1 := bytes.Index(b, []byte("APPL"))
		if index1 < 0 {
			err = fmt.Errorf("could not find header ")
			return
		}
		index2 := bytes.Index(b, []byte("SSND"))
		if index2 < 0 {
			err = fmt.Errorf("could not find JSON end")
			return
		}
		b = append(b[:index1], b[index2:]...)
	}

	ssndTagPosition := bytes.Index(b, []byte("SSND"))
	if ssndTagPosition < 0 {
		err = fmt.Errorf("no SND tag")
		return
	}

	op1dataBytes, err := json.Marshal(s)
	if err != nil {
		return
	}

	// filler is to pad the aif file so that it is a multiple of 4
	filler := []byte{10}
	for {
		b2 := append([]byte{}, b[:ssndTagPosition]...)
		// 4 bytes for AAPL tag, required to initiate op-1 data
		b2 = append(b2, []byte{65, 80, 80, 76}...)
		// 4 bytes to delcare size
		bsSize := make([]byte, 4)
		binary.BigEndian.PutUint32(bsSize, uint32(4+len(filler)+len(op1dataBytes)))
		b2 = append(b2, bsSize...)
		// 4 bytes to write magic op-1
		b2 = append(b2, []byte{111, 112, 45, 49}...)
		// write the op1 meta data
		b2 = append(b2, op1dataBytes...)
		// add filler
		b2 = append(b2, filler...)
		// write the rest of the bytes
		b2 = append(b2, b[ssndTagPosition:]...)

		// set bytes 4-8 with the total size - 8 bytes
		totalsize := len(b2) - 8
		bsTotalSize := make([]byte, 4)
		binary.BigEndian.PutUint32(bsTotalSize, uint32(totalsize))
		b3 := append([]byte{}, b2[:4]...)
		b3 = append(b3, bsTotalSize...)
		b3 = append(b3, b2[8:]...)

		// repeat until the the total bytes is a multiple of 2
		if math.Mod(float64(totalsize), 2.0) == 0 {
			err = ioutil.WriteFile(fnameOut, b3, 0644)
			break
		} else {
			filler = append(filler, []byte{30}...)
		}
	}
	return
}

// utils

// Range generates a list of numbers from specified range, inclusive
func Range(start, end, inc int) (r []int) {
	if start == end {
		return []int{start}
	}
	for i := start; i <= end; i++ {
		r = append(r, i)
	}
	// make sure the end is the end specified
	if r[len(r)-1] != end {
		r[len(r)-1] = end
	}
	return
}

func Hashid(ints []int) string {
	hd := hashids.NewData()
	hd.Salt = "op-1"
	h, err := hashids.NewWithData(hd)
	if err != nil {
		panic(err)
	}
	i2 := make([]int, len(ints)*2)
	for i, val := range ints {
		if val < 0 {
			i2[(i * 2)] = 0
		} else {
			i2[(i * 2)] = 1
		}
		i2[(i*2)+1] = int(math.Abs(float64(val)))
	}
	id, err := h.Encode(i2)
	if err != nil {
		panic(err)
	}
	return id
}

func Has(list []int, val int) bool {
	for _, val2 := range list {
		if val == val2 {
			return true
		}
	}
	return false
}
