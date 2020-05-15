package op1

import "encoding/json"

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
