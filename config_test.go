package main

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Test cases
	tests := []struct {
		name          string
		configContent string
		expectError   bool
		expectedPort  int
		expectedDebug bool
		backendCount  int
	}{
		{
			name: "Valid config with one primary backend",
			configContent: `
port = 3005
debug = true

[[backends]]
name = "Backend 1"
url = "https://example.com/api"
api_key = "key1"
is_primary = true

[[backends]]
name = "Backend 2"
url = "https://example2.com/api"
api_key = "key2"
`,
			expectError:   false,
			expectedPort:  3005,
			expectedDebug: true,
			backendCount:  2,
		},
		{
			name: "Default port when not specified",
			configContent: `
debug = false

[[backends]]
name = "Backend 1"
url = "https://example.com/api"
api_key = "key1"
is_primary = true
`,
			expectError:   false,
			expectedPort:  3000, // Default port
			expectedDebug: false,
			backendCount:  1,
		},
		{
			name: "Error when no primary backend",
			configContent: `
port = 3005
debug = true

[[backends]]
name = "Backend 1"
url = "https://example.com/api"
api_key = "key1"
is_primary = false

[[backends]]
name = "Backend 2"
url = "https://example2.com/api"
api_key = "key2"
is_primary = false
`,
			expectError: true,
		},
		{
			name: "Error when multiple primary backends",
			configContent: `
port = 3005
debug = true

[[backends]]
name = "Backend 1"
url = "https://example.com/api"
api_key = "key1"
is_primary = true

[[backends]]
name = "Backend 2"
url = "https://example2.com/api"
api_key = "key2"
is_primary = true
`,
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create temporary config file
			tmpfile, err := os.CreateTemp("", "config-*.toml")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpfile.Name())

			// Write config content
			if _, err := tmpfile.Write([]byte(tc.configContent)); err != nil {
				t.Fatalf("Failed to write to temp file: %v", err)
			}
			if err := tmpfile.Close(); err != nil {
				t.Fatalf("Failed to close temp file: %v", err)
			}

			// Load config
			cfg, err := loadConfig(tmpfile.Name())

			// Check results
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if cfg.Port != tc.expectedPort {
				t.Errorf("Expected port %d, got %d", tc.expectedPort, cfg.Port)
			}

			if cfg.Debug != tc.expectedDebug {
				t.Errorf("Expected debug %v, got %v", tc.expectedDebug, cfg.Debug)
			}

			if len(cfg.Backends) != tc.backendCount {
				t.Errorf("Expected %d backends, got %d", tc.backendCount, len(cfg.Backends))
			}

			// Verify one and only one primary backend
			primaryCount := 0
			for _, b := range cfg.Backends {
				if b.IsPrimary {
					primaryCount++
				}
			}
			if primaryCount != 1 {
				t.Errorf("Expected exactly 1 primary backend, got %d", primaryCount)
			}
		})
	}
}

func TestInvalidConfigFile(t *testing.T) {
	// Test with non-existent file
	_, err := loadConfig("nonexistent-file.toml")
	if err == nil {
		t.Error("Expected error for non-existent file, got none")
	}

	// Test with invalid TOML
	tmpfile, err := os.CreateTemp("", "invalid-*.toml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	// Write invalid TOML
	if _, err := tmpfile.Write([]byte("this is not valid TOML")); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	_, err = loadConfig(tmpfile.Name())
	if err == nil {
		t.Error("Expected error for invalid TOML, got none")
	}
}
