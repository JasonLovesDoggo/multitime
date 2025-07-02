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
	req, err := http.NewRequest("POST", backend.URL+"/v1/users/current/heartbeats", bytes.NewReader(heartbeat))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(backend.APIKey))))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent+" (JasonLovesDoggo/multitime)")

	debugLog.Printf("Forwarding heartbeat to %s", backend.URL)
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		debugLog.Printf("Error forwarding heartbeat to %s: %v", backend.URL, err)
		return resp, err
	}
	if resp.StatusCode != 201 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			debugLog.Printf("Error response from %s - Status: %d, Error reading body: %v", backend.URL, resp.StatusCode, err)
		} else {
			debugLog.Printf("Error response from %s - Status: %d, Body: %s", backend.URL, resp.StatusCode, string(body))
		}
		resp.Body.Close()
		// Create a new reader with the same content for the next consumer
		resp.Body = io.NopCloser(bytes.NewBuffer(body))
	}
	return resp, nil
}

func forwardHeartbeats(heartbeat []byte, userAgent string, backend Backend) (*http.Response, error) {
	req, err := http.NewRequest("POST", backend.URL+"/v1/users/current/heartbeats.bulk", bytes.NewReader(heartbeat))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(backend.APIKey))))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent+" (JasonLovesDoggo/multitime)")

	debugLog.Printf("Forwarding bulk heartbeats to %s", backend.URL)
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		debugLog.Printf("Error forwarding bulk heartbeats to %s: %v", backend.URL, err)
		return resp, err
	}
	if resp.StatusCode != 201 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			debugLog.Printf("Error response from %s - Status: %d, Error reading body: %v", backend.URL, resp.StatusCode, err)
		} else {
			debugLog.Printf("Error response from %s - Status: %d, Body: %s", backend.URL, resp.StatusCode, string(body))
		}
		resp.Body.Close()
		// Create a new reader with the same content for the next consumer
		resp.Body = io.NopCloser(bytes.NewBuffer(body))
	}
	return resp, nil
}
