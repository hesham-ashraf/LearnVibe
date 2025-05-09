package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/sony/gobreaker"
)

// LogLevel represents the severity level of a log message
type LogLevel string

const (
	// LogLevelDebug represents debug log level
	LogLevelDebug LogLevel = "debug"
	// LogLevelInfo represents info log level
	LogLevelInfo LogLevel = "info"
	// LogLevelWarning represents warning log level
	LogLevelWarning LogLevel = "warning"
	// LogLevelError represents error log level
	LogLevelError LogLevel = "error"
	// LogLevelFatal represents fatal log level
	LogLevelFatal LogLevel = "fatal"
)

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp   time.Time              `json:"@timestamp"`
	Level       LogLevel               `json:"level"`
	Message     string                 `json:"message"`
	ServiceName string                 `json:"service_name"`
	Hostname    string                 `json:"hostname"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// LoggingService provides centralized logging functionality
type LoggingService struct {
	serviceName   string
	opensearchURL string
	hostname      string
	client        *http.Client
	cb            *gobreaker.CircuitBreaker
	asyncLogs     chan LogEntry
}

// NewLoggingService creates a new instance of the logging service
func NewLoggingService(serviceName, opensearchURL string) (*LoggingService, error) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	// Set up circuit breaker for logging
	settings := gobreaker.Settings{
		Name:    "LoggingService",
		Timeout: 30 * time.Second,
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Create logging service
	ls := &LoggingService{
		serviceName:   serviceName,
		opensearchURL: opensearchURL,
		hostname:      hostname,
		client:        client,
		cb:            gobreaker.NewCircuitBreaker(settings),
		asyncLogs:     make(chan LogEntry, 1000), // Buffer for 1000 log entries
	}

	// Start the background worker to process logs asynchronously
	go ls.processLogs()

	// Verify connection to OpenSearch
	if err := ls.verifyConnection(); err != nil {
		log.Printf("Warning: Could not connect to OpenSearch: %v. Will retry in background.", err)
	}

	return ls, nil
}

// verifyConnection checks the connection to OpenSearch
func (ls *LoggingService) verifyConnection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", ls.opensearchURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := ls.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to OpenSearch: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("OpenSearch returned status code %d", resp.StatusCode)
	}

	return nil
}

// processLogs processes logs asynchronously
func (ls *LoggingService) processLogs() {
	for entry := range ls.asyncLogs {
		// Try to send the log entry, with backoff
		operation := func() error {
			_, err := ls.cb.Execute(func() (interface{}, error) {
				return ls.sendLogToOpenSearch(entry)
			})
			return err
		}

		// Use exponential backoff for retries
		backoffConfig := backoff.NewExponentialBackOff()
		backoffConfig.MaxElapsedTime = 1 * time.Minute

		if err := backoff.Retry(operation, backoffConfig); err != nil {
			// If all retries failed, log locally as fallback
			log.Printf("Failed to send log to OpenSearch after retries: %v", err)
			log.Printf("[%s] %s: %s", entry.Level, entry.Timestamp.Format(time.RFC3339), entry.Message)
		}
	}
}

// sendLogToOpenSearch sends a log entry to OpenSearch
func (ls *LoggingService) sendLogToOpenSearch(entry LogEntry) (interface{}, error) {
	// Format the log entry as JSON
	data, err := json.Marshal(entry)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal log entry: %v", err)
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Prepare the request
	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		fmt.Sprintf("%s/logs-%s/_doc", ls.opensearchURL, time.Now().Format("2006.01.02")),
		bytes.NewBuffer(data),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := ls.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send log to OpenSearch: %v", err)
	}
	defer resp.Body.Close()

	// Check if the request was successful
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("OpenSearch returned status code %d", resp.StatusCode)
	}

	return nil, nil
}

// createLogEntry creates a log entry with the given level and message
func (ls *LoggingService) createLogEntry(level LogLevel, message string, metadata map[string]interface{}) LogEntry {
	return LogEntry{
		Timestamp:   time.Now(),
		Level:       level,
		Message:     message,
		ServiceName: ls.serviceName,
		Hostname:    ls.hostname,
		Metadata:    metadata,
	}
}

// Log logs a message with the specified level
func (ls *LoggingService) Log(level LogLevel, message string, metadata map[string]interface{}) {
	entry := ls.createLogEntry(level, message, metadata)

	// Always log to console as fallback
	log.Printf("[%s] %s", level, message)

	// Queue for async processing
	select {
	case ls.asyncLogs <- entry:
		// Log entry queued
	default:
		// Channel is full, log locally as fallback
		log.Printf("Warning: Log buffer full. Logging locally: [%s] %s", level, message)
	}
}

// Debug logs a debug message
func (ls *LoggingService) Debug(message string, metadata map[string]interface{}) {
	ls.Log(LogLevelDebug, message, metadata)
}

// Info logs an info message
func (ls *LoggingService) Info(message string, metadata map[string]interface{}) {
	ls.Log(LogLevelInfo, message, metadata)
}

// Warning logs a warning message
func (ls *LoggingService) Warning(message string, metadata map[string]interface{}) {
	ls.Log(LogLevelWarning, message, metadata)
}

// Error logs an error message
func (ls *LoggingService) Error(message string, err error, metadata map[string]interface{}) {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	if err != nil {
		metadata["error"] = err.Error()
	}
	ls.Log(LogLevelError, message, metadata)
}

// Fatal logs a fatal message and exits the application
func (ls *LoggingService) Fatal(message string, err error, metadata map[string]interface{}) {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	if err != nil {
		metadata["error"] = err.Error()
	}
	ls.Log(LogLevelFatal, message, metadata)
	log.Fatalf("FATAL: %s: %v", message, err)
}

// Close closes the logging service and flushes any pending logs
func (ls *LoggingService) Close() {
	close(ls.asyncLogs)
}
