package server

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/schollz/httpfileserver"
	"github.com/schollz/logger"
	log "github.com/schollz/logger"
	"github.com/schollz/teoperator/src/audiosegment"
	"github.com/schollz/teoperator/src/download"
	"github.com/schollz/teoperator/src/models"
	"github.com/schollz/teoperator/src/op1"
	"github.com/schollz/teoperator/src/utils"
)

func Run(port int) (err error) {
	os.Mkdir("data", os.ModePerm)
	loadTemplates()
	log.Infof("listening on :%d", port)
	http.HandleFunc("/static/", httpfileserver.New("/static/", "static/", httpfileserver.OptionNoCache(true)).Handle())
	http.HandleFunc("/data/", httpfileserver.New("/data/", "data/", httpfileserver.OptionNoCache(true)).Handle())
	http.HandleFunc("/", handler)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	return
}

func handler(w http.ResponseWriter, r *http.Request) {
	t := time.Now().UTC()
	err := handle(w, r)
	if err != nil {
		log.Error(err)
		viewMain(w, r, err.Error(), "main")
	}
	log.Infof("%v %v %v %s\n", r.RemoteAddr, r.Method, r.URL.Path, time.Since(t))
}

type Href struct {
	Value string
	Href  string
	Flag  bool
}

type Metadata struct {
	Name        string
	UUID        string
	OriginalURL string
	Files       []FileData
	Start       float64
	Stop        float64
}

type FileData struct {
	Prefix string
	Start  float64
	Stop   float64
}

type Render struct {
	Title        string
	MessageError string
	MessageInfo  string
	Metadata     Metadata
}

var t map[string]*template.Template
var mu sync.Mutex

func loadTemplates() {
	mu.Lock()
	defer mu.Unlock()
	t = make(map[string]*template.Template)
	funcMap := template.FuncMap{
		"beforeFirstComma": func(s string) string {
			ss := strings.Split(s, ",")
			if len(ss) == 1 {
				return s
			}
			if len(ss[0]) > 8 {
				return strings.TrimSpace(ss[0])
			}
			return strings.TrimSpace(ss[0] + ", " + ss[1])
		},
		"humanizeTime": func(t time.Time) string {
			return humanize.Time(t)
		},
		"add": func(a, b int) int {
			return a + b
		},
		"removeSlashes": func(s string) string {
			return strings.TrimPrefix(strings.TrimSpace(strings.Replace(s, "/", "-", -1)), "-location-")
		},
		"minusOne": func(s int) int {
			return s - 1
		},
		"mod": func(i, j int) bool {
			return i%j == 0
		},
		"urlbase": func(s string) string {
			uparsed, _ := url.Parse(s)
			return filepath.Base(uparsed.Path)
		},
		"filebase": func(s string) string {
			_, base := filepath.Split(s)
			return base
		},
		"roundfloat": func(f float64) string {
			return fmt.Sprintf("%2.1f", f)
		},
	}
	for _, templateName := range []string{"main"} {
		b, err := ioutil.ReadFile("templates/base.html")
		if err != nil {
			panic(err)
		}
		t[templateName] = template.Must(template.New("base").Funcs(funcMap).Delims("((", "))").Parse(string(b)))
		b, err = ioutil.ReadFile("templates/" + templateName + ".html")
		if err != nil {
			panic(err)
		}
		t[templateName] = template.Must(t[templateName].Parse(string(b)))
		log.Tracef("loaded template %s", templateName)
	}

}

func handle(w http.ResponseWriter, r *http.Request) (err error) {
	if log.GetLevel() == "debug" || log.GetLevel() == "trace" {
		loadTemplates()
	}

	if r.URL.Path == "/ws" {
	} else if r.URL.Path == "/favicon.ico" {
		http.Redirect(w, r, "/static/img/favicon.ico", http.StatusFound)
	} else if r.URL.Path == "/robots.txt" {
		http.Redirect(w, r, "/static/robots.txt", http.StatusFound)
	} else if r.URL.Path == "/sitemap.xml" {
		http.Redirect(w, r, "/static/sitemap.xml", http.StatusFound)
	} else if r.URL.Path == "/" {
		return viewMain(w, r, "", "main")
	} else if r.URL.Path == "/patch" {
		return viewPatch(w, r)
	} else {
		t["main"].Execute(w, Render{})
	}

	return
}

func viewPatch(w http.ResponseWriter, r *http.Request) (err error) {
	audioURL, _ := r.URL.Query()["audioURL"]
	secondsStart, _ := r.URL.Query()["secondsStart"]
	secondsEnd, _ := r.URL.Query()["secondsEnd"]
	patchtypeA, _ := r.URL.Query()["patchType"]
	patchtype := "drum"
	if len(patchtypeA) > 0 && patchtypeA[0] == "synth" {
		patchtype = "synth"
	}

	if len(audioURL[0]) == 0 {
		err = fmt.Errorf("no URL")
		return
	}

	startStop := []float64{0, 0}
	if secondsStart[0] != "" {
		startStop[0], _ = strconv.ParseFloat(secondsStart[0], 64)
	}
	if secondsEnd[0] != "" {
		startStop[1], _ = strconv.ParseFloat(secondsEnd[0], 64)
	}

	uuid, err := generateUserData(audioURL[0], startStop, patchtype)
	if err != nil {
		return
	}

	metadatab, err := ioutil.ReadFile(path.Join("data", uuid, "metadata.json"))
	if err != nil {
		return
	}
	var metadata Metadata
	err = json.Unmarshal(metadatab, &metadata)
	if err != nil {
		return
	}

	t["main"].Execute(w, Render{
		Metadata: metadata,
	})
	return
}

