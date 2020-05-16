package audiosegment

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitAndMerge(t *testing.T) {
	err := Truncate("tests/1.aif", "1.wav", "0", "00:01:00")
	assert.Nil(t, err)

	segments, err := SplitOnSilence("1.wav", -18, 0.05)
	assert.Nil(t, err)
	for _, segment := range segments {
		fmt.Printf("segment: %+v\n", segment)
	}

	splitSegments, err := Split(segments, "split", true)
	assert.Nil(t, err)
	for _, segment := range splitSegments {
		fmt.Printf("segment: %+v\n", segment)
	}

	mergedSegments, err := Merge(splitSegments, "merge", 6)
	assert.Nil(t, err)
	for _, segment := range mergedSegments {
		fmt.Printf("segment: %+v\n", segment)
	}

	// split the new merged segments on silence
	for i, mergeSegment := range mergedSegments {
		mergedSegmentsSegments, err := SplitOnSilence(mergeSegment.Filename, -25, 0.05)
		assert.Nil(t, err)
		for _, segment := range mergedSegmentsSegments {
			fmt.Printf("segment: %+v\n", segment)
		}

		splitMergedSegments, err := Split(mergedSegmentsSegments, fmt.Sprintf("splitmerge%d", i), false)
		assert.Nil(t, err)
		for _, segment := range splitMergedSegments {
			fmt.Printf("segment: %+v\n", segment)
			os.Remove(segment.Filename)
		}
	}
}

func TestMerge(t *testing.T) {
	segment, err := MergeAudioFiles([]string{"tests/1.aif", "tests/1.aif"}, "1-1.wav")
	assert.Nil(t, err)
	fmt.Println(segment)
}

func TestSplitEqual(t *testing.T) {
	segments, err := SplitEqual("tests/creeley.mp3", 11.5, 1)
	assert.Nil(t, err)
	fmt.Println(segments)
}

func Test111(t *testing.T) {
	segs, err := SplitOnSilence("tests/creeley-0.025-11.525.mp3", -22, 0.2)
	assert.Nil(t, err)
	for _, segment := range segs {
		fmt.Printf("segment: %+v\n", segment)
	}
	err = DrawSegments(segs)
	assert.Nil(t, err)
}
