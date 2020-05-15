package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertToSeconds(t *testing.T) {
	seconds, err := ConvertToSeconds("00:11.5")
	assert.Nil(t, err)
	assert.Equal(t, 11.5, seconds)
	seconds, err = ConvertToSeconds("00:00:11.5")
	assert.Nil(t, err)
	assert.Equal(t, 11.5, seconds)
	seconds, err = ConvertToSeconds("00:01:11.5")
	assert.Nil(t, err)
	assert.Equal(t, 71.5, seconds)
	seconds, err = ConvertToSeconds("01:01:11.5")
	assert.Nil(t, err)
	assert.Equal(t, 3671.5, seconds)
	seconds, err = ConvertToSeconds("11.5")
	assert.Nil(t, err)
	assert.Equal(t, 11.5, seconds)
}

func TestSecondsToString(t *testing.T) {
	assert.Equal(t, "00:01:04.23", SecondsToString(64.23))
	assert.Equal(t, "00:01:30.23", SecondsToString(90.23))
	assert.Equal(t, "00:03:00.23", SecondsToString(180.23))
	assert.Equal(t, "01:01:30.23", SecondsToString(3690.23))
}
