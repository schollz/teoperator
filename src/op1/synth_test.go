package op1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildSynth(t *testing.T) {
	sp := NewSynthPatch()
	assert.Nil(t, sp.Build("Piano.mf.D3.aiff", true))
}