func viewMain(w http.ResponseWriter, r *http.Request, messageError string, templateName string) (err error) {

	t[templateName].Execute(w, Render{
		Title:        "Pianos for Travelers",
		MessageError: messageError,
	})
	return
}

func generateUserData(u string, startStop []float64, patchType string) (uuid string, err error) {
	log.Debug(u, startStop)
	if startStop[1]-startStop[0] < 12 {
		startStop[1] = startStop[0] + 60
	}

	uuid = fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%+v %+v", u, startStop))))

	// create path to data
	pathToData := path.Join("data", uuid)

	_, errstat := os.Stat(pathToData)
	if errstat == nil {
		// already exists, done here
		return
	}

	err = os.Mkdir(pathToData, os.ModePerm)
	if err != nil {
		return
	}

	// find filename of downloaded file
	fname := ""
	uparsed, err := url.Parse(u)
	if err != nil {
		return
	}
	fname = path.Join(pathToData, path.Base(uparsed.Path))
	if !strings.Contains(fname, ".") {
		fname += ".mp3"
	}

	fnameID := path.Join("data", fmt.Sprintf("%x%s", md5.Sum([]byte(u)), filepath.Ext(fname)))

	_, errstat = os.Stat(fnameID)
	var alternativeName string
	if errstat != nil {
		log.Debugf("downloading to %s", fnameID)
		alternativeName, err = download.Download(u, fnameID, 100000000)
		if err != nil {
			return
		}

	}

	folder0, _ := filepath.Split(fname)
	shortName := fmt.Sprintf("%x%s", md5.Sum([]byte(u+fmt.Sprintf("%+v", startStop))), filepath.Ext(fname))
	shortName = shortName[:6]
	shortName = path.Join(folder0, shortName+filepath.Ext(fname))

	// // copy file into folder
	// _, err = utils.CopyFile(fnameID, fname)
	// if err != nil {
	// 	return
	// }
	// truncate into folder
	err = audiosegment.Truncate(fnameID, shortName, utils.SecondsToString(startStop[0]), utils.SecondsToString(startStop[1]))
	if err != nil {
		return
	}

	// generate patches
	var segments [][]models.AudioSegment
	if patchType == "drum" {
		segments, err = audiosegment.SplitEqual(shortName, 11.5, 1)
		if err != nil {
			return
		}
	} else {
		segments, err = makeSynthPatch(shortName)
		if err != nil {
			return
		}
	}

	// write metadata
	files := make([]FileData, len(segments))
	for i, seg := range segments {
		files[i] = FileData{
			Prefix: seg[0].Filename[:len(seg[0].Filename)-4],
			Start:  seg[0].StartAbs + startStop[0],
			Stop:   seg[0].EndAbs + startStop[0],
		}
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].Start < files[j].Start
	})

	log.Debug(alternativeName)
	if alternativeName != "" {
		fname = alternativeName
	}
	b, _ := json.Marshal(Metadata{
		Name:        fname,
		UUID:        uuid,
		OriginalURL: u,
		Files:       files,
		Start:       startStop[0],
		Stop:        startStop[1],
	})
	err = ioutil.WriteFile(path.Join(pathToData, "metadata.json"), b, 0644)

	return
}

func makeSynthPatch(fname string) (segments [][]models.AudioSegment, err error) {
	sp := op1.NewSynthPatch()
	fnameout := fname + ".aif"
	err = sp.SaveSample(fname, fnameout, true)
	if err != nil {
		return
	}
	segments = [][]models.AudioSegment{
		[]models.AudioSegment{
			models.AudioSegment{
				Filename: fnameout,
				StartAbs: 0,
				EndAbs:   5.5,
			},
		},
	}

	fnamewav := fname + ".mp3"
	cmd := fmt.Sprintf("-y -i %s %s", fnameout, fnamewav)
	logger.Debug(cmd)
	out, err := exec.Command("ffmpeg", strings.Fields(cmd)...).CombinedOutput()
	logger.Debugf("ffmpeg: %s", out)
	if err != nil {
		err = fmt.Errorf("ffmpeg; %s", err.Error())
		return
	}

	waveformfname := fnamewav + ".png"
	cmd = fmt.Sprintf("-i %s -o %s --background-color ffffff00 --waveform-color ffffff --amplitude-scale 2 --no-axis-labels --pixels-per-second 100 --height 160 --width %2.0f",
		fnamewav, waveformfname, 5.5*100,
	)
	logger.Debug(cmd)
	out, err = exec.Command("audiowaveform", strings.Fields(cmd)...).CombinedOutput()
	if err != nil {
		logger.Errorf("audiowaveform: %s", out)
	}

	return
}
