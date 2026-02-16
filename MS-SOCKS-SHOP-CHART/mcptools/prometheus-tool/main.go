package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Build-time variables (set via -ldflags)
var (
	version   = "1.0.0"
	buildTime = "unknown"
)

// maxResponseSize limits the maximum response body size from Prometheus (10MB)
const maxResponseSize = 10 * 1024 * 1024

// App encapsulates all application dependencies — no globals
type App struct {
	prometheusURL string
	httpClient    *http.Client
	allowedOrigin string
}

// NewApp creates a new App with configuration from environment variables
func NewApp() *App {
	promURL := os.Getenv("PROMETHEUS_URL")
	if promURL == "" {
		log.Fatal("PROMETHEUS_URL environment variable is required (e.g. http://prometheus.monitoring.svc.cluster.local:9090)")
	}

	corsOrigin := os.Getenv("CORS_ALLOWED_ORIGIN")
	if corsOrigin == "" {
		corsOrigin = "*" // Default; restrict in production via env var
	}

	return &App{
		prometheusURL: promURL,
		allowedOrigin: corsOrigin,
		// Reusable HTTP client with timeouts
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				IdleConnTimeout:     60 * time.Second,
				DisableCompression:  false,
				MaxConnsPerHost:     10,
				MaxIdleConnsPerHost: 5,
			},
		},
	}
}

// jsonError writes a consistent JSON error response
func jsonError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "error",
		"error":  message,
		"code":   statusCode,
	})
}

// requestLogger is a middleware that logs method, path, status, and duration
func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		// Wrap ResponseWriter to capture status code
		sw := &statusWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(sw, r)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, sw.statusCode, time.Since(start).Round(time.Millisecond))
	})
}

// statusWriter wraps http.ResponseWriter to capture the status code
type statusWriter struct {
	http.ResponseWriter
	statusCode int
}

func (sw *statusWriter) WriteHeader(code int) {
	sw.statusCode = code
	sw.ResponseWriter.WriteHeader(code)
}

// corsMiddleware adds CORS headers with configurable origin
func (app *App) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", app.allowedOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("X-Content-Type-Options", "nosniff")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// methodGuard rejects non-GET requests with a proper JSON error
func methodGuard(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			jsonError(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		next(w, r)
	}
}

func main() {
	app := NewApp()

	// Set up routes with method guards
	mux := http.NewServeMux()
	mux.HandleFunc("/", methodGuard(app.handleRoot))
	mux.HandleFunc("/health", methodGuard(app.handleHealth))
	mux.HandleFunc("/query", methodGuard(app.handleQuery))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}

	// Chain middleware: CORS → Logging → Routes
	handler := app.corsMiddleware(requestLogger(mux))

	// Configure server with proper timeouts
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown on SIGTERM/SIGINT
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("MCP Prometheus Tool v%s (built: %s)", version, buildTime)
		log.Printf("Prometheus URL: %s", app.prometheusURL)
		log.Printf("CORS Origin: %s", app.allowedOrigin)
		log.Printf("Server starting on :%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Block until shutdown signal
	<-stop
	log.Println("Shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Forced shutdown: %v", err)
	}
	log.Println("Server stopped")
}

