package convert

import (
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/schollz/logger"
	"github.com/schollz/teoperator/src/ffmpeg"
)

func ToDrum(fnames []string) (err error) {
	log.Debugf("converting %+v", fnames)
	f, err := ioutil.TempFile(".", "concat")
	defer os.Remove(f.Name())
	for i, fname := range fnames {
		var fname2 string
		fname2, err = ffmpeg.ToMono(fname)
		if err != nil {
			return
		}
		_, fnames[i] = filepath.Split(fname2)
		defer os.Remove(fname2)
		log.Debug(ffmpeg.NumSamples(fname2))
	}
	f.Close()

	log.Debug(fnames)
	fname2, err := ffmpeg.Concatenate(fnames)
	log.Debug(fname2)
	return
}
