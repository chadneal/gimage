// Package progress provides interfaces and implementations for tracking
// long-running operations with progress reporting capabilities.
package progress

import (
	"context"
	"fmt"
	"io"
	"time"
)

// ProgressReporter is an interface for reporting progress of long-running operations.
// Implementations can provide different output mechanisms (TUI, CLI, logging, etc.).
type ProgressReporter interface {
	// Start initiates a progress tracking session for an operation.
	// The operation string describes what is being tracked.
	Start(ctx context.Context, operation string)

	// Update reports progress during an operation.
	// current and total represent the progress (e.g., bytes processed, items completed).
	// message provides additional context about the current state.
	Update(current, total int64, message string)

	// Complete marks the operation as successfully finished.
	// result can contain information about the final outcome.
	Complete(result interface{})

	// Error reports that the operation failed with an error.
	Error(err error)
}

// NoOpReporter is a silent reporter that does nothing.
// This is the default reporter for CLI operations that don't need progress output.
type NoOpReporter struct{}

// NewNoOpReporter creates a new no-op reporter.
func NewNoOpReporter() *NoOpReporter {
	return &NoOpReporter{}
}

func (r *NoOpReporter) Start(ctx context.Context, operation string)       {}
func (r *NoOpReporter) Update(current, total int64, message string)       {}
func (r *NoOpReporter) Complete(result interface{})                       {}
func (r *NoOpReporter) Error(err error)                                   {}

// LogReporter writes progress updates to a writer (typically os.Stderr or os.Stdout).
// This is used for CLI verbose output.
type LogReporter struct {
	writer    io.Writer
	operation string
	startTime time.Time
	lastUpdate time.Time
	verbose   bool
}

// NewLogReporter creates a new log-based reporter.
func NewLogReporter(writer io.Writer, verbose bool) *LogReporter {
	return &LogReporter{
		writer:  writer,
		verbose: verbose,
	}
}

func (r *LogReporter) Start(ctx context.Context, operation string) {
	r.operation = operation
	r.startTime = time.Now()
	r.lastUpdate = r.startTime
	if r.verbose {
		fmt.Fprintf(r.writer, "Starting: %s\n", operation)
	}
}

func (r *LogReporter) Update(current, total int64, message string) {
	if !r.verbose {
		return
	}

	// Throttle updates to once per 100ms to avoid spamming output
	now := time.Now()
	if now.Sub(r.lastUpdate) < 100*time.Millisecond {
		return
	}
	r.lastUpdate = now

	if total > 0 {
		percentage := float64(current) / float64(total) * 100
		fmt.Fprintf(r.writer, "Progress: %.1f%% (%d/%d) - %s\n", percentage, current, total, message)
	} else {
		fmt.Fprintf(r.writer, "Progress: %d - %s\n", current, message)
	}
}

func (r *LogReporter) Complete(result interface{}) {
	if !r.verbose {
		return
	}

	elapsed := time.Since(r.startTime)
	if result != nil {
		fmt.Fprintf(r.writer, "Completed: %s (took %s) - Result: %v\n", r.operation, elapsed, result)
	} else {
		fmt.Fprintf(r.writer, "Completed: %s (took %s)\n", r.operation, elapsed)
	}
}

func (r *LogReporter) Error(err error) {
	elapsed := time.Since(r.startTime)
	fmt.Fprintf(r.writer, "Error: %s (after %s) - %v\n", r.operation, elapsed, err)
}

// TUIReporter is used by the TUI to update progress displays via callbacks.
// It allows the TUI to render progress bars, spinners, and status updates.
type TUIReporter struct {
	operation string
	startTime time.Time

	// Callbacks for TUI to handle progress events
	OnStart    func(operation string)
	OnUpdate   func(current, total int64, message string, percentage float64)
	OnComplete func(result interface{}, duration time.Duration)
	OnError    func(err error, duration time.Duration)
}

// NewTUIReporter creates a new TUI reporter with the given callbacks.
func NewTUIReporter(
	onStart func(operation string),
	onUpdate func(current, total int64, message string, percentage float64),
	onComplete func(result interface{}, duration time.Duration),
	onError func(err error, duration time.Duration),
) *TUIReporter {
	return &TUIReporter{
		OnStart:    onStart,
		OnUpdate:   onUpdate,
		OnComplete: onComplete,
		OnError:    onError,
	}
}

func (r *TUIReporter) Start(ctx context.Context, operation string) {
	r.operation = operation
	r.startTime = time.Now()
	if r.OnStart != nil {
		r.OnStart(operation)
	}
}

func (r *TUIReporter) Update(current, total int64, message string) {
	if r.OnUpdate == nil {
		return
	}

	percentage := 0.0
	if total > 0 {
		percentage = float64(current) / float64(total) * 100
	}

	r.OnUpdate(current, total, message, percentage)
}

func (r *TUIReporter) Complete(result interface{}) {
	if r.OnComplete == nil {
		return
	}

	elapsed := time.Since(r.startTime)
	r.OnComplete(result, elapsed)
}

func (r *TUIReporter) Error(err error) {
	if r.OnError == nil {
		return
	}

	elapsed := time.Since(r.startTime)
	r.OnError(err, elapsed)
}

// ContextKey type for context values
type contextKey string

const reporterKey contextKey = "progress_reporter"

// WithReporter returns a new context with the given reporter attached.
func WithReporter(ctx context.Context, reporter ProgressReporter) context.Context {
	return context.WithValue(ctx, reporterKey, reporter)
}

// FromContext retrieves a reporter from the context.
// If no reporter is found, it returns a NoOpReporter.
func FromContext(ctx context.Context) ProgressReporter {
	if reporter, ok := ctx.Value(reporterKey).(ProgressReporter); ok {
		return reporter
	}
	return NewNoOpReporter()
}
