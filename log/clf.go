package log

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	// reference: https://httpd.apache.org/docs/2.2/logs.html#combined + execution time.
	apacheCombinedFormatPattern = "%s - - [%s] \"%s %s %s\" %d %d \"%s\" \"%s\" %.4f\n"
)

// CLFLogRecord common log format record
type CLFLogRecord struct {
	http.ResponseWriter
	ip                    string
	time                  time.Time
	method, uri, protocol string
	status                int
	responseBytes         int64
	referer               string
	userAgent             string
	elapsedTime           time.Duration
}

// Log
func (r *CLFLogRecord) Log(out io.Writer) {
	timeFormatted := r.time.Format("02/Jan/2006 03:04:05")
	fmt.Fprintf(out, apacheCombinedFormatPattern, r.ip, timeFormatted, r.method,
		r.uri, r.protocol, r.status, r.responseBytes, r.referer, r.userAgent,
		r.elapsedTime.Seconds())
}

// Write implement Write method with interface to io.Writer
func (r *CLFLogRecord) Write(p []byte) (int, error) {
	written, err := r.ResponseWriter.Write(p)
	r.responseBytes += int64(written)
	return written, err
}

// WriteHeader implement WriteHeader with interface to http.ResponseWriter
func (r *CLFLogRecord) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

type CLFLoggingHandler struct {
	handler http.Handler
	out     io.Writer
}

func NewCLFLoggingHandler(handler http.Handler, out io.Writer) http.Handler {
	return &CLFLoggingHandler{
		handler: handler,
		out:     out,
	}
}

// ServeHTTP overwrite ServerHTTP
func (h *CLFLoggingHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	clientIP := r.RemoteAddr
	if colon := strings.LastIndex(clientIP, ":"); colon != -1 {
		clientIP = clientIP[:colon]
	}

	referer := r.Referer()
	if referer == "" {
		referer = "-"
	}

	userAgent := r.UserAgent()
	if userAgent == "" {
		userAgent = "-"
	}

	record := &CLFLogRecord{
		ResponseWriter: rw,
		ip:             clientIP,
		time:           time.Time{},
		method:         r.Method,
		uri:            r.RequestURI,
		protocol:       r.Proto,
		status:         http.StatusOK,
		referer:        referer,
		userAgent:      userAgent,
		elapsedTime:    time.Duration(0),
	}

	startTime := time.Now()
	h.handler.ServeHTTP(record, r)
	finishTime := time.Now()

	record.time = finishTime.UTC()
	record.elapsedTime = finishTime.Sub(startTime)

	record.Log(h.out)
}
