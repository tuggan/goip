package main

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/tuggan/goip/logger"
)

var (
	version string
	date    string
	branch  string
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

func handleFavicon(w http.ResponseWriter, r *http.Request) {
	file, err := os.Open("html/favicon.ico")
	if err != nil {
		renderError(w, fmt.Sprintf("Could not find %s", r.URL.Path), http.StatusNotFound)
		logger.Access(r, http.StatusNotFound)
	}
	io.Copy(w, file)
	logger.Access(r, http.StatusOK)
}

func printVersion() {
	fmt.Printf("GoIP %s (%s) branch %s © Dennis Vesterlund <dennisvesterlund@gmail.com>\n", version, date, branch)
}

func printHelp() {
	fmt.Printf("Usage:\n\n")
	fmt.Printf("\t %s [arguments]\n\n", os.Args[0])
	fmt.Printf("Arguments:\n\n")
	pflag.PrintDefaults()
}

func main() {

	pflag.StringP("address", "a", "127.0.0.1", "Address to bind the server to")
	pflag.Uint16P("port", "p", 3000, "port to listen on")
	version := pflag.BoolP("version", "v", false, "Print version and exit")
	help := pflag.BoolP("help", "h", false, "Print help and exit")
	configFile := pflag.StringP("config", "c", ".", "Path to config file")

	pflag.Parse()

	viper.BindPFlags(pflag.CommandLine)

	if *help {
		printHelp()
		os.Exit(0)
	}

	if *version {
		printVersion()
		os.Exit(0)
	}

	viper.SetConfigName("goip")
	viper.AddConfigPath(*configFile)
	viper.AddConfigPath("/etc/GoIP/")
	viper.AddConfigPath("$HOME/.GoIP/")
	viper.AddConfigPath("config/")

	err := viper.ReadInConfig()
	if err != nil {
		logger.Error("Error with config file: %s", err)
	}

	logger.Init(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)

	logger.Info("User: %s", viper.GetString("user"))

	addr := viper.GetString("address") + ":" + strconv.Itoa(viper.GetInt("port"))

	logger.Info("Starting %s", os.Args[0])
	if 0 < len(os.Args[1:]) {
		logger.Info("Arguments: %s", os.Args[1:])
	}

	fmt.Println(viper.GetString("address"))

	logger.Info("Listening on %s", addr)
	http.HandleFunc("/", handler)
	http.HandleFunc("/GET", handleGET)
	http.HandleFunc("/POST", handlePOST)
	http.HandleFunc("/favicon.ico", handleFavicon)
	log.Fatal(http.ListenAndServe(addr, nil))
}
