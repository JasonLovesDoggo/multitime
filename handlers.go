package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

func handleStatusBar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var primaryResp *http.Response
	var primaryErr error

	// Find primary backend
	var primaryBackend Backend
	for _, b := range config.Backends {
		if b.IsPrimary {
			primaryBackend = b
			break
		}
	}

	// Forward to primary backend only since this is a GET request
	req, err := http.NewRequest("GET", primaryBackend.URL+"/api/v1/users/current/status_bar/today", nil)
	if err != nil {
		http.Error(w, "Error creating request", http.StatusInternalServerError)
		return
	}

	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(primaryBackend.APIKey))))
	req.Header.Set("User-Agent", r.UserAgent()+" (JasonLovesDoggo/multitime)")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	primaryResp, primaryErr = client.Do(req)

	if primaryErr != nil {
		debugLog.Printf("Primary backend error: %v", primaryErr)
		// Return empty response as specified in the API docs
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"data":{"grand_total":{"decimal":"","digital":"","hours":0,"minutes":0,"text":"","total_seconds":0},"categories":[],"dependencies":[],"editors":[],"languages":[],"machines":[],"operating_systems":[],"projects":[],"range":{"text":"Today","timezone":"UTC"}}}`)
		return
	}

	defer primaryResp.Body.Close()

	// Copy headers from primary response
	for key, values := range primaryResp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Copy status code and body
	w.WriteHeader(primaryResp.StatusCode)
	if _, err := io.Copy(w, primaryResp.Body); err != nil {
		debugLog.Printf("Error copying response body: %v", err)
	}
}

func handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	heartbeat, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Validate JSON
	if !json.Valid(heartbeat) {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	debugLog.Printf("Received heartbeat: %s", string(heartbeat))

	var wg sync.WaitGroup
	var primaryResp *http.Response
	var primaryErr error
	respChan := make(chan struct {
		resp    *http.Response
		err     error
		backend Backend
	}, len(config.Backends))

	// Forward to all backends concurrently
	for _, backend := range config.Backends {
		wg.Add(1)
		go func(b Backend) {
			defer wg.Done()
			resp, err := forwardHeartbeat(heartbeat, r.UserAgent(), b)
			respChan <- struct {
				resp    *http.Response
				err     error
				backend Backend
			}{resp, err, b}
		}(backend)
	}

	// Close channel when all goroutines complete
	go func() {
		wg.Wait()
		close(respChan)
	}()

	// Collect responses
	for result := range respChan {
		if result.backend.IsPrimary {
			primaryResp = result.resp
			primaryErr = result.err
		} else if result.resp != nil {
			result.resp.Body.Close()
		}
	}

	// Handle primary response
	if primaryErr != nil {
		debugLog.Printf("Primary backend error: %v", primaryErr)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if primaryResp == nil {
		http.Error(w, "No response from primary backend", http.StatusInternalServerError)
		return
	}

	defer primaryResp.Body.Close()

	// Copy headers from primary response
	for key, values := range primaryResp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Copy status code
	w.WriteHeader(primaryResp.StatusCode)

	// Copy body
	if _, err := io.Copy(w, primaryResp.Body); err != nil {
		debugLog.Printf("Error copying response body: %v", err)
	}
}
