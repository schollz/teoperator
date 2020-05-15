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
