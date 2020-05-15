package ffmpeg

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitOnSilence(t *testing.T) {
	segments, err := SplitOnSilence("tests/1.aif", -18, 0.05)
	assert.Nil(t, err)
	for _, segment := range segments {
		fmt.Printf("segment: %+v\n", segment)
	}

	splitSegments, err := Split(segments, "split")
	assert.Nil(t, err)
	for _, segment := range splitSegments {
		fmt.Printf("segment: %+v\n", segment)
	}
}
