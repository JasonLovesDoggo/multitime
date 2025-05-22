package main

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestForwardHeartbeat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.URL.Path != "/v1/users/current/heartbeats" {
			t.Errorf("Expected path /v1/users/current/heartbeats, got %s", r.URL.Path)
		}

		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type: application/json, got %s", r.Header.Get("Content-Type"))
		}

		// Check authorization header
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Basic ") {
			t.Errorf("Expected Basic auth header, got %s", authHeader)
		}

		// Check user agent
		userAgent := r.Header.Get("User-Agent")
		if !strings.Contains(userAgent, "(JasonLovesDoggo/multitime)") {
			t.Errorf("Expected User-Agent to contain (JasonLovesDoggo/multitime), got %s", userAgent)
		}

		// Read request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}
		defer r.Body.Close()

		if string(body) != `{"test":"heartbeat"}` {
			t.Errorf("Expected request body {\"test\":\"heartbeat\"}, got %s", string(body))
		}

		// Return success response
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte(`{"data":"success"}`))
	}))
	defer server.Close()

	// Setup test backend
	backend := Backend{
		Name:      "Test Backend",
		URL:       server.URL,
		APIKey:    "test-api-key",
		IsPrimary: true,
	}

	// Setup debug logging
	originalDebugLog := debugLog
	debugLog = log.New(io.Discard, "", 0) // Silence debug logs during test
	defer func() { debugLog = originalDebugLog }()

	// Test forwarding heartbeat
	heartbeat := []byte(`{"test":"heartbeat"}`)
	userAgent := "TestUserAgent"

	resp, err := forwardHeartbeat(heartbeat, userAgent, backend)
	if err != nil {
		t.Fatalf("forwardHeartbeat returned error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		t.Errorf("Expected status code %d, got %d", http.StatusAccepted, resp.StatusCode)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	if string(body) != `{"data":"success"}` {
		t.Errorf("Expected response body {\"data\":\"success\"}, got %s", string(body))
	}
}

func TestForwardHeartbeats(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.URL.Path != "/v1/users/current/heartbeats.bulk" {
			t.Errorf("Expected path /v1/users/current/heartbeats.bulk, got %s", r.URL.Path)
		}

		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type: application/json, got %s", r.Header.Get("Content-Type"))
		}

		// Check authorization header
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Basic ") {
			t.Errorf("Expected Basic auth header, got %s", authHeader)
		}

		// Check user agent
		userAgent := r.Header.Get("User-Agent")
		if !strings.Contains(userAgent, "(JasonLovesDoggo/multitime)") {
			t.Errorf("Expected User-Agent to contain (JasonLovesDoggo/multitime), got %s", userAgent)
		}

		// Read request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}
		defer r.Body.Close()

		if string(body) != `[{"test":"heartbeat1"},{"test":"heartbeat2"}]` {
			t.Errorf("Expected request body [{\"test\":\"heartbeat1\"},{\"test\":\"heartbeat2\"}], got %s", string(body))
		}

		// Return success response
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte(`{"data":"success"}`))
	}))
	defer server.Close()

	// Setup test backend
	backend := Backend{
		Name:      "Test Backend",
		URL:       server.URL,
		APIKey:    "test-api-key",
		IsPrimary: true,
	}

	// Setup debug logging
	originalDebugLog := debugLog
	debugLog = log.New(io.Discard, "", 0) // Silence debug logs during test
	defer func() { debugLog = originalDebugLog }()

	// Test forwarding heartbeats
	heartbeats := []byte(`[{"test":"heartbeat1"},{"test":"heartbeat2"}]`)
	userAgent := "TestUserAgent"

	resp, err := forwardHeartbeats(heartbeats, userAgent, backend)
	if err != nil {
		t.Fatalf("forwardHeartbeats returned error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		t.Errorf("Expected status code %d, got %d", http.StatusAccepted, resp.StatusCode)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	if string(body) != `{"data":"success"}` {
		t.Errorf("Expected response body {\"data\":\"success\"}, got %s", string(body))
	}
}

func TestForwardHeartbeatError(t *testing.T) {
	// Setup test backend with invalid URL
	backend := Backend{
		Name:      "Test Backend",
		URL:       "http://invalid-url-that-does-not-exist.example",
		APIKey:    "test-api-key",
		IsPrimary: true,
	}

	// Setup debug logging
	originalDebugLog := debugLog
	debugLog = log.New(io.Discard, "", 0) // Silence debug logs during test
	defer func() { debugLog = originalDebugLog }()

	// Test forwarding heartbeat
	heartbeat := []byte(`{"test":"heartbeat"}`)
	userAgent := "TestUserAgent"

	_, err := forwardHeartbeat(heartbeat, userAgent, backend)
	if err == nil {
		t.Error("Expected error for invalid URL, got none")
	}
}

func TestForwardHeartbeatsError(t *testing.T) {
	// Setup test backend with invalid URL
	backend := Backend{
		Name:      "Test Backend",
		URL:       "http://invalid-url-that-does-not-exist.example",
		APIKey:    "test-api-key",
		IsPrimary: true,
	}

	// Setup debug logging
	originalDebugLog := debugLog
	debugLog = log.New(io.Discard, "", 0) // Silence debug logs during test
	defer func() { debugLog = originalDebugLog }()

	// Test forwarding heartbeats
	heartbeats := []byte(`[{"test":"heartbeat1"},{"test":"heartbeat2"}]`)
	userAgent := "TestUserAgent"

	_, err := forwardHeartbeats(heartbeats, userAgent, backend)
	if err == nil {
		t.Error("Expected error for invalid URL, got none")
	}
}
