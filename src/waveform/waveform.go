package waveform

import (
	"fmt"
	"os/exec"

	"github.com/schollz/logger"
)

// Image generates image of the waveform given a filename
func Image(fnameIn, color string, length float64) (err error) {
	cmd := []string{"-i", fnameIn, "-o", fnameIn + ".png", "--background-color", "ffffff00", "--waveform-color", color, "--amplitude-scale", "1", "--no-axis-labels", "--pixels-per-second", "100", "--height", "120", "--width",
		fmt.Sprintf("%2.0f", length*100)}
	logger.Debug(cmd)
	out, err := exec.Command("audiowaveform", cmd...).CombinedOutput()
	if err != nil {
		logger.Errorf("audiowaveform: %s", out)
	}
	return
}
