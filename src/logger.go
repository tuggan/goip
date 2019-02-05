package logger

import (
	"io"
	"log"
	"net/http"
)

var logger struct {
	trace   *log.Logger
	info    *log.Logger
	warning *log.Logger
	err     *log.Logger
}

func Init(tH io.Writer, iH io.Writer, wH io.Writer, eH io.Writer) {
	logger.trace = log.New(tH, "", log.Ldate|log.Ltime|log.Lmicroseconds)
	logger.info = log.New(iH, "", log.Ldate|log.Ltime|log.Lmicroseconds)
	logger.warning = log.New(wH, "", log.Ldate|log.Ltime|log.Lmicroseconds)
	logger.err = log.New(eH, "", log.Ldate|log.Ltime|log.Lmicroseconds)
}

func Trace(format string, v ...interface{}) {
	logger.trace.Printf(format, v...)
}

func Info(format string, v ...interface{}) {
	logger.info.Printf(format, v...)
}

func Warning(format string, v ...interface{}) {
	logger.warning.Printf(format, v...)
}

func Error(format string, v ...interface{}) {
	logger.err.Printf(format, v...)
}

func Access(r *http.Request, status int) {
	logger.info.Printf("[Info] [%d] %s: %s", status, r.RemoteAddr, r.URL.Path)
}
