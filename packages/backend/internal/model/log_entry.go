package model

import "time"

// LogEntry represents a single structured log record for API responses.
type LogEntry struct {
	ID         string                 `json:"id"`
	Channel    string                 `json:"channel"`
	Timestamp  time.Time              `json:"timestamp"`
	Level      string                 `json:"level"`
	Module     string                 `json:"module"`
	Action     string                 `json:"action"`
	Message    string                 `json:"message"`
	RequestID  string                 `json:"request_id"`
	UserID     string                 `json:"user_id,omitempty"`
	Method     string                 `json:"method,omitempty"`
	Path       string                 `json:"path,omitempty"`
	Route      string                 `json:"route,omitempty"`
	ClientIP   string                 `json:"client_ip,omitempty"`
	StatusCode int                    `json:"status_code,omitempty"`
	LatencyMS  int64                  `json:"latency_ms,omitempty"`
	Error      string                 `json:"error,omitempty"`
	Raw        map[string]interface{} `json:"raw,omitempty"`
}
