package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/tuggan/goip/logger"
	"github.com/tuggan/goip/web"
)

var (
	version string
	date    string
	branch  string
	author  string = "Dennis Vesterlund"
	email   string = "dennisvesterlund@gmail.com"
)

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
	versionFlag := pflag.BoolP("version", "v", false, "Print version and exit")
	help := pflag.BoolP("help", "h", false, "Print help and exit")
	configFile := pflag.StringP("config", "c", ".", "Path to config file")

	pflag.Parse()

	viper.BindPFlags(pflag.CommandLine)

	if *help {
		printHelp()
		os.Exit(0)
	}

	if *versionFlag {
		printVersion()
		os.Exit(0)
	}

	logger.Init(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)

	viper.SetConfigName("goip")
	viper.AddConfigPath(*configFile)
	viper.AddConfigPath("/etc/GoIP/")
	viper.AddConfigPath("$HOME/.GoIP/")
	viper.AddConfigPath("config/")

	err := viper.ReadInConfig()
	if err != nil {
		logger.Error("Error with config file: %s", err)
	}

	addr := viper.GetString("address") + ":" + strconv.Itoa(viper.GetInt("port"))

	logger.Info("Starting %s", os.Args[0])
	if 0 < len(os.Args[1:]) {
		logger.Info("Arguments: %s", os.Args[1:])
	}

	var t string
	if viper.IsSet("templateDir") {
		t = viper.GetString("templateDir")
	} else {
		t = "html"
	}

	s, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Error("Error binding listening socket: %s", err)
		os.Exit(1)
	}

	h := web.NewHandler(t, version, branch, date, author, email)

	logger.Info("Listening on %s", addr)
	http.HandleFunc("/", h.MainHandler)
	http.HandleFunc("/GET", h.GETHandler)
	http.HandleFunc("/POST", h.POSTHandler)
	http.HandleFunc("/favicon.ico", h.FaviconHandler)
	log.Fatal(http.Serve(s, nil))
}