// handleRoot provides API documentation
func (app *App) handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>MCP Prometheus Tool - SockShop</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; background: #f9f9f9; }
        h1 { color: #e6522c; }
        code { background: #f4f4f4; padding: 2px 6px; border-radius: 3px; }
        .endpoint { background: #fff; padding: 15px; margin: 10px 0; border-radius: 5px; border-left: 4px solid #e6522c; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
        .tag { display: inline-block; background: #e6522c; color: white; padding: 2px 8px; border-radius: 3px; font-size: 12px; }
        .version { color: #888; font-size: 12px; }
    </style>
</head>
<body>
    <h1>🔥 MCP Prometheus Tool - SockShop</h1>
    <p>Query Prometheus metrics for SockShop microservices via HTTP API</p>
    <p class="version">Version: %s | Built: %s</p>
    
    <div class="endpoint">
        <h3><span class="tag">GET</span> /query</h3>
        <p>Execute an instant PromQL query (query_prometheus)</p>
        <p><strong>Parameters:</strong></p>
        <ul>
            <li><code>query</code> - PromQL expression (required)</li>
            <li><code>time</code> - Evaluation timestamp (optional, RFC3339 or Unix)</li>
        </ul>
        <p><strong>Examples:</strong></p>
        <code>/query?query=up{job=~"sock-shop/.*"}</code><br><br>
        <code>/query?query=go_goroutines{job=~"sock-shop/.*"}</code><br><br>
        <code>/query?query=go_memstats_alloc_bytes{job=~"sock-shop/.*"}</code>
    </div>

    <div class="endpoint">
        <h3><span class="tag">GET</span> /health</h3>
        <p>Health check - also verifies Prometheus connectivity</p>
    </div>

    <h3>Useful PromQL Queries for SockShop</h3>
    <table border="1" cellpadding="8" cellspacing="0" style="border-collapse: collapse;">
        <tr><th>Query</th><th>Description</th></tr>
        <tr><td><code>up{job=~"sock-shop/.*"}</code></td><td>Service availability</td></tr>
        <tr><td><code>go_goroutines{job=~"sock-shop/.*"}</code></td><td>Goroutines per service</td></tr>
        <tr><td><code>go_memstats_alloc_bytes{job=~"sock-shop/.*"}</code></td><td>Memory usage</td></tr>
        <tr><td><code>rate(process_cpu_seconds_total{job=~"sock-shop/.*"}[5m])</code></td><td>CPU usage rate</td></tr>
        <tr><td><code>process_open_fds{job=~"sock-shop/.*"}</code></td><td>Open file descriptors</td></tr>
    </table>
</body>
</html>
`, version, buildTime)
}

// handleHealth provides health check with Prometheus connectivity verification
func (app *App) handleHealth(w http.ResponseWriter, r *http.Request) {
	promHealth := "unknown"

	// Use a short timeout for health check — don't block liveness probes
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, app.prometheusURL+"/-/healthy", nil)
	if err == nil {
		resp, err := app.httpClient.Do(req)
		if err == nil {
			defer resp.Body.Close()
			// Discard body to reuse connection
			_, _ = io.Copy(io.Discard, resp.Body)
			if resp.StatusCode == http.StatusOK {
				promHealth = "healthy"
			} else {
				promHealth = "unhealthy"
			}
		} else {
			promHealth = "unreachable"
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":            "healthy",
		"service":           "sockshop-prometheus-tool",
		"version":           version,
		"prometheus_status": promHealth,
		"timestamp":         time.Now().Format(time.RFC3339),
	})
}

// handleQuery executes an instant PromQL query
func (app *App) handleQuery(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	if query == "" {
		jsonError(w, "query parameter is required. Example: /query?query=up", http.StatusBadRequest)
		return
	}

	evalTime := r.URL.Query().Get("time")

	// Build Prometheus API URL
	params := url.Values{}
	params.Set("query", query)
	if evalTime != "" {
		params.Set("time", evalTime)
	}

	apiURL := fmt.Sprintf("%s/api/v1/query?%s", app.prometheusURL, params.Encode())
	log.Printf("Querying Prometheus: query=%s", query)

	result, err := app.queryPrometheus(r.Context(), apiURL)
	if err != nil {
		log.Printf("Prometheus query failed: %v", err)
		jsonError(w, "Prometheus query failed", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(result)
}

// queryPrometheus makes an HTTP GET request to the Prometheus API with context support
func (app *App) queryPrometheus(ctx context.Context, apiURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := app.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Prometheus: %w", err)
	}
	defer resp.Body.Close()

	// Limit response body size to prevent memory exhaustion
	limitedReader := io.LimitReader(resp.Body, maxResponseSize)
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Prometheus returned status %d", resp.StatusCode)
	}

	return body, nil
}
