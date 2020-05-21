package op1

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDrumPatch(t *testing.T) {
	dp := NewDrumPatch()
	assert.Nil(t, dp.Save("tests/1.aif", "drum.aif"))
}

func TestReadSynth(t *testing.T) {
	sp, err := ReadSynthPatch("reverse/lfo/tremelo/minspeed_-100_-100_minslope_env0.aif")
	assert.Nil(t, err)
	fmt.Printf("\n%+v\n", sp.LfoParams)
	sp, err = ReadSynthPatch("reverse/lfo/tremelo/maxspeed_100_100_maxslope_env1.aif")
	assert.Nil(t, err)
	fmt.Printf("\n%+v\n", sp.LfoParams)

	// sp, err = ReadSynthPatch("reverse/engines/cluster_max.aif")
	// assert.Nil(t, err)
	// fmt.Println(sp.Knobs)

	// ioutil.WriteFile("default.aif", defaultSynthAif, 0644)
	// sp, err = ReadSynthPatch("reverse/engines/cluster0.aif")
	// assert.Nil(t, err)
	// sp.Knobs[0] = 6999
	// sp.Adsr[0] = 5003
	// sp.Adsr[1] = 5003
	// sp.Adsr[2] = 5003
	// sp.Adsr[3] = 5003
	// assert.Nil(t, sp.Save("default3.aif"))

	// b, _ := ioutil.ReadFile("reverse/engines/cluster0.aif")
	// b64 := base64.StdEncoding.EncodeToString(b)
	// ioutil.WriteFile("out.base64", []byte(b64), 0644)

	// sp, err = ReadSynthPatch("reverse/engines/drwave_min.aif")
	// assert.Nil(t, err)
	// fmt.Println(sp.Knobs)
	// sp, err = ReadSynthPatch("reverse/engines/drwave_max.aif")
	// assert.Nil(t, err)
	// fmt.Println(sp.Knobs)

	// sp, err = ReadSynthPatch("reverse/portamendo/portamendo_off.aif")
	// assert.Nil(t, err)
	// fmt.Println(sp.Adsr)
	// sp, err = ReadSynthPatch("reverse/portamendo/portamendo1.aif")
	// assert.Nil(t, err)
	// fmt.Println(sp.Adsr)
	// sp, err = ReadSynthPatch("reverse/portamendo/portamendo_127.aif")
	// assert.Nil(t, err)
	// fmt.Println(sp.Adsr)

	// sp, err = ReadSynthPatch("reverse/adsr/adsr_0.aif")
	// assert.Nil(t, err)
	// fmt.Println(sp.Adsr)
	// sp, err = ReadSynthPatch("reverse/adsr/adsr_1.aif")
	// assert.Nil(t, err)
	// fmt.Println(sp.Adsr)
	// sp, err = ReadSynthPatch("reverse/adsr/adsr_max.aif")
	// assert.Nil(t, err)
	// fmt.Println(sp.Adsr)

	// sp, err = ReadSynthPatch("reverse/playmode/playmode_0_poly.aif")
	// assert.Nil(t, err)
	// fmt.Println(sp.Adsr[Playmode])
	// sp, err = ReadSynthPatch("reverse/playmode/playmode_1_mono.aif")
	// assert.Nil(t, err)
	// fmt.Println(sp.Adsr[Playmode])
	// sp, err = ReadSynthPatch("reverse/playmode/playmode_3_unison.aif")
	// assert.Nil(t, err)
	// fmt.Println(sp.Adsr[Playmode])

	// fmt.Println(sp)
	// fmt.Println(AllowedAttack)
}
