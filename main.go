package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
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
		return
	}

	// Placed here for *a bit* better performance
	if r.URL.Path == "/Ip" {
		ip, _, e := net.SplitHostPort(r.RemoteAddr)
		if e == nil {
			io.WriteString(w, ip)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			renderError(w, "Error while parsing ip", http.StatusInternalServerError)
		}
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
			log.Printf("Duplicate keys! %s", key)
		} else {
			info[key] = strings.Join(val, "\n")
		}
	}

	for key, val := range r.Trailer {
		if _, ok := info[key]; ok {
			log.Printf("Duplicate keys! %s", key)
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
	} else if c, ok := info[r.URL.Path[1:]]; ok {
		io.WriteString(w, c)
	} else {
		renderError(w, fmt.Sprintf("Could not find %s", r.URL.Path[1:]), http.StatusNotFound)
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

func main() {
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
