// Package batch provides batch operation tracking and history functionality.
package batch

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// OperationStatus represents the status of a batch operation.
type OperationStatus string

const (
	StatusPending   OperationStatus = "pending"
	StatusRunning   OperationStatus = "running"
	StatusCompleted OperationStatus = "completed"
	StatusFailed    OperationStatus = "failed"
	StatusCancelled OperationStatus = "cancelled"
)

// OperationType represents the type of batch operation.
type OperationType string

const (
	OpTypeResize   OperationType = "resize"
	OpTypeScale    OperationType = "scale"
	OpTypeCrop     OperationType = "crop"
	OpTypeCompress OperationType = "compress"
	OpTypeConvert  OperationType = "convert"
	OpTypeGenerate OperationType = "generate"
)

// OperationResult represents the result of a single batch operation.
type OperationResult struct {
	ID        string            `json:"id"`
	Operation OperationType     `json:"operation"`
	Status    OperationStatus   `json:"status"`
	Input     string            `json:"input"`
	Output    string            `json:"output"`
	StartTime time.Time         `json:"start_time"`
	EndTime   time.Time         `json:"end_time"`
	Duration  time.Duration     `json:"duration"`
	Error     string            `json:"error,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"` // Operation-specific metadata
}

// BatchHistory tracks the history of batch operations.
type BatchHistory struct {
	mu      sync.RWMutex
	results []OperationResult
	maxSize int // Maximum number of results to keep in memory
}

// NewBatchHistory creates a new batch history tracker.
// maxSize limits the number of results kept in memory (0 = unlimited).
func NewBatchHistory(maxSize int) *BatchHistory {
	if maxSize < 0 {
		maxSize = 0
	}
	return &BatchHistory{
		results: []OperationResult{},
		maxSize: maxSize,
	}
}

// Add adds a new operation result to the history.
func (bh *BatchHistory) Add(result OperationResult) {
	bh.mu.Lock()
	defer bh.mu.Unlock()

	// Set ID if not provided
	if result.ID == "" {
		result.ID = fmt.Sprintf("%s-%d", result.Operation, time.Now().UnixNano())
	}

	// Calculate duration if not set
	if result.Duration == 0 && !result.EndTime.IsZero() {
		result.Duration = result.EndTime.Sub(result.StartTime)
	}

	bh.results = append(bh.results, result)

	// Trim to max size if set
	if bh.maxSize > 0 && len(bh.results) > bh.maxSize {
		bh.results = bh.results[len(bh.results)-bh.maxSize:]
	}
}

// GetAll returns all operation results.
func (bh *BatchHistory) GetAll() []OperationResult {
	bh.mu.RLock()
	defer bh.mu.RUnlock()

	// Return a copy to prevent external modification
	results := make([]OperationResult, len(bh.results))
	copy(results, bh.results)
	return results
}

// GetByStatus returns operations with a specific status.
func (bh *BatchHistory) GetByStatus(status OperationStatus) []OperationResult {
	bh.mu.RLock()
	defer bh.mu.RUnlock()

	filtered := []OperationResult{}
	for _, result := range bh.results {
		if result.Status == status {
			filtered = append(filtered, result)
		}
	}
	return filtered
}

// GetByOperation returns operations of a specific type.
func (bh *BatchHistory) GetByOperation(opType OperationType) []OperationResult {
	bh.mu.RLock()
	defer bh.mu.RUnlock()

	filtered := []OperationResult{}
	for _, result := range bh.results {
		if result.Operation == opType {
			filtered = append(filtered, result)
		}
	}
	return filtered
}

// GetByID returns a specific operation by ID.
func (bh *BatchHistory) GetByID(id string) (OperationResult, bool) {
	bh.mu.RLock()
	defer bh.mu.RUnlock()

	for _, result := range bh.results {
		if result.ID == id {
			return result, true
		}
	}
	return OperationResult{}, false
}

// Count returns the total number of operations in history.
func (bh *BatchHistory) Count() int {
	bh.mu.RLock()
	defer bh.mu.RUnlock()
	return len(bh.results)
}

// Clear removes all results from history.
func (bh *BatchHistory) Clear() {
	bh.mu.Lock()
	defer bh.mu.Unlock()
	bh.results = []OperationResult{}
}

