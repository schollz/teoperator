package op1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildSynth(t *testing.T) {
	sp := NewSynthPatch()
	assert.Nil(t, sp.SaveSample("Piano.mf.D3.aiff", "mfd3.aif", true))
}
