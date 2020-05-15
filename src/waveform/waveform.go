package waveform

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/schollz/logger"
)

// Image generates image of the waveform given a filename
func Image(fnameIn, color string, length float64) (err error) {
	cmd := fmt.Sprintf("-i %s -o %s.png --background-color ffffff00 --waveform-color %s --amplitude-scale 1 --no-axis-labels --pixels-per-second 100 --height 120 --width %2.0f",
		fnameIn, fnameIn, color, length*100,
	)
	logger.Debug(cmd)
	out, err := exec.Command("audiowaveform", strings.Fields(cmd)...).CombinedOutput()
	if err != nil {
		logger.Errorf("audiowaveform: %s", out)
	}
	return
}
