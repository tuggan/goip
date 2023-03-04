package web

import (
	"compress/gzip"
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
	gzipEnabled bool
	templateDir string
	version     string
	branch      string
	date        string
	author      string
	email       string
	server      string
}

func NewHandler(gzipEnabled bool, templateDir, version, branch, date, author, email string) handler {
	return handler{
		gzipEnabled: gzipEnabled,
		templateDir: templateDir,
		version:     version,
		branch:      branch,
		date:        date,
		author:      author,
		email:       email,
		server:      fmt.Sprintf("GoIP %s", version),
	}
}

func (h handler) MainHandler(w http.ResponseWriter, r *http.Request) {

	ip, _, e := net.SplitHostPort(r.RemoteAddr)
	if e != nil {
		h.renderError(w, r, path.Join(h.templateDir, "error"), "Error while parsing host and port", http.StatusInternalServerError)
		logger.Error("[%d] error while parsing host and port %s", http.StatusInternalServerError, r.URL.Path)
		return
	}

	if r.Header.Get("X-Forwarded-For") != "" {
		ip = r.Header.Get("X-Forwarded-For")
	}

	w.Header().Set("Server", h.server)

	s := strings.ToLower(r.URL.Path)

	switch s {
	case "/ip":
		io.WriteString(w, ip)
	case "/user-agent":
		io.WriteString(w, r.Header.Get("User-Agent"))
	case "/host":
		io.WriteString(w, r.Host)
	case "/proto":
		io.WriteString(w, r.Proto)
	case "/accept":
		io.WriteString(w, r.Header.Get("Accept"))
	case "/accept-encoding":
		io.WriteString(w, r.Header.Get("Accept-Encoding"))
	case "/":
		info := []head{
			{"Ip", ip},
			{"User-Agent", r.Header.Get("User-Agent")},
			{"Host", r.Host},
			{"Proto", r.Proto},
			{"Accept", r.Header.Get("Accept")},
			{"Accept-Encoding", r.Header.Get("Accept-Encoding")},
		}
		data := page{
			Title:      "IPConf",
			Clientinfo: info,
			IP:         ip,
			Version:    h.version,
			Branch:     h.branch,
			CommitDate: h.date,
			Author:     h.author,
			Email:      h.email,
		}
		h.renderTemplate(w, r, path.Join(h.templateDir, "index"), data)
	default:
		h.renderError(w, r, path.Join(h.templateDir, "error"), fmt.Sprintf("%s not found", r.URL.Path), http.StatusNotFound)
		logger.Access(r, http.StatusNotFound)
		return
	}
	logger.Access(r, http.StatusOK)
}

func (h handler) renderTemplate(w http.ResponseWriter, r *http.Request, tmpl string, m page) {
	var tw io.Writer = w
	t, _ := template.ParseFiles(tmpl + ".html")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if h.gzipEnabled && strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		w.Header().Set("Content-Encoding", "gzip")
		gw := gzip.NewWriter(w)
		defer gw.Close()
		tw = gw
	}
	t.Execute(tw, m)
}

func (h handler) renderError(w http.ResponseWriter, r *http.Request, tmpl string, s string, code int) {
	var tw io.Writer = w
	p := page{
		Title:   fmt.Sprintf("%d: %s", code, http.StatusText(code)),
		Header:  http.StatusText(code),
		Message: s,
		Code:    strconv.Itoa(code),
	}
	t, _ := template.ParseFiles(tmpl + ".html")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if h.gzipEnabled && strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		w.Header().Set("Content-Encoding", "gzip")
		gw := gzip.NewWriter(w)
		defer gw.Close()
		tw = gw
	}
	w.WriteHeader(code)
	t.Execute(tw, p)
}

func (h handler) GETHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		h.renderError(w, r, path.Join(h.templateDir, "error"), "method not GET", http.StatusBadRequest)
		logger.Error("[Error] [%d] method not GET %s", http.StatusBadRequest, r.URL.Path)
		return
	}

	io.WriteString(w, r.URL.RawQuery)
	logger.Access(r, http.StatusOK)

}

func (h handler) POSTHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		h.renderError(w, r, path.Join(h.templateDir, "error"), "method not POST", http.StatusBadRequest)
		logger.Error("[Error] [%d] method not POST %s", http.StatusBadRequest, r.URL.Path)
		return
	}

	io.Copy(w, r.Body)
	logger.Access(r, http.StatusOK)
}

func (h handler) FaviconHandler(w http.ResponseWriter, r *http.Request) {
	file, err := os.Open(path.Join(h.templateDir, "/favicon.ico"))
	if err != nil {
		h.renderError(w, r, path.Join(h.templateDir, "error"), fmt.Sprintf("Could not find %s", r.URL.Path), http.StatusNotFound)
		logger.Access(r, http.StatusNotFound)
	}
	io.Copy(w, file)
	logger.Access(r, http.StatusOK)
}

func (h handler) RobotsHandler(w http.ResponseWriter, r *http.Request) {
	file, err := os.Open(path.Join(h.templateDir, "/robots.txt"))
	if err != nil {
		h.renderError(w, r, path.Join(h.templateDir, "error"), fmt.Sprintf("Could not find %s", r.URL.Path), http.StatusNotFound)
		logger.Access(r, http.StatusNotFound)
	}
	io.Copy(w, file)
	logger.Access(r, http.StatusOK)
}
