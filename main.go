package main

import (
	"fmt"
	"io/ioutil"
	"log"
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

	var t string
	if viper.IsSet("templateDir") {
		t = viper.GetString("templateDir")
	} else {
		t = "html"
	}

	h := web.NewHandler(t)

	logger.Info("Listening on %s", addr)
	http.HandleFunc("/", h.MainHandler)
	http.HandleFunc("/GET", h.GETHandler)
	http.HandleFunc("/POST", h.POSTHandler)
	http.HandleFunc("/favicon.ico", h.FaviconHandler)
	log.Fatal(http.ListenAndServe(addr, nil))
}
