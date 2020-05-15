package build

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDrumpatchFromAudio(t *testing.T) {
	_, err := DrumpatchFromAudio("../audiosegment/tests/1.aif")
	assert.Nil(t, err)
}
