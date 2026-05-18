package web

import (
	"compress/gzip"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"html/template"

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
	gzipEnabled   bool
	templateDir   string
	version       string
	branch        string
	date          string
	author        string
	email         string
	server        string
	trustedIPNets []*net.IPNet
}

func NewHandler(gzipEnabled bool, templateDir, version, branch, date, author, email string, trustedProxies []string) handler {
	h := handler{
		gzipEnabled: gzipEnabled,
		templateDir: templateDir,
		version:     version,
		branch:      branch,
		date:        date,
		author:      author,
		email:       email,
		server:      fmt.Sprintf("GoIP %s", version),
	}

	for _, p := range trustedProxies {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		// Try as CIDR notation first (e.g. "10.0.0.0/8", "192.168.1.0/24")
		_, cidr, err := net.ParseCIDR(p)
		if err == nil {
			h.trustedIPNets = append(h.trustedIPNets, cidr)
			continue
		}
		// Try as a plain IP, convert to /32 or /128
		ip := net.ParseIP(p)
		if ip != nil {
			if ip.To4() != nil {
				_, cidr, _ = net.ParseCIDR(ip.String() + "/32")
			} else {
				_, cidr, _ = net.ParseCIDR(ip.String() + "/128")
			}
			h.trustedIPNets = append(h.trustedIPNets, cidr)
			continue
		}
		logger.Warning("Ignoring invalid trusted proxy entry: %q", p)
	}

	return h
}

// isTrustedProxy checks whether the remote address (host:port) matches
// any of the configured trusted proxy IPs or CIDR ranges.
func (h handler) isTrustedProxy(remoteAddr string) bool {
	if len(h.trustedIPNets) == 0 {
		return false
	}
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return false
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	for _, cidr := range h.trustedIPNets {
		if cidr.Contains(ip) {
			return true
		}
	}
	return false
}

func (h handler) MainHandler(w http.ResponseWriter, r *http.Request) {

	ip, _, e := net.SplitHostPort(r.RemoteAddr)
	if e != nil {
		h.renderError(w, r, path.Join(h.templateDir, "error"), "Error while parsing host and port", http.StatusInternalServerError)
		logger.Error("[%d] error while parsing host and port %s", http.StatusInternalServerError, r.URL.Path)
		return
	}

	// Only trust X-Forwarded-For when the connection comes from a
	// configured trusted proxy. This prevents direct clients from
	// spoofing their IP address via the header.
	if r.Header.Get("X-Forwarded-For") != "" && h.isTrustedProxy(r.RemoteAddr) {
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

// safeTemplatePath validates that the given path stays within the configured
// template directory to prevent directory traversal attacks.
func (h handler) safeTemplatePath(fullPath string) (string, error) {
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return "", err
	}
	absDir, err := filepath.Abs(h.templateDir)
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(absPath, absDir) {
		return "", fmt.Errorf("template path %q escapes template directory %q", absPath, absDir)
	}
	return fullPath, nil
}

func (h handler) renderTemplate(w http.ResponseWriter, r *http.Request, tmpl string, m page) {
	safeTmpl, err := h.safeTemplatePath(tmpl)
	if err != nil {
		logger.Error("Template path validation failed: %v", err)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "500 Internal Server Error")
		return
	}
	var tw io.Writer = w
	t, err := template.ParseFiles(safeTmpl + ".html")
	if err != nil {
		logger.Error("Failed to parse template %s: %v", tmpl, err)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "500 Internal Server Error")
		return
	}
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
	safeTmpl, err := h.safeTemplatePath(tmpl)
	if err != nil {
		logger.Error("Template path validation failed: %v", err)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "500 Internal Server Error")
		return
	}
	var tw io.Writer = w
	p := page{
		Title:   fmt.Sprintf("%d: %s", code, http.StatusText(code)),
		Header:  http.StatusText(code),
		Message: s,
		Code:    strconv.Itoa(code),
	}
	t, err := template.ParseFiles(safeTmpl + ".html")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if h.gzipEnabled && strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		w.Header().Set("Content-Encoding", "gzip")
		gw := gzip.NewWriter(w)
		defer gw.Close()
		tw = gw
	}
	w.WriteHeader(code)
	if err != nil {
		logger.Error("Failed to parse error template %s: %v", tmpl, err)
		fmt.Fprintf(tw, "<h1>%d: %s</h1><p>%s</p>", code, http.StatusText(code), s)
		return
	}
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

func (h handler) FaviconHandler(w http.ResponseWriter, r *http.Request) {
	favPath := path.Join(h.templateDir, "/favicon.ico")
	if _, err := h.safeTemplatePath(favPath); err != nil {
		logger.Error("Favicon path validation failed: %v", err)
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	file, err := os.Open(favPath)
	if err != nil {
		h.renderError(w, r, path.Join(h.templateDir, "error"), fmt.Sprintf("Could not find %s", r.URL.Path), http.StatusNotFound)
		logger.Access(r, http.StatusNotFound)
		return
	}
	defer file.Close()
	io.Copy(w, file)
	logger.Access(r, http.StatusOK)
}

func (h handler) RobotsHandler(w http.ResponseWriter, r *http.Request) {
	robPath := path.Join(h.templateDir, "/robots.txt")
	if _, err := h.safeTemplatePath(robPath); err != nil {
		logger.Error("Robots path validation failed: %v", err)
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	file, err := os.Open(robPath)
	if err != nil {
		h.renderError(w, r, path.Join(h.templateDir, "error"), fmt.Sprintf("Could not find %s", r.URL.Path), http.StatusNotFound)
		logger.Access(r, http.StatusNotFound)
		return
	}
	defer file.Close()
	io.Copy(w, file)
	logger.Access(r, http.StatusOK)
}
