package server

import (
	"crypto/md5"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/schollz/httpfileserver"
	log "github.com/schollz/logger"
)

func Run(port int) (err error) {
	loadTemplates()
	log.Infof("listening on :%d", port)
	http.HandleFunc("/static/", httpfileserver.New("/static/", "static/", httpfileserver.OptionNoCache(true)).Handle())
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

type Render struct {
	PianoCount   string
	CountryCount string
	Heading      string
	Title        string
	MessageError string
	MessageInfo  string
	BreadCrumbs  []Href
	Links        []Href
	RefererURL   string
	Description  string
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

	log.Debug(audioURL, startStop)

	uuid := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprint("%+v %+v", audioURL, startStop))))
	log.Debug(uuid)
	t["main"].Execute(w, Render{})
	err = nil
	return
}

func viewMain(w http.ResponseWriter, r *http.Request, messageError string, templateName string) (err error) {

	t[templateName].Execute(w, Render{
		Title:        "Pianos for Travelers",
		MessageError: messageError,
	})
	return
}
