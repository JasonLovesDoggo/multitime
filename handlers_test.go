package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func setupTestConfig() {
	config = &Config{
		Port:  3000,
		Debug: false,
		Backends: []Backend{
			{
				Name:      "Primary Backend",
				URL:       "http://primary.example.com",
				APIKey:    "primary-key",
				IsPrimary: true,
			},
			{
				Name:      "Secondary Backend",
				URL:       "http://secondary.example.com",
				APIKey:    "secondary-key",
				IsPrimary: false,
			},
		},
	}

	debugLog = log.New(io.Discard, "", 0)
}

func TestHandleStatusBar(t *testing.T) {
	setupTestConfig()

	primaryServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/users/current/statusbar/today" {
			t.Errorf("Expected path /v1/users/current/statusbar/today, got %s", r.URL.Path)
		}

		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Basic ") {
			t.Errorf("Expected Basic auth header, got %s", authHeader)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":{"grand_total":{"decimal":"5.5","digital":"5:30","hours":5,"minutes":30,"text":"5 hrs 30 mins","total_seconds":19800},"range":{"text":"Today","timezone":"UTC"}}}`))
	}))
	defer primaryServer.Close()

	// Update primary backend URL to point to our test server
	for i := range config.Backends {
		if config.Backends[i].IsPrimary {
			config.Backends[i].URL = primaryServer.URL
		}
	}

	req, err := http.NewRequest("GET", "/users/current/statusbar/today", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("User-Agent", "TestUserAgent")

	rr := httptest.NewRecorder()

	handleStatusBar(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := `{"data":{"grand_total":{"decimal":"5.5","digital":"5:30","hours":5,"minutes":30,"text":"5 hrs 30 mins","total_seconds":19800},"range":{"text":"Today","timezone":"UTC"}}}`
	if rr.Body.String() != expected {
		t.Errorf("Handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}

	req, _ = http.NewRequest("POST", "/users/current/statusbar/today", nil)
	rr = httptest.NewRecorder()
	handleStatusBar(rr, req)
	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("Handler should return 405 for non-GET requests, got %v", status)
	}
}

func TestHandleHeartbeat(t *testing.T) {
	setupTestConfig()

	primaryServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/users/current/heartbeats" {
			t.Errorf("Expected path /v1/users/current/heartbeats, got %s", r.URL.Path)
		}

		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}
		defer r.Body.Close()

		if string(body) != `{"test":"heartbeat"}` {
			t.Errorf("Expected request body {\"test\":\"heartbeat\"}, got %s", string(body))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte(`{"data":"primary success"}`))
	}))
	defer primaryServer.Close()

	secondaryServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.URL.Path != "/v1/users/current/heartbeats" {
			t.Errorf("Expected path /v1/users/current/heartbeats, got %s", r.URL.Path)
		}

		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}
		defer r.Body.Close()

		if string(body) != `{"test":"heartbeat"}` {
			t.Errorf("Expected request body {\"test\":\"heartbeat\"}, got %s", string(body))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte(`{"data":"secondary success"}`))
	}))
	defer secondaryServer.Close()

	for i := range config.Backends {
		if config.Backends[i].IsPrimary {
			config.Backends[i].URL = primaryServer.URL
		} else {
			config.Backends[i].URL = secondaryServer.URL
		}
	}

	heartbeat := []byte(`{"test":"heartbeat"}`)
	req, err := http.NewRequest("POST", "/users/current/heartbeats", bytes.NewReader(heartbeat))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("User-Agent", "TestUserAgent")
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	handleHeartbeat(rr, req)

	if status := rr.Code; status != http.StatusAccepted {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusAccepted)
	}

	expected := `{"data":"primary success"}`
	if rr.Body.String() != expected {
		t.Errorf("Handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}

	req, _ = http.NewRequest("GET", "/users/current/heartbeats", nil)
	rr = httptest.NewRecorder()
	handleHeartbeat(rr, req)
	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("Handler should return 405 for non-POST requests, got %v", status)
	}

	// Test with invalid JSON
	req, _ = http.NewRequest("POST", "/users/current/heartbeats", bytes.NewReader([]byte("invalid json")))
	rr = httptest.NewRecorder()
	handleHeartbeat(rr, req)
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Handler should return 400 for invalid JSON, got %v", status)
	}
}

func TestHandleHeartbeatsBulk(t *testing.T) {

	setupTestConfig()

	primaryServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.URL.Path != "/v1/users/current/heartbeats.bulk" {
			t.Errorf("Expected path /v1/users/current/heartbeats.bulk, got %s", r.URL.Path)
		}

		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}
		defer r.Body.Close()

		if string(body) != `[{"test":"heartbeat1"},{"test":"heartbeat2"}]` {
			t.Errorf("Expected request body [{\"test\":\"heartbeat1\"},{\"test\":\"heartbeat2\"}], got %s", string(body))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte(`{"data":"primary bulk success"}`))
	}))
	defer primaryServer.Close()

	secondaryServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.URL.Path != "/v1/users/current/heartbeats.bulk" {
			t.Errorf("Expected path /v1/users/current/heartbeats.bulk, got %s", r.URL.Path)
		}

		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}
		defer r.Body.Close()

		if string(body) != `[{"test":"heartbeat1"},{"test":"heartbeat2"}]` {
			t.Errorf("Expected request body [{\"test\":\"heartbeat1\"},{\"test\":\"heartbeat2\"}], got %s", string(body))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte(`{"data":"secondary bulk success"}`))
	}))
	defer secondaryServer.Close()

	for i := range config.Backends {
		if config.Backends[i].IsPrimary {
			config.Backends[i].URL = primaryServer.URL
		} else {
			config.Backends[i].URL = secondaryServer.URL
		}
	}

	heartbeats := []byte(`[{"test":"heartbeat1"},{"test":"heartbeat2"}]`)
	req, err := http.NewRequest("POST", "/users/current/heartbeats.bulk", bytes.NewReader(heartbeats))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("User-Agent", "TestUserAgent")
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	handleHeartbeatsBulk(rr, req)

	if status := rr.Code; status != http.StatusAccepted {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusAccepted)
	}

	expected := `{"data":"primary bulk success"}`
	if rr.Body.String() != expected {
		t.Errorf("Handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}

	req, _ = http.NewRequest("GET", "/users/current/heartbeats.bulk", nil)
	rr = httptest.NewRecorder()
	handleHeartbeatsBulk(rr, req)
	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("Handler should return 405 for non-POST requests, got %v", status)
	}

	req, _ = http.NewRequest("POST", "/users/current/heartbeats.bulk", bytes.NewReader([]byte("invalid json")))
	rr = httptest.NewRecorder()
	handleHeartbeatsBulk(rr, req)
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Handler should return 400 for invalid JSON, got %v", status)
	}
}
