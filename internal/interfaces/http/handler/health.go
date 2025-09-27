package handler

import (
	"encoding/json"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/jorge-j1m/hackspark_server/internal/infrastructure/config"
)

// HealthResponse is the response structure for the health check endpoint
type HealthResponse struct {
	Status      string            `json:"status"`
	Version     string            `json:"version"`
	Environment string            `json:"environment"`
	Timestamp   time.Time         `json:"timestamp"`
	Uptime      string            `json:"uptime"`
	SystemInfo  map[string]string `json:"system_info"`
}

// StartTime is the time when the server started
var StartTime = time.Now()

// HealthHandler handles health check requests
type HealthHandler struct {
	cfg *config.Config
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(cfg *config.Config) *HealthHandler {
	return &HealthHandler{
		cfg: cfg,
	}
}

// HealthCheck returns a handler for the health check endpoint
func (h *HealthHandler) Handle(w http.ResponseWriter, r *http.Request) {
	// Prepare system info
	systemInfo := map[string]string{
		"go_version":   runtime.Version(),
		"os":           runtime.GOOS,
		"architecture": runtime.GOARCH,
		"max_procs":    strconv.Itoa(runtime.GOMAXPROCS(0)),
		"goroutines":   strconv.Itoa(runtime.NumGoroutine()),
	}

	// Return response as JSON
	w.Header().Set("Content-Type", "application/json")

	status := "ok"
	statusCode := http.StatusOK

	w.WriteHeader(statusCode)

	// Create health response
	response := HealthResponse{
		Status:      status,
		Version:     h.cfg.Version,
		Environment: h.cfg.Environment,
		Timestamp:   time.Now(),
		Uptime:      time.Since(StartTime).String(),
		SystemInfo:  systemInfo,
	}
	result, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = w.Write(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
