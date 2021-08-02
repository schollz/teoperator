package convert

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	log "github.com/schollz/logger"
	"github.com/schollz/teoperator/src/ffmpeg"
)

func ToDrum(fnames []string) (err error) {
	finalName := strings.TrimSuffix(fnames[0], filepath.Ext(fnames[0])) + "_patch.wav"
	log.Debugf("converting %+v", fnames)
	f, err := ioutil.TempFile(".", "concat")
	defer os.Remove(f.Name())
	sampleEnd := make([]int64, len(fnames))
	for i, fname := range fnames {
		var fname2 string
		fname2, err = ffmpeg.ToMono(fname)
		defer os.Remove(fname2)
		if err != nil {
			return
		}
		_, fnames[i] = filepath.Split(fname2)
		sampleEnd[i], err = ffmpeg.NumSamples(fname2)
		if err != nil {
			return
		}
		if i > 0 {
			sampleEnd[i] = sampleEnd[i] + sampleEnd[i-1]
		}
	}
	f.Close()

	log.Debug(fnames)
	fname2, err := ffmpeg.Concatenate(fnames)
	defer os.Remove(fname2)
	if err != nil {
		return
	}

	fname3, err := ffmpeg.ToMono(fname2)
	if err != nil {
		return
	}
	os.Rename(fname3, finalName)
	log.Infof("written to %s", finalName)

	return
}
