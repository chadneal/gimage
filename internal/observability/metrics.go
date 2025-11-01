package observability

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// ToolMetrics tracks invocation metrics for MCP tools
type ToolMetrics struct {
	mu                sync.RWMutex
	toolInvocations   map[string]*ToolStats
	totalInvocations  int64
	totalSuccesses    int64
	totalFailures     int64
	totalLatencyMs    int64
}

// ToolStats holds statistics for a single tool
type ToolStats struct {
	Name         string
	Invocations  int64
	Successes    int64
	Failures     int64
	TotalLatency time.Duration
	AvgLatency   time.Duration
	MinLatency   time.Duration
	MaxLatency   time.Duration
	LastInvoked  time.Time
}

var (
	// globalMetrics is the singleton metrics instance
	globalMetrics *ToolMetrics
	metricsOnce   sync.Once
)

// GetMetrics returns the global metrics instance
func GetMetrics() *ToolMetrics {
	metricsOnce.Do(func() {
		globalMetrics = &ToolMetrics{
			toolInvocations: make(map[string]*ToolStats),
		}
	})
	return globalMetrics
}

// RecordToolInvocation records a tool invocation with timing and success/failure
func (m *ToolMetrics) RecordToolInvocation(ctx context.Context, toolName string, duration time.Duration, success bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get or create tool stats
	stats, exists := m.toolInvocations[toolName]
	if !exists {
		stats = &ToolStats{
			Name:       toolName,
			MinLatency: duration,
			MaxLatency: duration,
		}
		m.toolInvocations[toolName] = stats
	}

	// Update stats
	stats.Invocations++
	stats.TotalLatency += duration
	stats.AvgLatency = time.Duration(int64(stats.TotalLatency) / stats.Invocations)
	stats.LastInvoked = time.Now()

	if duration < stats.MinLatency {
		stats.MinLatency = duration
	}
	if duration > stats.MaxLatency {
		stats.MaxLatency = duration
	}

	if success {
		stats.Successes++
		m.totalSuccesses++
	} else {
		stats.Failures++
		m.totalFailures++
	}

	m.totalInvocations++
	m.totalLatencyMs += duration.Milliseconds()

	// Log metrics event
	logger := Logger(ctx).With().
		Str("component", "metrics").
		Str("tool", toolName).
		Int64("duration_ms", duration.Milliseconds()).
		Bool("success", success).
		Logger()

	if success {
		logger.Debug().Msg("Tool invocation completed")
	} else {
		logger.Warn().Msg("Tool invocation failed")
	}
}

// GetToolStats returns statistics for a specific tool
func (m *ToolMetrics) GetToolStats(toolName string) *ToolStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if stats, exists := m.toolInvocations[toolName]; exists {
		// Return a copy to avoid race conditions
		statsCopy := *stats
		return &statsCopy
	}
	return nil
}

// GetAllStats returns statistics for all tools
func (m *ToolMetrics) GetAllStats() map[string]*ToolStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy of the map
	result := make(map[string]*ToolStats, len(m.toolInvocations))
	for name, stats := range m.toolInvocations {
		statsCopy := *stats
		result[name] = &statsCopy
	}
	return result
}

// GetSummary returns overall metrics summary
func (m *ToolMetrics) GetSummary() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	avgLatencyMs := int64(0)
	if m.totalInvocations > 0 {
		avgLatencyMs = m.totalLatencyMs / m.totalInvocations
	}

	successRate := float64(0)
	if m.totalInvocations > 0 {
		successRate = float64(m.totalSuccesses) / float64(m.totalInvocations) * 100
	}

	return map[string]interface{}{
		"total_invocations": m.totalInvocations,
		"total_successes":   m.totalSuccesses,
		"total_failures":    m.totalFailures,
		"success_rate_pct":  successRate,
		"avg_latency_ms":    avgLatencyMs,
		"tools_count":       len(m.toolInvocations),
	}
}

// LogSummary logs a summary of all metrics
func (m *ToolMetrics) LogSummary(logger zerolog.Logger) {
	summary := m.GetSummary()

	logger.Info().
		Int64("total_invocations", summary["total_invocations"].(int64)).
		Int64("total_successes", summary["total_successes"].(int64)).
		Int64("total_failures", summary["total_failures"].(int64)).
		Float64("success_rate_pct", summary["success_rate_pct"].(float64)).
		Int64("avg_latency_ms", summary["avg_latency_ms"].(int64)).
		Int("tools_count", summary["tools_count"].(int)).
		Msg("Metrics summary")

	// Log per-tool stats
	allStats := m.GetAllStats()
	for _, stats := range allStats {
		logger.Debug().
			Str("tool", stats.Name).
			Int64("invocations", stats.Invocations).
			Int64("successes", stats.Successes).
			Int64("failures", stats.Failures).
			Int64("avg_latency_ms", stats.AvgLatency.Milliseconds()).
			Int64("min_latency_ms", stats.MinLatency.Milliseconds()).
			Int64("max_latency_ms", stats.MaxLatency.Milliseconds()).
			Time("last_invoked", stats.LastInvoked).
			Msg("Tool metrics")
	}
}

// Reset clears all metrics (useful for testing)
func (m *ToolMetrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.toolInvocations = make(map[string]*ToolStats)
	m.totalInvocations = 0
	m.totalSuccesses = 0
	m.totalFailures = 0
	m.totalLatencyMs = 0
}
