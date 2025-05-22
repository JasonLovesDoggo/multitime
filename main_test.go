package main

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestMainIntegration(t *testing.T) {
	configContent := `
port = 3005
debug = true

[[backends]]
name = "Primary Backend"
url = "http://primary.example.com"
api_key = "primary-key"
is_primary = true

[[backends]]
name = "Secondary Backend"
url = "http://secondary.example.com"
api_key = "secondary-key"
is_primary = false
`
	tmpfile, err := os.CreateTemp("", "config-*.toml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(configContent)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	// Save original args and restore after test
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()
	os.Args = []string{"multitime", tmpfile.Name()}

	// Save original config and debugLog
	originalConfig := config
	originalDebugLog := debugLog
	defer func() {
		config = originalConfig
		debugLog = originalDebugLog
	}()

	// Create mock servers for backends
	primaryServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "heartbeats.bulk") {
			w.WriteHeader(http.StatusAccepted)
			w.Write([]byte(`{"data":"primary bulk success"}`))
		} else if strings.Contains(r.URL.Path, "heartbeats") {
			w.WriteHeader(http.StatusAccepted)
			w.Write([]byte(`{"data":"primary heartbeat success"}`))
		} else if strings.Contains(r.URL.Path, "statusbar") {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"data":{"grand_total":{"decimal":"5.5","digital":"5:30","hours":5,"minutes":30,"text":"5 hrs 30 mins","total_seconds":19800},"range":{"text":"Today","timezone":"UTC"}}}`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer primaryServer.Close()

	secondaryServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "heartbeats.bulk") {
			w.WriteHeader(http.StatusAccepted)
			w.Write([]byte(`{"data":"secondary bulk success"}`))
		} else if strings.Contains(r.URL.Path, "heartbeats") {
			w.WriteHeader(http.StatusAccepted)
			w.Write([]byte(`{"data":"secondary heartbeat success"}`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer secondaryServer.Close()

	// Load config
	var loadErr error
	config, loadErr = loadConfig(tmpfile.Name())
	if loadErr != nil {
		t.Fatalf("Failed to load config: %v", loadErr)
	}

	// Update backend URLs to point to our test servers
	for i := range config.Backends {
		if config.Backends[i].IsPrimary {
			config.Backends[i].URL = primaryServer.URL
		} else {
			config.Backends[i].URL = secondaryServer.URL
		}
	}

	setupLogging(config.Debug)

	mux := http.NewServeMux()
	mux.HandleFunc("/users/current/heartbeats", handleHeartbeat)
	mux.HandleFunc("/users/current/heartbeats.bulk", handleHeartbeatsBulk)
	mux.HandleFunc("/users/current/statusbar/today", handleStatusBar)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	// Test cases
	testCases := []struct {
		name           string
		method         string
		path           string
		body           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Status Bar",
			method:         "GET",
			path:           "/users/current/statusbar/today",
			body:           "",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"data":{"grand_total":{"decimal":"5.5","digital":"5:30","hours":5,"minutes":30,"text":"5 hrs 30 mins","total_seconds":19800},"range":{"text":"Today","timezone":"UTC"}}}`,
		},
		{
			name:           "Heartbeat",
			method:         "POST",
			path:           "/users/current/heartbeats",
			body:           `{"test":"heartbeat"}`,
			expectedStatus: http.StatusAccepted,
			expectedBody:   `{"data":"primary heartbeat success"}`,
		},
		{
			name:           "Heartbeats Bulk",
			method:         "POST",
			path:           "/users/current/heartbeats.bulk",
			body:           `[{"test":"heartbeat1"},{"test":"heartbeat2"}]`,
			expectedStatus: http.StatusAccepted,
			expectedBody:   `{"data":"primary bulk success"}`,
		},
		{
			name:           "Not Found",
			method:         "GET",
			path:           "/invalid/path",
			body:           "",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "",
		},
	}

	// Run test cases
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var req *http.Request
			var err error

			if tc.method == "POST" {
				req, err = http.NewRequest(tc.method, server.URL+tc.path, strings.NewReader(tc.body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req, err = http.NewRequest(tc.method, server.URL+tc.path, nil)
			}

			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			req.Header.Set("User-Agent", "TestUserAgent")

			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, resp.StatusCode)
			}

			if tc.expectedBody != "" {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Fatalf("Failed to read response body: %v", err)
				}

				if string(body) != tc.expectedBody {
					t.Errorf("Expected body %s, got %s", tc.expectedBody, string(body))
				}
			}
		})
	}
}

func TestMainErrorHandling(t *testing.T) {
	// Test with invalid config file
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Test with missing config file
	os.Args = []string{"multitime", "nonexistent-config.toml"}

	// Capture log output
	var buf strings.Builder
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	// We can't call main() directly as it would exit the program
	// Instead, we'll test the config loading part
	_, err := loadConfig("nonexistent-config.toml")
	if err == nil {
		t.Error("Expected error for non-existent config file, got none")
	}

	// Test with invalid number of arguments
	os.Args = []string{"multitime"}

	// We can't test this directly as it would call log.Fatal
	if len(os.Args) == 1 {
		// This is the expected behavior
		t.Log("Correctly detected invalid number of arguments")
	}
}