// GetStats returns statistics about the batch operations.
func (bh *BatchHistory) GetStats() BatchStats {
	bh.mu.RLock()
	defer bh.mu.RUnlock()

	stats := BatchStats{
		Total: len(bh.results),
	}

	var totalDuration time.Duration
	for _, result := range bh.results {
		switch result.Status {
		case StatusCompleted:
			stats.Completed++
		case StatusFailed:
			stats.Failed++
		case StatusCancelled:
			stats.Cancelled++
		case StatusRunning:
			stats.Running++
		case StatusPending:
			stats.Pending++
		}
		totalDuration += result.Duration
	}

	if stats.Total > 0 {
		stats.AverageDuration = totalDuration / time.Duration(stats.Total)
	}

	return stats
}

// BatchStats contains statistics about batch operations.
type BatchStats struct {
	Total           int
	Completed       int
	Failed          int
	Cancelled       int
	Running         int
	Pending         int
	AverageDuration time.Duration
}

// SaveToFile persists the batch history to a JSON file.
func (bh *BatchHistory) SaveToFile(path string) error {
	bh.mu.RLock()
	defer bh.mu.RUnlock()

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(bh.results, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}

	// Write to file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write history file: %w", err)
	}

	return nil
}

// LoadFromFile loads batch history from a JSON file.
func (bh *BatchHistory) LoadFromFile(path string) error {
	bh.mu.Lock()
	defer bh.mu.Unlock()

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read history file: %w", err)
	}

	// Unmarshal JSON
	var results []OperationResult
	if err := json.Unmarshal(data, &results); err != nil {
		return fmt.Errorf("failed to unmarshal history: %w", err)
	}

	bh.results = results

	// Trim to max size if set
	if bh.maxSize > 0 && len(bh.results) > bh.maxSize {
		bh.results = bh.results[len(bh.results)-bh.maxSize:]
	}

	return nil
}

// ExportSummary creates a human-readable summary of the batch history.
func (bh *BatchHistory) ExportSummary() string {
	bh.mu.RLock()
	defer bh.mu.RUnlock()

	stats := bh.GetStats()

	summary := fmt.Sprintf("Batch Operation Summary\n")
	summary += fmt.Sprintf("=======================\n\n")
	summary += fmt.Sprintf("Total Operations: %d\n", stats.Total)
	summary += fmt.Sprintf("Completed:        %d\n", stats.Completed)
	summary += fmt.Sprintf("Failed:           %d\n", stats.Failed)
	summary += fmt.Sprintf("Cancelled:        %d\n", stats.Cancelled)
	summary += fmt.Sprintf("Running:          %d\n", stats.Running)
	summary += fmt.Sprintf("Pending:          %d\n", stats.Pending)
	summary += fmt.Sprintf("Average Duration: %s\n\n", stats.AverageDuration)

	summary += "Recent Operations:\n"
	summary += "==================\n\n"

	// Show last 10 operations
	start := 0
	if len(bh.results) > 10 {
		start = len(bh.results) - 10
	}

	for i := start; i < len(bh.results); i++ {
		result := bh.results[i]
		summary += fmt.Sprintf("%s: %s %s -> %s [%s] (%s)\n",
			result.StartTime.Format("2006-01-02 15:04:05"),
			result.Operation,
			filepath.Base(result.Input),
			filepath.Base(result.Output),
			result.Status,
			result.Duration,
		)
		if result.Error != "" {
			summary += fmt.Sprintf("  Error: %s\n", result.Error)
		}
	}

	return summary
}

// GetReplayInfo returns information needed to replay an operation.
type ReplayInfo struct {
	Operation OperationType
	Input     string
	Output    string
	Metadata  map[string]string
}

// GetReplayInfo returns the information needed to replay a specific operation.
func (bh *BatchHistory) GetReplayInfo(id string) (ReplayInfo, error) {
	result, found := bh.GetByID(id)
	if !found {
		return ReplayInfo{}, fmt.Errorf("operation %s not found", id)
	}

	return ReplayInfo{
		Operation: result.Operation,
		Input:     result.Input,
		Output:    result.Output,
		Metadata:  result.Metadata,
	}, nil
}
