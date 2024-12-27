package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

var (
	debugLog *log.Logger
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Usage: multitime <config_file>")
	}

	var err error
	config, err = loadConfig(os.Args[1])
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	setupLogging(config.Debug)

	http.HandleFunc("/users/current/heartbeats", handleHeartbeat)
	http.HandleFunc("/users/current/heartbeats.bulk", handleHeartbeatsBulk)
	http.HandleFunc("/users/current/statusbar/today", handleStatusBar)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// The "/" matches anything not handled elsewhere. If it's not the root
		// then report not found.

		debugLog.Printf("404 Not Found: %s", r.URL.Path)
		http.NotFound(w, r)
	})
	log.Printf("Starting MultiTime server on port %d", config.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config.Port), nil))
}
