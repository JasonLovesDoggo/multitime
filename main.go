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

	http.HandleFunc("/api/v1/users/current/heartbeats", handleHeartbeat)
	http.HandleFunc("/api/v1/users/current/status_bar/today", handleStatusBar)

	log.Printf("Starting MultiTime server on port %d", config.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config.Port), nil))
}
