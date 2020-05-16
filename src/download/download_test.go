package download

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDownloadTooBig(t *testing.T) {
	assert.NotNil(t, Download("https://dl.google.com/go/go1.14.3.windows-amd64.msi", "toobig", 5000))
	assert.Nil(t, Download("https://dl.google.com/go/go1.14.3.src.tar.gz", "nottoobig", 5000000000))
}

func TestYoutube(t *testing.T) {
	err := Youtube("https://www.youtube.com/watch?v=cssXKXCXdLA", "test.mp3")
	assert.Nil(t, err)
}
