package transform

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAudioToDrumPatch(t *testing.T) {
	_, err := AudioToDrumPatch("../audiosegment/tests/1.aif")
	assert.Nil(t, err)
}
