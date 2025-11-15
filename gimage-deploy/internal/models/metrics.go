package models

import "time"

// DeploymentMetrics represents CloudWatch metrics for a deployment
type DeploymentMetrics struct {
	DeploymentID    string      `json:"deployment_id"`
	Period          string      `json:"period"` // 1h, 24h, 7d
	Invocations     int64       `json:"invocations"`
	Errors          int64       `json:"errors"`
	Throttles       int64       `json:"throttles"`
	ConcurrentExec  int         `json:"concurrent_executions"`
	AvgDuration     float64     `json:"avg_duration_ms"`
	P50Latency      float64     `json:"p50_latency_ms"`
	P95Latency      float64     `json:"p95_latency_ms"`
	P99Latency      float64     `json:"p99_latency_ms"`
	ErrorRate       float64     `json:"error_rate"`
	DataPointsTime  []time.Time `json:"data_points_time"`
	DataPointsCount []int64     `json:"data_points_count"`
}

// LogEntry represents a CloudWatch log entry
type LogEntry struct {
	Timestamp   time.Time `json:"timestamp"`
	Message     string    `json:"message"`
	Level       string    `json:"level"` // INFO, WARN, ERROR, DEBUG
	StreamName  string    `json:"stream_name"`
	RequestID   string    `json:"request_id,omitempty"`
}
