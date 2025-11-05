package batch

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewBatchHistory(t *testing.T) {
	bh := NewBatchHistory(100)
	if bh == nil {
		t.Fatal("NewBatchHistory returned nil")
	}
	if bh.Count() != 0 {
		t.Errorf("Expected empty history, got %d items", bh.Count())
	}
	if bh.maxSize != 100 {
		t.Errorf("Expected maxSize 100, got %d", bh.maxSize)
	}
}

func TestBatchHistoryAdd(t *testing.T) {
	bh := NewBatchHistory(0)

	result := OperationResult{
		Operation: OpTypeResize,
		Status:    StatusCompleted,
		Input:     "/path/to/input.png",
		Output:    "/path/to/output.png",
		StartTime: time.Now(),
		EndTime:   time.Now().Add(2 * time.Second),
	}

	bh.Add(result)

	if bh.Count() != 1 {
		t.Errorf("Expected 1 result, got %d", bh.Count())
	}

	results := bh.GetAll()
	if len(results) != 1 {
		t.Errorf("Expected 1 result from GetAll, got %d", len(results))
	}

	// Check that ID was generated
	if results[0].ID == "" {
		t.Error("Expected ID to be generated")
	}

	// Check that duration was calculated
	if results[0].Duration == 0 {
		t.Error("Expected duration to be calculated")
	}
}

func TestBatchHistoryMaxSize(t *testing.T) {
	bh := NewBatchHistory(5)

	// Add 10 results
	for i := 0; i < 10; i++ {
		result := OperationResult{
			Operation: OpTypeResize,
			Status:    StatusCompleted,
			Input:     "/path/to/input.png",
			Output:    "/path/to/output.png",
			StartTime: time.Now(),
			EndTime:   time.Now(),
		}
		bh.Add(result)
	}

	// Should only keep 5 most recent
	if bh.Count() != 5 {
		t.Errorf("Expected 5 results (max size), got %d", bh.Count())
	}
}

func TestBatchHistoryGetByStatus(t *testing.T) {
	bh := NewBatchHistory(0)

	// Add results with different statuses
	statuses := []OperationStatus{
		StatusCompleted,
		StatusFailed,
		StatusCompleted,
		StatusCancelled,
		StatusCompleted,
	}

	for _, status := range statuses {
		result := OperationResult{
			Operation: OpTypeResize,
			Status:    status,
			Input:     "/path/to/input.png",
			Output:    "/path/to/output.png",
			StartTime: time.Now(),
			EndTime:   time.Now(),
		}
		bh.Add(result)
	}

	// Test filtering
	completed := bh.GetByStatus(StatusCompleted)
	if len(completed) != 3 {
		t.Errorf("Expected 3 completed results, got %d", len(completed))
	}

	failed := bh.GetByStatus(StatusFailed)
	if len(failed) != 1 {
		t.Errorf("Expected 1 failed result, got %d", len(failed))
	}

	cancelled := bh.GetByStatus(StatusCancelled)
	if len(cancelled) != 1 {
		t.Errorf("Expected 1 cancelled result, got %d", len(cancelled))
	}
}

func TestBatchHistoryGetByOperation(t *testing.T) {
	bh := NewBatchHistory(0)

	// Add results with different operation types
	operations := []OperationType{
		OpTypeResize,
		OpTypeCrop,
		OpTypeResize,
		OpTypeCompress,
		OpTypeResize,
	}

	for _, op := range operations {
		result := OperationResult{
			Operation: op,
			Status:    StatusCompleted,
			Input:     "/path/to/input.png",
			Output:    "/path/to/output.png",
			StartTime: time.Now(),
			EndTime:   time.Now(),
		}
		bh.Add(result)
	}

	// Test filtering
	resizes := bh.GetByOperation(OpTypeResize)
	if len(resizes) != 3 {
		t.Errorf("Expected 3 resize operations, got %d", len(resizes))
	}

	crops := bh.GetByOperation(OpTypeCrop)
	if len(crops) != 1 {
		t.Errorf("Expected 1 crop operation, got %d", len(crops))
	}
}

