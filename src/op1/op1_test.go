package op1

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDrumPatch(t *testing.T) {
	assert.Nil(t, DrumPatch("tests/1.aif", "drum.aif", Default()))
}

func TestReadSynth(t *testing.T) {
	sp, err := ReadSynthPatch("reverse/engines/cluster_min.aif")
	assert.Nil(t, err)
	fmt.Println(sp.Knobs)
	sp, err = ReadSynthPatch("reverse/engines/cluster_max.aif")
	assert.Nil(t, err)
	fmt.Println(sp.Knobs)

	sp, err = ReadSynthPatch("reverse/engines/dna_min.aif")
	assert.Nil(t, err)
	fmt.Println(sp.Knobs)
	sp, err = ReadSynthPatch("reverse/engines/dna_max.aif")
	assert.Nil(t, err)
	fmt.Println(sp.Knobs)

	sp, err = ReadSynthPatch("reverse/portamendo/portamendo_off.aif")
	assert.Nil(t, err)
	fmt.Println(sp.Adsr)
	sp, err = ReadSynthPatch("reverse/portamendo/portamendo1.aif")
	assert.Nil(t, err)
	fmt.Println(sp.Adsr)
	sp, err = ReadSynthPatch("reverse/portamendo/portamendo_127.aif")
	assert.Nil(t, err)
	fmt.Println(sp.Adsr)

	sp, err = ReadSynthPatch("reverse/adsr/adsr_0.aif")
	assert.Nil(t, err)
	fmt.Println(sp.Adsr)
	sp, err = ReadSynthPatch("reverse/adsr/adsr_1.aif")
	assert.Nil(t, err)
	fmt.Println(sp.Adsr)
	sp, err = ReadSynthPatch("reverse/adsr/adsr_max.aif")
	assert.Nil(t, err)
	fmt.Println(sp.Adsr)

	sp, err = ReadSynthPatch("reverse/playmode/playmode_0_poly.aif")
	assert.Nil(t, err)
	fmt.Println(sp.Adsr[Playmode])
	sp, err = ReadSynthPatch("reverse/playmode/playmode_1_mono.aif")
	assert.Nil(t, err)
	fmt.Println(sp.Adsr[Playmode])
	sp, err = ReadSynthPatch("reverse/playmode/playmode_3_unison.aif")
	assert.Nil(t, err)
	fmt.Println(sp.Adsr[Playmode])

	fmt.Println(sp)
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
