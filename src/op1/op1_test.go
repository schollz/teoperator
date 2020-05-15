package op1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDrumPatch(t *testing.T) {
	assert.Nil(t, DrumPatch("tests/1.aif", "drum.aif", Default()))
}
