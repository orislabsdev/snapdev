package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/orislabsdev/snapdev/config"
	"github.com/orislabsdev/snapdev/logger"
)

func TestServerRouting(t *testing.T) {
	// 1. Setup mock environment
	tmpDir, err := os.MkdirTemp("", "snapdev-test-*")
	if err != nil {
		t.Fatalf("failed to create tmp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	distDir := filepath.Join(tmpDir, "dist")
	os.Mkdir(distDir, 0755)

	// Create some static files
	os.WriteFile(filepath.Join(distDir, "index.html"), []byte("ROOT_INDEX"), 0644)
	os.WriteFile(filepath.Join(distDir, "test.txt"), []byte("STATIC_FILE"), 0644)

	// Create a subdirectory with an index.html
	subDir := filepath.Join(distDir, "sub")
	os.Mkdir(subDir, 0755)
	os.WriteFile(filepath.Join(subDir, "index.html"), []byte("SUB_INDEX"), 0644)

	// 2. Setup mock backend server
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Proxied", "true")
		w.Header().Set("X-Host", r.Host)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("BACKEND_RESPONSE: " + r.URL.Path))
	}))
	defer backend.Close()

	// 3. Initialize snapdev server
	cfg := &config.Config{
		OutputDir:    distDir,
		Host:         "localhost",
		Port:         3000,
		ReverseProxy: backend.URL,
		LiveReload:   false,
	}
	log := logger.New().WithLevel(logger.LevelDebug)
	srv := New(cfg, log)

	// Use httptest to simulate requests to the server's handler
	handler := srv.httpSrv.Handler

	tests := []struct {
		name          string
		path          string
		headers       map[string]string
		expectedBody  string
		expectedProxy bool
		expectedHost  string // Only checked if proxied
	}{
		{
			name:         "Serve root index.html",
			path:         "/",
			expectedBody: "ROOT_INDEX",
		},
		{
			name:         "Serve static file",
			path:         "/test.txt",
			expectedBody: "STATIC_FILE",
		},
		{
			name:         "Serve subdirectory index.html",
			path:         "/sub/",
			expectedBody: "SUB_INDEX",
		},
		{
			name:         "Serve subdirectory index.html without trailing slash",
			path:         "/sub",
			expectedBody: "SUB_INDEX",
		},
		{
			name:          "Proxy unknown path",
			path:          "/api/data",
			expectedBody:  "BACKEND_RESPONSE: /api/data",
			expectedProxy: true,
			expectedHost:  strings.TrimPrefix(backend.URL, "http://"),
		},
		{
			name:          "SPA Fallback for browser navigation (Accept: text/html)",
			path:          "/some-route",
			headers:       map[string]string{"Accept": "text/html"},
			expectedBody:  "ROOT_INDEX",
			expectedProxy: false,
		},
		{
			name:          "Proxy unknown path (no HTML preference)",
			path:          "/some-route",
			expectedBody:  "BACKEND_RESPONSE: /some-route",
			expectedProxy: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			resp := w.Result()
			body, _ := io.ReadAll(resp.Body)

			if !strings.Contains(string(body), tt.expectedBody) {
				t.Errorf("Expected body containing %q, got %q", tt.expectedBody, string(body))
			}

			isProxied := resp.Header.Get("X-Proxied") == "true"
			if isProxied != tt.expectedProxy {
				t.Errorf("Expected proxied=%v, got %v", tt.expectedProxy, isProxied)
			}

			if tt.expectedProxy && tt.expectedHost != "" {
				host := resp.Header.Get("X-Host")
				if host != tt.expectedHost {
					t.Errorf("Expected Host header %q, got %q", tt.expectedHost, host)
				}
			}
		})
	}
}

func TestServerSPAWithNoProxy(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "snapdev-spa-test-*")
	defer os.RemoveAll(tmpDir)
	distDir := filepath.Join(tmpDir, "dist")
	os.Mkdir(distDir, 0755)
	os.WriteFile(filepath.Join(distDir, "index.html"), []byte("SPA_INDEX"), 0644)

	cfg := &config.Config{
		OutputDir:    distDir,
		ReverseProxy: "", // No proxy
		LiveReload:   false,
	}
	srv := New(cfg, logger.New())
	handler := srv.httpSrv.Handler

	req := httptest.NewRequest("GET", "/any-route", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "SPA_INDEX" {
		t.Errorf("Expected SPA_INDEX, got %q", string(body))
	}
}