func TestBatchHistoryGetByID(t *testing.T) {
	bh := NewBatchHistory(0)

	result := OperationResult{
		ID:        "test-123",
		Operation: OpTypeResize,
		Status:    StatusCompleted,
		Input:     "/path/to/input.png",
		Output:    "/path/to/output.png",
		StartTime: time.Now(),
		EndTime:   time.Now(),
	}

	bh.Add(result)

	// Test finding by ID
	found, ok := bh.GetByID("test-123")
	if !ok {
		t.Error("Expected to find result by ID")
	}
	if found.ID != "test-123" {
		t.Errorf("Expected ID 'test-123', got '%s'", found.ID)
	}

	// Test not finding invalid ID
	_, ok = bh.GetByID("invalid-id")
	if ok {
		t.Error("Expected not to find invalid ID")
	}
}

func TestBatchHistoryGetStats(t *testing.T) {
	bh := NewBatchHistory(0)

	// Add various results
	results := []OperationResult{
		{Operation: OpTypeResize, Status: StatusCompleted, StartTime: time.Now(), EndTime: time.Now().Add(1 * time.Second)},
		{Operation: OpTypeCrop, Status: StatusCompleted, StartTime: time.Now(), EndTime: time.Now().Add(2 * time.Second)},
		{Operation: OpTypeCompress, Status: StatusFailed, StartTime: time.Now(), EndTime: time.Now().Add(1 * time.Second)},
		{Operation: OpTypeConvert, Status: StatusCancelled, StartTime: time.Now(), EndTime: time.Now().Add(1 * time.Second)},
		{Operation: OpTypeResize, Status: StatusRunning, StartTime: time.Now(), EndTime: time.Time{}},
		{Operation: OpTypeCrop, Status: StatusPending, StartTime: time.Now(), EndTime: time.Time{}},
	}

	for _, result := range results {
		bh.Add(result)
	}

	stats := bh.GetStats()

	if stats.Total != 6 {
		t.Errorf("Expected 6 total, got %d", stats.Total)
	}
	if stats.Completed != 2 {
		t.Errorf("Expected 2 completed, got %d", stats.Completed)
	}
	if stats.Failed != 1 {
		t.Errorf("Expected 1 failed, got %d", stats.Failed)
	}
	if stats.Cancelled != 1 {
		t.Errorf("Expected 1 cancelled, got %d", stats.Cancelled)
	}
	if stats.Running != 1 {
		t.Errorf("Expected 1 running, got %d", stats.Running)
	}
	if stats.Pending != 1 {
		t.Errorf("Expected 1 pending, got %d", stats.Pending)
	}
}

func TestBatchHistoryClear(t *testing.T) {
	bh := NewBatchHistory(0)

	// Add some results
	for i := 0; i < 5; i++ {
		result := OperationResult{
			Operation: OpTypeResize,
			Status:    StatusCompleted,
			Input:     "/path/to/input.png",
			Output:    "/path/to/output.png",
			StartTime: time.Now(),
			EndTime:   time.Now(),
		}
		bh.Add(result)
	}

	if bh.Count() != 5 {
		t.Errorf("Expected 5 results before clear, got %d", bh.Count())
	}

	bh.Clear()

	if bh.Count() != 0 {
		t.Errorf("Expected 0 results after clear, got %d", bh.Count())
	}
}

