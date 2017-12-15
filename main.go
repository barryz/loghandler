package main

import (
	"net/http"
	"os"
	"time"

	"github.com/barryz/loghandler/log"
)

func main() {
	mux := http.DefaultServeMux
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello world"))
		// For test, sleep 0.5 seconds
		time.Sleep(500 * time.Millisecond)
		return
	})
	loggingHandler := log.NewCLFLoggingHandler(mux, os.Stdout)
	server := &http.Server{
		Addr:    ":8000",
		Handler: loggingHandler,
	}
	server.ListenAndServe()
}
