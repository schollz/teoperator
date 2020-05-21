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

var defaultDrumPatch DrumPatch

func init() {
	var b = []byte(`{"drum_version":2,"dyna_env":[0,8192,0,8192,0,0,0,0],"end":[97643143,165163892,211907777,282025634,313446583,372916297,413167412,454132733,478541489,492549640,582028126,642634075,642634075,642634075,642634075,642634075,642634075,642634075,642634075,642634075,642634075,642634075,642634075,2032606256],"fx_active":false,"fx_params":[8000,8000,8000,8000,8000,8000,8000,8000],"fx_type":"delay","lfo_active":false,"lfo_params":[16000,16000,16000,16000,0,0,0,0],"lfo_type":"tremolo","name":"boombap1","octave":0,"pitch":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],"playmode":[8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192],"reverse":[8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192],"start":[0,97647201,165167950,211911835,282029692,313450641,372920355,413171470,454136790,478545547,492553698,582032184,582032184,582032184,582032184,582032184,582032184,582032184,582032184,582032184,582032184,582032184,582032184,642638133],"type":"drum","volume":[8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192,8192]}`)
	err := json.Unmarshal(b, &defaultDrumPatch)
	if err != nil {
		panic(err)
	}
}

// DrumPatch is a list of custom fields sometimes set by OP-1
type DrumPatch struct {
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

// NewDrumPatch returns a new DrumPatch with correct defaults
func NewDrumPatch() DrumPatch {
	return defaultDrumPatch
}

// Save creates a drum patch from op1 meta data and a song clip
func (drumpatch *DrumPatch) Save(audioClip string, fnameOut string) (err error) {
	if !strings.HasSuffix(fnameOut, ".aif") {
		err = fmt.Errorf("%s does not have .aif", fnameOut)
		return
	}
	// generate a merged audio waveform, downsampled to 1 channel
	cmd := fmt.Sprintf("-y -i %s -ss 0 -to 11.5 -ar 44100  -ac 1 %s", audioClip, fnameOut)
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

	// normalize drumpatch, all the start/stop blocks need to be factors of 8192
	for i := range drumpatch.End {
		drumpatch.End[i] = drumpatch.End[i] * 8192 / 8192
	}
	for i := range drumpatch.Start {
		drumpatch.Start[i] = drumpatch.Start[i] * 8192 / 8192
	}

	op1dataBytes, err := json.Marshal(drumpatch)
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
