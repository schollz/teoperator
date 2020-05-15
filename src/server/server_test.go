package server

import (
	"testing"

	log "github.com/schollz/logger"
	"github.com/stretchr/testify/assert"
)

func TestGenerateUserData(t *testing.T) {
	log.SetLevel("debug")
	u := `https://upload.wikimedia.org/wikipedia/commons/6/68/Turdus_merula_male_song_at_dawn%2820s%29.ogg`
	startStop := []float64{0, 10}
	assert.Nil(t, generateUserData(u, startStop))
}
