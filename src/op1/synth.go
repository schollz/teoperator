package op1

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"strings"
)

var defaultSynthPatch SynthPatch

func init() {
	b := []byte(`{"adsr":[64,64,0,64,14336,64,4000,4000],"fx_active":true,"fx_params":[64,-14337,4515,7232,0,0,0,0],"fx_type":"nitro","knobs":[3072,0,512,3,0,0,0,0],"lfo_active":false,"lfo_params":[4608,32767,8448,15360,0,0,0,0],"lfo_type":"value","name":"default","octave":0,"synth_version":2,"type":"cluster"}`)
	err := json.Unmarshal(b, &defaultSynthPatch)
	if err != nil {
		panic(err)
	}

}

type SynthPatch struct {
	Adsr         [8]int `json:"adsr"`
	FxActive     bool   `json:"fx_active"`
	FxParams     [8]int `json:"fx_params"`
	FxType       string `json:"fx_type"`
	Knobs        [8]int `json:"knobs"`
	LfoActive    bool   `json:"lfo_active"`
	LfoParams    [8]int `json:"lfo_params"`
	LfoType      string `json:"lfo_type"`
	Name         string `json:"name"`
	Octave       int    `json:"octave"`
	SynthVersion int    `json:"synth_version"`
	Type         string `json:"type"`
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

	AllowedEngine = map[string][][]int{
		"cluster": [][]int{
			Range(3072, 17408, 128),
			Range(0, 32767, 128),
			Range(512, 24064, 128),
			Range(3, 1638, 128),
		},
		"digital": [][]int{
			Range(0, 32767, 128),
			Range(2048, 26624, 128),
			Range(-32768, 32767, 128),
			Range(0, 32767, 128),
		},
		"dna": [][]int{
			Range(-29491, 32767, 128),
			Range(4608, 12800, 128),
			Range(0, 32767, 128),
			Range(0, 32767, 128),
		},
		"drwave": [][]int{
			Range(0, 32767, 128),
			Range(0, 32767, 128),
			Range(0, 32767, 128),
			Range(0, 32767, 128),
			Range(32000, 32000, 128),
		},
		// TODO rest of engines
	}

	AllowedEffects = map[string][][]int{
		"nitro": [][]int{
			Range(64, 16448, 128),
			Range(-32768, 32768, 512),
			Range(0, 20643, 128),
			Range(64, 16448, 128),
		},
		"cwo": [][]int{
			Range(0, 32767, 128),
			Range(0, 32767, 128),
			Range(0, 32767, 128),
			Range(0, 32767, 128),
		},
		"delay": [][]int{
			Range(1024, 11264, 128),
			Range(3276, 32767, 128),
			Range(0, 16384, 128),
			Range(0, 32767, 128),
		},
		"grid": [][]int{
			Range(1344, 16704, 128),
			Range(1344, 16704, 128),
			Range(0, 32767, 128),
			Range(0, 32767, 128),
			Range(8000, 8000, 128),
			Range(8000, 8000, 128),
			Range(8000, 8000, 128),
			Range(8000, 8000, 128),
		},
		// TODO rest of effects
	}

	AllowedLFO = map[string][][]int{
		"element": [][]int{
			[]int{7168, 5280, 2000, 2144},   // sum, adsr, g, mic
			Range(-32767, 32767, 512),       // speed
			[]int{1024, 2448, 5056, 7168},   // wave, adsr, fx, sound
			[]int{1024, 5824, 10526, 15360}, // blue, green, white, red
		},
		"tremelo": [][]int{
			Range(16400, 32440, 512),  // speed
			Range(-32767, 32767, 512), // pitch flucuation
			Range(-32767, 32767, 512), // volume flucuation
			Range(0, 32767, 512),      // slope
			Range(0, 0, 512),          // n/a
			Range(0, 0, 512),          // n/a
			Range(0, 0, 512),          // n/a
			[]int{0, 9216},            // wave, triangle TODO: get rest of these parameters
		},
	}
)

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
	if _, ok := AllowedEngine[s.Type]; !ok {
		err = fmt.Errorf("engine '%s' not available", s.Type)
		return
	}
	for i := 0; i < len(AllowedEngine[s.Type]); i++ {
		min := AllowedEngine[s.Type][i][0]
		max := AllowedEngine[s.Type][i][len(AllowedEngine[s.Type][i])-1]
		if s.Knobs[i] < min || s.Knobs[i] > max {
			err = fmt.Errorf("knob %d value %d is not in bounds for %s", i, s.Knobs[i], s.Type)
			return
		}
	}

	// check effects knobs
	if _, ok := AllowedEffects[s.FxType]; !ok {
		err = fmt.Errorf("effect '%s' not available", s.FxType)
		return
	}
	for i := 0; i < len(AllowedEffects[s.FxType]); i++ {
		min := AllowedEffects[s.FxType][i][0]
		max := AllowedEffects[s.FxType][i][len(AllowedEffects[s.FxType][i])-1]
		if s.Knobs[i] < min || s.Knobs[i] > max {
			err = fmt.Errorf("knob %d value %d is not in bounds for %s", i, s.Knobs[i], s.FxType)
			return
		}
	}

	// check lfo knobs
	if _, ok := AllowedLFO[s.LfoType]; !ok {
		err = fmt.Errorf("effect '%s' not available", s.LfoType)
		return
	}
	for i := 0; i < len(AllowedLFO[s.LfoType]); i++ {
		min := AllowedLFO[s.LfoType][i][0]
		max := AllowedLFO[s.LfoType][i][len(AllowedLFO[s.LfoType][i])-1]
		if s.Knobs[i] < min || s.Knobs[i] > max {
			err = fmt.Errorf("knob %d value %d is not in bounds for %s", i, s.Knobs[i], s.FxType)
			return
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

func (s SynthPatch) Save(fnameOut string) (err error) {
	if !strings.HasSuffix(fnameOut, ".aif") {
		err = fmt.Errorf("%s does not have .aif", fnameOut)
		return
	}

	// validate the patch
	if err = s.Check(); err != nil {
		return
	}

	// clip out the current data
	b := defaultSynthAif
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
