package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/tuggan/goip/logger"
	"github.com/tuggan/goip/web"
)

var (
	Version string
	Date    string
	Branch  string
	author  = "Dennis Vesterlund"
	email   = "dennisvesterlund@gmail.com"
)

func printVersion() {
	fmt.Printf("GoIP %s (%s) branch %s © Dennis Vesterlund <dennisvesterlund@gmail.com>\n", Version, Date, Branch)
}

func printHelp() {
	fmt.Printf("Usage:\n\n")
	fmt.Printf("\t %s [arguments]\n\n", os.Args[0])
	fmt.Printf("Arguments:\n\n")
	pflag.PrintDefaults()
}

func serve(wg *sync.WaitGroup, srv *http.Server, l net.Listener, certFile, keyFile string) {
	defer wg.Done()

	if certFile != "" && keyFile != "" {
		if err := srv.ServeTLS(l, certFile, keyFile); err != http.ErrServerClosed {
			log.Fatalf("HTTP server ServeTLS: %v", err)
		}
	} else {
		if err := srv.Serve(l); err != http.ErrServerClosed {
			log.Fatalf("HTTP server Serve: %v", err)
		}
	}
	logger.Info("Shutting down serve routine")
}

// recoveryMiddleware wraps an http.Handler with panic recovery.
// If a handler panics, it logs the error and returns a 500 response
// instead of crashing the server process.
func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				logger.Error("Panic recovered: %v", rec)
				http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func main() {

	pflag.StringP("endpoint", "e", "127.0.0.1:3000", "Endpoint to listen on")
	pflag.String("tlsEndpoint", "127.0.0.1:3000", "Endpoint to listen on")
	pflag.String("tlsCert", "", "Path to TLS Certificate file")
	pflag.String("tlsKey", "", "Path to TLS Key file")

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
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.goip/")
	viper.AddConfigPath("/etc/goip/")
	viper.AddConfigPath("config/")

	err := viper.ReadInConfig()
	if err != nil {
		logger.Error("Error with config file: %s", err)
	}

	addr := viper.GetStringSlice("endpoint")

	logger.Info("Starting %s", os.Args[0])
	if 0 < len(os.Args[1:]) {
		logger.Info("Arguments: %s", os.Args[1:])
	}

	var t = "html"
	if viper.IsSet("templateDir") {
		t = viper.GetString("templateDir")
	}

	var egzip = true
	if viper.IsSet("enablegzip") {
		egzip = viper.GetBool("enablegzip")
	}

	var tlsEndpoint []string
	if viper.IsSet("tlsEndpoint") {
		tlsEndpoint = viper.GetStringSlice("tlsEndpoint")
	}

	var tlsCert = ""
	if viper.IsSet("tlsCert") {
		tlsCert = viper.GetString("tlsCert")
	}

	var tlsKey = ""
	if viper.IsSet("tlsKey") {
		tlsKey = viper.GetString("tlsKey")
	}

	handler := http.NewServeMux()

	h := web.NewHandler(egzip, t, Version, Branch, Date, author, email)
	handler.HandleFunc("/", h.MainHandler)
	handler.HandleFunc("/GET", h.GETHandler)
	handler.HandleFunc("/favicon.ico", h.FaviconHandler)
	handler.HandleFunc("/robots.txt", h.RobotsHandler)

	// Wrap the mux with panic recovery middleware.
	wrappedHandler := recoveryMiddleware(handler)

	// Separate server instances for TLS and plain HTTP.
	// Go's http.Server docs state Serve/ServeTLS must not be called
	// concurrently on the same server.
	var plainSrv http.Server
	plainSrv.Handler = wrappedHandler
	plainSrv.ReadTimeout = 10 * time.Second
	plainSrv.ReadHeaderTimeout = 5 * time.Second
	plainSrv.WriteTimeout = 10 * time.Second
	plainSrv.IdleTimeout = 60 * time.Second
	plainSrv.MaxHeaderBytes = 1 << 20 // 1 MB

	var tlsSrv http.Server
	tlsSrv.Handler = wrappedHandler
	tlsSrv.ReadTimeout = 10 * time.Second
	tlsSrv.ReadHeaderTimeout = 5 * time.Second
	tlsSrv.WriteTimeout = 10 * time.Second
	tlsSrv.IdleTimeout = 60 * time.Second
	tlsSrv.MaxHeaderBytes = 1 << 20 // 1 MB

	var wg sync.WaitGroup

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		logger.Info("Shutting down server")

		shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownRelease()

		// Shutdown both servers. Calling Shutdown on a server that
		// never had Serve called is a safe no-op.
		if err := plainSrv.Shutdown(shutdownCtx); err != nil {
			log.Printf("Plain HTTP server Shutdown: %v", err)
		}
		if err := tlsSrv.Shutdown(shutdownCtx); err != nil {
			log.Printf("TLS server Shutdown: %v", err)
		}
	}()

	if len(tlsEndpoint) > 0 {
		if tlsKey == "" || tlsCert == "" {
			log.Fatal("Both certFile and keyFile must be set")
		}

		for _, e := range tlsEndpoint {
			tlsListener, err := net.Listen("tcp", e)
			if err != nil {
				logger.Error("Error binding listening socket: %s", err)
				os.Exit(1)
			}
			logger.Info("Starting HTTPS server on https://%s", e)

			wg.Add(1)
			go serve(&wg, &tlsSrv, tlsListener, tlsCert, tlsKey)
		}
	}

	for _, e := range addr {
		listener, err := net.Listen("tcp", e)
		if err != nil {
			logger.Error("Error binding listening socket: %s", err)
			os.Exit(1)
		}
		logger.Info("Starting HTTP server on http://%s", e)
		wg.Add(1)
		go serve(&wg, &plainSrv, listener, "", "")
	}

	logger.Info("Waiting for waitgroups")
	wg.Wait()
	logger.Info("Shutting down GoIP server")
}
