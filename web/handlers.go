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

type page struct {
	Title      string
	Clientinfo map[string]string
	Header     string
	Message    string
	Code       string
}

type handler struct {
	templateDir string
}

func NewHandler(templateDir string) handler {
	return handler{templateDir: templateDir}
}

func (h handler) MainHandler(w http.ResponseWriter, r *http.Request) {

	ip, port, e := net.SplitHostPort(r.RemoteAddr)
	if e != nil {
		renderError(w, path.Join(h.templateDir, "error"), "Error while parsing host and port", http.StatusInternalServerError)
		logger.Error("[%d] error while parsing host and port %s", http.StatusInternalServerError, r.URL.Path)
		return
	}

	// Placed here for *a bit* better performance
	if r.URL.Path == "/Ip" {
		io.WriteString(w, ip)
		logger.Access(r, http.StatusOK)
		return
	}

	info := map[string]string{
		"Ip":             ip,
		"Port":           port,
		"Method":         r.Method,
		"Host":           r.Host,
		"Proto":          r.Proto,
		"Content-Length": strconv.FormatInt(r.ContentLength, 10),
	}

	for key, val := range r.Header {
		if _, ok := info[key]; ok {
			logger.Error("[Error] [-] duplicate keys! %s", key)
		} else {
			info[key] = strings.Join(val, "\n")
		}
	}

	for key, val := range r.Trailer {
		if _, ok := info[key]; ok {
			logger.Error("[Error] [-] duplicate keys! %s", key)
		} else {
			info[key] = strings.Join(val, "\n")
		}
	}

	if r.URL.Path == "/" {
		data := page{
			Title:      "Index",
			Clientinfo: info,
		}
		renderTemplate(w, path.Join(h.templateDir, "index"), data)
		logger.Access(r, http.StatusOK)
	} else if c, ok := info[r.URL.Path[1:]]; ok {
		io.WriteString(w, c)
		logger.Access(r, http.StatusOK)
	} else {
		renderError(w, path.Join(h.templateDir, "error"), fmt.Sprintf("Could not find %s", r.URL.Path), http.StatusNotFound)
		logger.Access(r, http.StatusNotFound)
	}
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
	file, err := os.Open("html/favicon.ico")
	if err != nil {
		renderError(w, path.Join(h.templateDir, "error"), fmt.Sprintf("Could not find %s", r.URL.Path), http.StatusNotFound)
		logger.Access(r, http.StatusNotFound)
	}
	io.Copy(w, file)
	logger.Access(r, http.StatusOK)
}
