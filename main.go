package main

import (
	"flag"
	"fmt"
	"html/template"
	"io"
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
		log.Printf("[Error] [%d] error while parsing host and port %s", http.StatusInternalServerError, r.URL.Path[1:])
		return
	}

	// Placed here for *a bit* better performance
	if r.URL.Path == "/Ip" {
		io.WriteString(w, ip)
		log.Printf("[Info] [%d] %s: %s", http.StatusOK, ip, r.URL.Path[1:])
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
			log.Printf("[Error] [-] duplicate keys! %s", key)
		} else {
			info[key] = strings.Join(val, "\n")
		}
	}

	for key, val := range r.Trailer {
		if _, ok := info[key]; ok {
			log.Printf("[Error] [-] duplicate keys! %s", key)
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
		log.Printf("[Info] [%d] %s: %s", http.StatusOK, ip, r.URL.Path[1:])
	} else if c, ok := info[r.URL.Path[1:]]; ok {
		io.WriteString(w, c)
		log.Printf("[Info] [%d] %s: %s", http.StatusOK, ip, r.URL.Path[1:])
	} else {
		renderError(w, fmt.Sprintf("Could not find %s", r.URL.Path[1:]), http.StatusNotFound)
		log.Printf("[Error] [%d] could not find %s", http.StatusNotFound, r.URL.Path[1:])
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

	bindAddr := flag.String("bind-address", "127.0.0.1", "Address to bind the server to")
	bindPort := flag.String("port", "3000", "port to listen on")

	flag.Parse()

	addr := *bindAddr + ":" + *bindPort

	f, err := os.OpenFile("access.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, os.FileMode(0666))
	if err != nil {
		log.Fatal("Coulnd not open file", err.Error())
	}

	log.SetOutput(f)

	log.Printf("Starting %s", os.Args[0])
	if 0 < len(os.Args[1:]) {
		log.Printf("Arguments: %s", os.Args[1:])
	}
	log.Printf("Listening on %s", addr)
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(addr, nil))
}
