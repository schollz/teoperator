package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/go-audio/aiff"
)

func main() {
	err := run2()
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
	fmt.Println(a.Data)
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

	op1data.End[0] = op1data.End[0] / 2

	bop1, err := json.Marshal(op1data)
	if err != nil {
		return
	}
	b2 := append([]byte{}, b[:start+4]...)
	b2 = append(b2, bop1...)
	b2 = append(b2, b[start+end+1:]...)
	err = ioutil.WriteFile("2.tif", b2, 0644)

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
