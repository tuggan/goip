package web

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"text/template"

	"github.com/tuggan/goip/logger"
)

type head struct {
	Key string
	Val string
}

type page struct {
	Title      string
	Clientinfo []head
	Header     string
	Message    string
	Code       string
	IP         string
	Version    string
	Branch     string
	CommitDate string
	Author     string
	Email      string
}

type handler struct {
	templateDir string
	version     string
	branch      string
	date        string
	author      string
	email       string
}

func NewHandler(templateDir, version, branch, date, author, email string) handler {
	return handler{
		templateDir: templateDir,
		version:     version,
		branch:      branch,
		date:        date,
		author:      author,
		email:       email,
	}
}

func (h handler) MainHandler(w http.ResponseWriter, r *http.Request) {

	ip, port, e := net.SplitHostPort(r.RemoteAddr)
	if e != nil {
		renderError(w, path.Join(h.templateDir, "error"), "Error while parsing host and port", http.StatusInternalServerError)
		logger.Error("[%d] error while parsing host and port %s", http.StatusInternalServerError, r.URL.Path)
		return
	}

	// Placed here for *a bit* better performance
	if r.URL.Path == "/ip" {
		io.WriteString(w, ip)
		logger.Access(r, http.StatusOK)
		return
	}

	s := strings.ToLower(r.URL.Path)

	switch s {
	case "/ip":
		io.WriteString(w, ip)
	case "/port":
		io.WriteString(w, port)
	case "/user-agent":
		io.WriteString(w, r.Header.Get("User-Agent"))
	case "/method":
		io.WriteString(w, r.Method)
	case "/host":
		io.WriteString(w, r.Host)
	case "/proto":
		io.WriteString(w, r.Proto)
	case "/content-length":
		io.WriteString(w, strconv.FormatInt(r.ContentLength, 10))
	case "/accept":
		io.WriteString(w, r.Header.Get("Accept"))
	case "/accept-encoding":
		io.WriteString(w, r.Header.Get("Accept-Encoding"))
	case "/":
		info := []head{
			{"Ip", ip},
			{"Port", port},
			{"User-Agent", r.Header.Get("User-Agent")},
			{"Method", r.Method},
			{"Host", r.Host},
			{"Proto", r.Proto},
			{"Content-Length", strconv.FormatInt(r.ContentLength, 10)},
			{"Accept", r.Header.Get("Accept")},
			{"Accept-Encoding", r.Header.Get("Accept-Encoding")},
		}
		data := page{
			Title:      "Index",
			Clientinfo: info,
			IP:         ip,
			Version:    h.version,
			Branch:     h.branch,
			CommitDate: h.date,
			Author:     h.author,
			Email:      h.email,
		}
		renderTemplate(w, path.Join(h.templateDir, "index"), data)
	default:
		renderError(w, path.Join(h.templateDir, "error"), fmt.Sprintf("Could not find %s", r.URL.Path), http.StatusNotFound)
		logger.Access(r, http.StatusNotFound)
		return
	}
	logger.Access(r, http.StatusOK)
}

func renderTemplate(w http.ResponseWriter, tmpl string, m page) {
	t, _ := template.ParseFiles(tmpl + ".html")
	t.Execute(w, m)
}

func renderError(w http.ResponseWriter, tmpl string, s string, code int) {
	p := page{
		Title:   strconv.Itoa(code),
		Header:  http.StatusText(code),
		Message: s,
		Code:    strconv.Itoa(code),
	}
	w.WriteHeader(code)
	t, _ := template.ParseFiles(tmpl + ".html")
	t.Execute(w, p)
}

func (h handler) GETHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		renderError(w, path.Join(h.templateDir, "error"), "method not GET", http.StatusBadRequest)
		logger.Error("[Error] [%d] method not GET %s", http.StatusBadRequest, r.URL.Path)
		return
	}

	io.WriteString(w, r.URL.RawQuery)
	logger.Access(r, http.StatusOK)

}

func (h handler) POSTHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		renderError(w, path.Join(h.templateDir, "error"), "method not POST", http.StatusBadRequest)
		logger.Error("[Error] [%d] method not POST %s", http.StatusBadRequest, r.URL.Path)
		return
	}

	io.Copy(w, r.Body)
	logger.Access(r, http.StatusOK)
}

func (h handler) FaviconHandler(w http.ResponseWriter, r *http.Request) {
	file, err := os.Open(path.Join(h.templateDir, "/favicon.ico"))
	if err != nil {
		renderError(w, path.Join(h.templateDir, "error"), fmt.Sprintf("Could not find %s", r.URL.Path), http.StatusNotFound)
		logger.Access(r, http.StatusNotFound)
	}
	io.Copy(w, file)
	logger.Access(r, http.StatusOK)
}

func (h handler) RobotsHandler(w http.ResponseWriter, r *http.Request) {
	file, err := os.Open(path.Join(h.templateDir, "/robots.txt"))
	if err != nil {
		renderError(w, path.Join(h.templateDir, "error"), fmt.Sprintf("Could not find %s", r.URL.Path), http.StatusNotFound)
		logger.Access(r, http.StatusNotFound)
	}
	io.Copy(w, file)
	logger.Access(r, http.StatusOK)
}
