package progress

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"
)

func TestNoOpReporter(t *testing.T) {
	reporter := NewNoOpReporter()
	ctx := context.Background()

	// Should not panic with any operation
	reporter.Start(ctx, "test operation")
	reporter.Update(50, 100, "halfway")
	reporter.Complete("done")
	reporter.Error(nil)
}

func TestLogReporter(t *testing.T) {
	tests := []struct {
		name        string
		verbose     bool
		wantOutput  bool
	}{
		{
			name:       "verbose mode",
			verbose:    true,
			wantOutput: true,
		},
		{
			name:       "silent mode",
			verbose:    false,
			wantOutput: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			reporter := NewLogReporter(buf, tt.verbose)
			ctx := context.Background()

			reporter.Start(ctx, "test operation")
			time.Sleep(150 * time.Millisecond) // Ensure throttle passes
			reporter.Update(50, 100, "halfway")
			reporter.Complete("result")

			output := buf.String()
			if tt.wantOutput {
				if !strings.Contains(output, "Starting") {
					t.Errorf("Expected 'Starting' in output, got: %s", output)
				}
				if !strings.Contains(output, "Progress") {
					t.Errorf("Expected 'Progress' in output, got: %s", output)
				}
				if !strings.Contains(output, "Completed") {
					t.Errorf("Expected 'Completed' in output, got: %s", output)
				}
			} else {
				if output != "" {
					t.Errorf("Expected no output in silent mode, got: %s", output)
				}
			}
		})
	}
}

func TestLogReporterError(t *testing.T) {
	buf := &bytes.Buffer{}
	reporter := NewLogReporter(buf, true)
	ctx := context.Background()

	reporter.Start(ctx, "test operation")
	reporter.Error(context.DeadlineExceeded)

	output := buf.String()
	if !strings.Contains(output, "Error") {
		t.Errorf("Expected 'Error' in output, got: %s", output)
	}
	if !strings.Contains(output, "context deadline exceeded") {
		t.Errorf("Expected error message in output, got: %s", output)
	}
}

func TestLogReporterUpdateThrottling(t *testing.T) {
	buf := &bytes.Buffer{}
	reporter := NewLogReporter(buf, true)
	ctx := context.Background()

	reporter.Start(ctx, "test operation")

	// Send multiple updates rapidly
	for i := 0; i < 10; i++ {
		reporter.Update(int64(i*10), 100, "progress")
	}

	output := buf.String()
	// Should not have 10 progress lines due to throttling
	progressCount := strings.Count(output, "Progress:")
	if progressCount >= 10 {
		t.Errorf("Expected fewer than 10 progress updates due to throttling, got %d", progressCount)
	}
}

func TestTUIReporter(t *testing.T) {
	var startCalled, updateCalled, completeCalled, errorCalled bool
	var lastOperation string
	var lastPercentage float64

	reporter := NewTUIReporter(
		func(operation string) {
			startCalled = true
			lastOperation = operation
		},
		func(current, total int64, message string, percentage float64) {
			updateCalled = true
			lastPercentage = percentage
		},
		func(result interface{}, duration time.Duration) {
			completeCalled = true
		},
		func(err error, duration time.Duration) {
			errorCalled = true
		},
	)

	ctx := context.Background()

	reporter.Start(ctx, "test operation")
	if !startCalled || lastOperation != "test operation" {
		t.Errorf("Start callback not called properly")
	}

	reporter.Update(50, 100, "halfway")
	if !updateCalled || lastPercentage != 50.0 {
		t.Errorf("Update callback not called properly, got percentage: %f", lastPercentage)
	}

	reporter.Complete("done")
	if !completeCalled {
		t.Errorf("Complete callback not called")
	}

	reporter.Error(context.Canceled)
	if !errorCalled {
		t.Errorf("Error callback not called")
	}
}

func TestTUIReporterNilCallbacks(t *testing.T) {
	// Should not panic with nil callbacks
	reporter := NewTUIReporter(nil, nil, nil, nil)
	ctx := context.Background()

	reporter.Start(ctx, "test")
	reporter.Update(50, 100, "test")
	reporter.Complete("test")
	reporter.Error(nil)
}

func TestContextReporter(t *testing.T) {
	// Test WithReporter and FromContext
	baseCtx := context.Background()
	reporter := NewNoOpReporter()

	ctx := WithReporter(baseCtx, reporter)
	retrieved := FromContext(ctx)

	if retrieved != reporter {
		t.Errorf("Retrieved reporter doesn't match stored reporter")
	}

	// Test FromContext with no reporter
	emptyReporter := FromContext(baseCtx)
	if _, ok := emptyReporter.(*NoOpReporter); !ok {
		t.Errorf("Expected NoOpReporter from context without reporter")
	}
}

func TestLogReporterPercentageCalculation(t *testing.T) {
	buf := &bytes.Buffer{}
	reporter := NewLogReporter(buf, true)
	ctx := context.Background()

	reporter.Start(ctx, "test")
	time.Sleep(150 * time.Millisecond)
	reporter.Update(25, 100, "quarter")

	output := buf.String()
	if !strings.Contains(output, "25.0%") {
		t.Errorf("Expected '25.0%%' in output, got: %s", output)
	}
}

func TestLogReporterNoTotal(t *testing.T) {
	buf := &bytes.Buffer{}
	reporter := NewLogReporter(buf, true)
	ctx := context.Background()

	reporter.Start(ctx, "test")
	time.Sleep(150 * time.Millisecond)
	reporter.Update(50, 0, "no total")

	output := buf.String()
	if !strings.Contains(output, "Progress: 50") {
		t.Errorf("Expected 'Progress: 50' in output when total is 0, got: %s", output)
	}
}
