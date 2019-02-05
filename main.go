package main

import (
	"flag"
	"fmt"
	"goip/src"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type page struct {
	Title      string
	Clientinfo map[string]string
	Header     string
	Message    string
	Code       string
}

func handler(w http.ResponseWriter, r *http.Request) {

	ip, port, e := net.SplitHostPort(r.RemoteAddr)
	if e != nil {
		renderError(w, "Error while parsing host and port", http.StatusInternalServerError)
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
		renderTemplate(w, "html/index", data)
		logger.Access(r, http.StatusOK)
	} else if c, ok := info[r.URL.Path[1:]]; ok {
		io.WriteString(w, c)
		logger.Access(r, http.StatusOK)
	} else {
		renderError(w, fmt.Sprintf("Could not find %s", r.URL.Path), http.StatusNotFound)
		logger.Access(r, http.StatusNotFound)
	}
}

func renderTemplate(w http.ResponseWriter, tmpl string, m page) {
	t, _ := template.ParseFiles(tmpl + ".html")
	t.Execute(w, m)
}

func renderError(w http.ResponseWriter, s string, code int) {
	p := page{
		Title:   strconv.Itoa(code),
		Header:  http.StatusText(code),
		Message: s,
		Code:    strconv.Itoa(code),
	}
	w.WriteHeader(code)
	t, _ := template.ParseFiles("html/error.html")
	t.Execute(w, p)
}

func handleGET(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		renderError(w, "method not GET", http.StatusBadRequest)
		logger.Error("[Error] [%d] method not GET %s", http.StatusBadRequest, r.URL.Path)
		return
	}

	io.WriteString(w, r.URL.RawQuery)
	logger.Access(r, http.StatusOK)

}

func handlePOST(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		renderError(w, "method not POST", http.StatusBadRequest)
		logger.Error("[Error] [%d] method not POST %s", http.StatusBadRequest, r.URL.Path)
		return
	}

	io.Copy(w, r.Body)
	logger.Access(r, http.StatusOK)
}

func main() {

	bindAddr := flag.String("bind-address", "127.0.0.1", "Address to bind the server to")
	bindPort := flag.String("port", "3000", "port to listen on")

	flag.Parse()

	addr := *bindAddr + ":" + *bindPort

	logger.Init(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)

	logger.Info("Starting %s", os.Args[0])
	if 0 < len(os.Args[1:]) {
		logger.Info("Arguments: %s", os.Args[1:])
	}
	logger.Info("Listening on %s", addr)
	http.HandleFunc("/", handler)
	http.HandleFunc("/GET", handleGET)
	http.HandleFunc("/POST", handlePOST)
	log.Fatal(http.ListenAndServe(addr, nil))
}