func TestBatchHistorySaveAndLoad(t *testing.T) {
	bh := NewBatchHistory(0)

	// Add some results
	results := []OperationResult{
		{
			ID:        "test-1",
			Operation: OpTypeResize,
			Status:    StatusCompleted,
			Input:     "/path/to/input1.png",
			Output:    "/path/to/output1.png",
			StartTime: time.Now(),
			EndTime:   time.Now().Add(1 * time.Second),
		},
		{
			ID:        "test-2",
			Operation: OpTypeCrop,
			Status:    StatusFailed,
			Input:     "/path/to/input2.png",
			Output:    "/path/to/output2.png",
			StartTime: time.Now(),
			EndTime:   time.Now().Add(2 * time.Second),
			Error:     "test error",
		},
	}

	for _, result := range results {
		bh.Add(result)
	}

	// Save to file
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history.json")

	if err := bh.SaveToFile(historyFile); err != nil {
		t.Fatalf("SaveToFile failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(historyFile); os.IsNotExist(err) {
		t.Fatal("History file was not created")
	}

	// Load from file
	bh2 := NewBatchHistory(0)
	if err := bh2.LoadFromFile(historyFile); err != nil {
		t.Fatalf("LoadFromFile failed: %v", err)
	}

	// Verify loaded data
	if bh2.Count() != 2 {
		t.Errorf("Expected 2 results after load, got %d", bh2.Count())
	}

	loaded := bh2.GetAll()
	if len(loaded) != 2 {
		t.Fatalf("Expected 2 results in GetAll, got %d", len(loaded))
	}

	// Check first result
	if loaded[0].ID != "test-1" {
		t.Errorf("Expected ID 'test-1', got '%s'", loaded[0].ID)
	}
	if loaded[0].Operation != OpTypeResize {
		t.Errorf("Expected operation 'resize', got '%s'", loaded[0].Operation)
	}

	// Check second result
	if loaded[1].ID != "test-2" {
		t.Errorf("Expected ID 'test-2', got '%s'", loaded[1].ID)
	}
	if loaded[1].Error != "test error" {
		t.Errorf("Expected error 'test error', got '%s'", loaded[1].Error)
	}
}

func TestBatchHistoryExportSummary(t *testing.T) {
	bh := NewBatchHistory(0)

	// Add some results
	for i := 0; i < 3; i++ {
		result := OperationResult{
			Operation: OpTypeResize,
			Status:    StatusCompleted,
			Input:     "/path/to/input.png",
			Output:    "/path/to/output.png",
			StartTime: time.Now(),
			EndTime:   time.Now().Add(1 * time.Second),
		}
		bh.Add(result)
	}

	summary := bh.ExportSummary()

	if summary == "" {
		t.Error("Expected non-empty summary")
	}

	// Summary should contain key information
	if !contains(summary, "Total Operations:") {
		t.Error("Summary missing 'Total Operations:'")
	}
	if !contains(summary, "Completed:") {
		t.Error("Summary missing 'Completed:'")
	}
	if !contains(summary, "Recent Operations:") {
		t.Error("Summary missing 'Recent Operations:'")
	}
}

func TestBatchHistoryGetReplayInfo(t *testing.T) {
	bh := NewBatchHistory(0)

	metadata := map[string]string{
		"width":  "800",
		"height": "600",
	}

	result := OperationResult{
		ID:        "test-replay",
		Operation: OpTypeResize,
		Status:    StatusCompleted,
		Input:     "/path/to/input.png",
		Output:    "/path/to/output.png",
		StartTime: time.Now(),
		EndTime:   time.Now(),
		Metadata:  metadata,
	}

	bh.Add(result)

	// Test getting replay info
	info, err := bh.GetReplayInfo("test-replay")
	if err != nil {
		t.Fatalf("GetReplayInfo failed: %v", err)
	}

	if info.Operation != OpTypeResize {
		t.Errorf("Expected operation 'resize', got '%s'", info.Operation)
	}
	if info.Input != "/path/to/input.png" {
		t.Errorf("Expected input '/path/to/input.png', got '%s'", info.Input)
	}
	if len(info.Metadata) != 2 {
		t.Errorf("Expected 2 metadata items, got %d", len(info.Metadata))
	}

	// Test invalid ID
	_, err = bh.GetReplayInfo("invalid-id")
	if err == nil {
		t.Error("Expected error for invalid ID")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
		findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
