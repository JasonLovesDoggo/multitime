package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func setupLogging(debug bool) {
	if debug {
		debugLog = log.New(os.Stdout, "DEBUG: ", log.Ltime|log.Lmicroseconds)
	} else {
		debugLog = log.New(io.Discard, "", 0)
	}
}

func forwardHeartbeat(heartbeat []byte, userAgent string, backend Backend) (*http.Response, error) {
	req, err := http.NewRequest("POST", backend.URL+"/api/v1/users/current/heartbeats", bytes.NewReader(heartbeat))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(backend.APIKey))))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent+" (JasonLovesDoggo/multitime)")

	debugLog.Printf("Forwarding to %s", backend.URL+"/heartbeat")
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	return client.Do(req)
}
