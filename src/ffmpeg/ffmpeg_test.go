package ffmpeg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalize(t *testing.T) {
	assert.Nil(t, Normalize("normalize.aif", "normalized.aif"))
}
