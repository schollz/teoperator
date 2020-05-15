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
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/schollz/httpfileserver"
	log "github.com/schollz/logger"
	"github.com/schollz/teoperator/src/audiosegment"
	"github.com/schollz/teoperator/src/build"
	"github.com/schollz/teoperator/src/download"
	"github.com/schollz/teoperator/src/op1"
	"github.com/schollz/teoperator/src/utils"
)

func Run(port int) (err error) {
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
	Segments    []int
	OriginalURL string
	StartStop   []float64
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

	uuid, err := generateUserData(audioURL[0], startStop)
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

func generateUserData(u string, startStop []float64) (uuid string, err error) {
	log.Debug(u, startStop)

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
	fname := ""
	uparsed, err := url.Parse(u)
	if err != nil {
		return
	}
	fname = path.Join(pathToData, path.Base(uparsed.Path))

	log.Debugf("downloading to %s", fname)
	err = download.Download(u, fname, 10000000)
	if err != nil {
		return
	}

	// build the drum patch
	fnameShort, numSegments, err := build.DrumpatchFromAudio(fname, startStop)
	if err != nil {
		return
	}

	// no breaks, just get the first 11.75 seconds
	err = audiosegment.Truncate(fname, path.Join(pathToData, fnameShort+"-full.wav"), utils.SecondsToString(startStop[0]), utils.SecondsToString(startStop[0]+11.75))
	if err != nil {
		return
	}
	err = op1.DrumPatch(path.Join(pathToData, fnameShort+"-full.wav"), path.Join(pathToData, fnameShort+"-full.aif"), op1.Default())
	if err != nil {
		return
	}
	err = os.Remove(path.Join(pathToData, fnameShort+"-full.wav"))
	if err != nil {
		return
	}

	// write metadata
	segnums := []int{}
	for i := 0; i < numSegments; i++ {
		segnums = append(segnums, i)
	}
	b, _ := json.Marshal(Metadata{
		Name:        fnameShort,
		UUID:        uuid,
		Segments:    segnums,
		OriginalURL: u,
		StartStop:   startStop,
	})
	err = ioutil.WriteFile(path.Join(pathToData, "metadata.json"), b, 0644)

	return
}
