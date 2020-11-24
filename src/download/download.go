package download

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/schollz/logger"
	"github.com/schollz/teoperator/src/utils"
)

var Duct = ""
var ServerName = "alsdkjflkasjdlfajsld"

// PassThru wraps an existing io.Reader.
//
// It simply forwards the Read() call, while displaying
// the results from individual calls to it.
type PassThru struct {
	io.Reader
	total     int64 // Total # of bytes transferred
	byteLimit int64
}

// Read 'overrides' the underlying io.Reader's Read method.
// This is the one that will be called by io.Copy(). We simply
// use it to keep track of byte counts and then forward the call.
func (pt *PassThru) Read(p []byte) (int, error) {
	n, err := pt.Reader.Read(p)
	if err == nil {
		pt.total += int64(n)
	}
	if pt.total > pt.byteLimit {
		err = fmt.Errorf("too many bytes")
	}

	return n, err
}

// Download a file and limit the number of bytes. If the bytes exceed,
// it will throw an error and delete the downloaded file.
func Download(u string, fname string, byteLimit int64) (alternativeName string, err error) {
	if Duct != "" && !strings.Contains(u, ServerName) {
		return DownloadFromDuct(u, fname)
	}
	return download(u, fname, byteLimit)
}

func download(u string, fname string, byteLimit int64) (alternativeName string, err error) {
	// download youtube
	if strings.Contains(u, "youtube") || strings.Contains(u, "instagram") || strings.Contains(u, "soundcloud") {
		return Youtube(u, fname)
	}

	// Get the data
	resp, err := http.Get(u)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(fname)
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			os.Remove(fname)
		}
	}()
	defer out.Close()

	// Wrap it with our custom io.Reader.
	src := &PassThru{Reader: resp.Body, byteLimit: byteLimit}

	_, err = io.Copy(out, src)

	return
}

func Youtube(u string, fname string) (alternativeName string, err error) {
	cmd := fmt.Sprintf("--extract-audio --audio-format mp3 %s",
		u,
	)
	logger.Debug(cmd)
	out, err := exec.Command("youtube-dl", strings.Fields(cmd)...).CombinedOutput()
	if err != nil {
		logger.Errorf("youtube-dl: %s", out)
		return
	}
	destFile := ""
	for _, line := range strings.Split(string(out), "\n") {
		if !strings.Contains(line, "Destination:") {
			continue
		}
		foo := strings.Split(line, "Destination:")
		destFile = strings.TrimSpace(foo[1])
		break
	}
	if destFile == "" {
		err = fmt.Errorf("no dest file")
		return
	}
	alternativeName = destFile
	err = os.Rename(destFile, fname)
	return
}

// worker stuff

type Job struct {
	Job             string `json:"j,omitempty"`
	Data            []byte `json:"d,omitempty"`
	AlternativeName string `json:"a,omitempty"`
	Error           string `json:"e,omitempty"`
}

func DownloadFromDuct(u string, fname string) (alternativeName string, err error) {
	// establish channel
	specialChannel := utils.RandStringBytesMaskImpr(8)
	err = sendjob(Duct, Job{
		Job: specialChannel,
	}, 1*time.Minute)
	if err != nil {
		logger.Error(err)
		return
	}

	// send job to new channel
	err = sendjob(specialChannel, Job{
		Job: u,
	}, 10*time.Second)
	if err != nil {
		logger.Error(err)
		return
	}

	// get job from special channel
	j, err := getjob(specialChannel, 10*time.Minute)
	if err != nil {
		logger.Error(err)
		return
	}
	err = ioutil.WriteFile(fname, j.Data, 0644)
	if err != nil {
		logger.Error(err)
		return
	}
	alternativeName = j.AlternativeName
	return
}

func Work() (err error) {
	for {
		err = dowork()
		if err != nil {
			time.Sleep(1 * time.Second)
		}
	}
	return
}

func dowork() (err error) {
	// get channel
	logger.Debugf("listening to %s for job", Duct)
	j, err := getjob(Duct, 10*time.Hour)
	if err != nil {
		logger.Error(err)
		return
	}

	// subscribe to channel
	specialChannel := j.Job
	logger.Debugf("subscribing to channel %s", specialChannel)
	j, err = getjob(specialChannel, 10*time.Second)
	if err != nil {
		logger.Error(err)
		return
	}
	// send job back
	defer func() {
		if err != nil {
			j.Error = err.Error()
		}
		err = sendjob(specialChannel, j, 10*time.Minute)
		if err != nil {
			logger.Error(err)
		}
	}()

	tempfile := utils.RandStringBytesMaskImpr(8)
	defer os.Remove(tempfile)
	j.AlternativeName, err = download(j.Job, tempfile, 1000000000)
	if err != nil {
		logger.Error(err)
		return
	}
	j.Data, err = ioutil.ReadFile(tempfile)
	if err != nil {
		logger.Error(err)
		return
	}
	return
}

func getjob(duct string, timeout time.Duration) (j Job, err error) {
	var myClient = &http.Client{Timeout: timeout}
	r, err := myClient.Get("https://duct.schollz.com/" + duct + Duct)
	if err != nil {
		return
	}
	defer r.Body.Close()
	err = json.NewDecoder(r.Body).Decode(&j)
	return
}

func sendjob(duct string, j Job, timeout time.Duration) (err error) {
	b, err := json.Marshal(j)
	if err != nil {
		return
	}
	logger.Debugf("sending job via %s: %s", duct, j.Job)
	var myClient = &http.Client{Timeout: timeout}
	_, err = myClient.Post("https://duct.schollz.com/"+duct+Duct, "application/json", bytes.NewBuffer(b))
	return
}
