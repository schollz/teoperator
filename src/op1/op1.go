package op1

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os/exec"
	"strings"

	"github.com/schollz/logger"
)

var op1default OP1MetaData

func init() {
	var b = []byte(`{"drum_version":2,"dyna_env":[0,8192,0,8192,0,0,0,0],"end":[97643143,165163892,211907777,282025634,313446583,372916297,413167412,454132733,478541489,492549640,582028126,642634075,642634075,642634075,642634075,642634075,642634075,642634075,642634075,642634075,642634075,642634075,642634075,2032606256],"fx_active":false,"fx_params":[8000,8000,8000,8000,8000,8000,8000,8000],"fx_type":"delay","lfo_active":false,"lfo_params":[16000,16000,16000,16000,0,0,0,0],"lfo_type":"tremolo","name":"boombap1","octave":0,"pitch":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],"playmode":[8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192],"reverse":[8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192],"start":[0,97647201,165167950,211911835,282029692,313450641,372920355,413171470,454136790,478545547,492553698,582032184,582032184,582032184,582032184,582032184,582032184,582032184,582032184,582032184,582032184,582032184,582032184,642638133],"type":"drum","volume":[8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192]}`)
	err := json.Unmarshal(b, &op1default)
	if err != nil {
		panic(err)
	}
}

// Default returns the default OP1 struct
func Default() OP1MetaData {
	return op1default
}

// OP1MetaData is a list of custom fields sometimes set by OP-1
type OP1MetaData struct {
	DrumVersion int     `json:"drum_version"`
	DynaEnv     []int64 `json:"dyna_env"`
	End         []int64 `json:"end"`
	FxActive    bool    `json:"fx_active"`
	FxParams    []int64 `json:"fx_params"`
	FxType      string  `json:"fx_type"`
	LfoActive   bool    `json:"lfo_active"`
	LfoParams   []int64 `json:"lfo_params"`
	LfoType     string  `json:"lfo_type"`
	Name        string  `json:"name"`
	Octave      int64   `json:"octave"`
	Pitch       []int64 `json:"pitch"`
	Playmode    []int64 `json:"playmode"`
	Reverse     []int64 `json:"reverse"`
	Start       []int64 `json:"start"`
	Type        string  `json:"type"`
	Volume      []int64 `json:"volume"`
}

type SynthPatch struct {
	Adsr [8]int `json:"adsr"`
	// Attack [64, 16320],, increments of 512
	// Decay [64, 16320], increments of 512
	// Sustain [0, 32767], increments of 512
	// Release [64, 16320], increments of 512
	// Playmode values 2048, 5120, 14336 (poly, mono, unison)
	// Portamendo 64 = off, value of [192, 6140] correspond to Portamendo [1,127]
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

	err = json.Unmarshal(b[index1+4:index2+index1+1], &sp)

	return

}

// DrumPatch creates a drum patch from op1 meta data and a song clip
func DrumPatch(fnameIn string, fnameOut string, op1data OP1MetaData) (err error) {
	if !strings.HasSuffix(fnameOut, ".aif") {
		err = fmt.Errorf("%s does not have .aif", fnameOut)
		return
	}
	// generate a merged audio waveform, downsampled to 1 channel
	cmd := fmt.Sprintf("-y -i %s -ss 0 -to 11.5 -ar 44100  -ac 1 %s", fnameIn, fnameOut)
	logger.Debug(cmd)
	out, err := exec.Command("ffmpeg", strings.Fields(cmd)...).CombinedOutput()
	if err != nil {
		logger.Errorf("ffmpeg: %s", out)
		return
	}

	// inject the OP-1 metadata before teh SSND tag
	b, err := ioutil.ReadFile(fnameOut)
	if err != nil {
		return
	}

	ssndTagPosition := bytes.Index(b, []byte("SSND"))
	if ssndTagPosition < 0 {
		err = fmt.Errorf("no SND tag")
		return
	}

	// normalize op1data, all the start/stop blocks need to be factors of 8192
	for i := range op1data.End {
		op1data.End[i] = op1data.End[i] * 8192 / 8192
	}
	for i := range op1data.Start {
		op1data.Start[i] = op1data.Start[i] * 8192 / 8192
	}

	op1dataBytes, err := json.Marshal(op1data)
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

		// repeat until the the total bytes is a multiple of 4
		if math.Mod(float64(totalsize), 4.0) == 0 {
			err = ioutil.WriteFile(fnameOut, b3, 0644)
			break
		} else {
			filler = append(filler, []byte{30}...)
		}
	}

	return
}
