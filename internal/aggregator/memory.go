package aggregator

import (
	"sync"
	"time"

	"github.com/kpblcaoo/sboxagent/internal/logger"
)

// LogLevel represents the log level
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

// LogEntry represents a log entry
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     LogLevel               `json:"level"`
	Message   string                 `json:"message"`
	Source    string                 `json:"source"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	ID        string                 `json:"id"`
}

// MemoryAggregator represents an in-memory log aggregator
type MemoryAggregator struct {
	logger *logger.Logger

	// Configuration
	maxEntries int
	maxAge     time.Duration

	// Storage
	mu      sync.RWMutex
	entries []LogEntry
	index   int // Current position in circular buffer
	count   int // Total number of entries added

	// Statistics
	statsMu sync.RWMutex
	stats   AggregatorStats
}

// AggregatorStats holds aggregator statistics
type AggregatorStats struct {
	TotalEntries   int64
	DroppedEntries int64
	CurrentEntries int64
	OldestEntry    time.Time
	NewestEntry    time.Time
	StartTime      time.Time
}

// NewMemoryAggregator creates a new memory aggregator
func NewMemoryAggregator(log *logger.Logger, maxEntries int, maxAge time.Duration) *MemoryAggregator {
	return &MemoryAggregator{
		logger:     log,
		maxEntries: maxEntries,
		maxAge:     maxAge,
		entries:    make([]LogEntry, maxEntries),
		stats: AggregatorStats{
			StartTime: time.Now(),
		},
	}
}

// Add adds a log entry to the aggregator
func (a *MemoryAggregator) Add(entry LogEntry) {
	// Set timestamp if not set
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	// Generate ID if not provided
	if entry.ID == "" {
		entry.ID = generateLogID(entry)
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	// Update statistics
	a.statsMu.Lock()
	a.stats.TotalEntries++
	a.stats.NewestEntry = entry.Timestamp
	if a.stats.OldestEntry.IsZero() {
		a.stats.OldestEntry = entry.Timestamp
	}
	a.statsMu.Unlock()

	// Add entry to circular buffer
	a.entries[a.index] = entry
	a.index = (a.index + 1) % a.maxEntries
	a.count++

	// Update current entries count
	if a.count < a.maxEntries {
		a.statsMu.Lock()
		a.stats.CurrentEntries = int64(a.count)
		a.statsMu.Unlock()
	} else {
		a.statsMu.Lock()
		a.stats.CurrentEntries = int64(a.maxEntries)
		// Update oldest entry
		oldestIndex := (a.index) % a.maxEntries
		a.stats.OldestEntry = a.entries[oldestIndex].Timestamp
		a.statsMu.Unlock()
	}

	// Cleanup old entries if maxAge is set
	if a.maxAge > 0 {
		go a.cleanupOldEntries()
	}
}

// GetEntries returns log entries with optional filtering
func (a *MemoryAggregator) GetEntries(limit int, level LogLevel, since time.Time) []LogEntry {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if limit <= 0 {
		limit = a.maxEntries
	}

	var result []LogEntry
	count := 0

	// Determine how many entries to check
	entriesToCheck := a.count
	if entriesToCheck > a.maxEntries {
		entriesToCheck = a.maxEntries
	}

	// Start from the most recent entry
	startIndex := (a.index - 1 + a.maxEntries) % a.maxEntries

	for i := 0; i < entriesToCheck && count < limit; i++ {
		idx := (startIndex - i + a.maxEntries) % a.maxEntries
		entry := a.entries[idx]

		// Skip if entry is zero (not yet filled)
		if entry.Timestamp.IsZero() {
			continue
		}

		// Apply filters
		if level != "" && entry.Level != level {
			continue
		}

		if !since.IsZero() && entry.Timestamp.Before(since) {
			continue
		}

		result = append(result, entry)
		count++
	}

	return result
}

// GetEntriesByLevel returns entries filtered by level
func (a *MemoryAggregator) GetEntriesByLevel(level LogLevel, limit int) []LogEntry {
	return a.GetEntries(limit, level, time.Time{})
}

// GetRecentEntries returns the most recent entries
func (a *MemoryAggregator) GetRecentEntries(limit int) []LogEntry {
	return a.GetEntries(limit, "", time.Time{})
}

// GetEntriesSince returns entries since a specific time
func (a *MemoryAggregator) GetEntriesSince(since time.Time, limit int) []LogEntry {
	return a.GetEntries(limit, "", since)
}

// GetStats returns aggregator statistics
func (a *MemoryAggregator) GetStats() AggregatorStats {
	a.statsMu.RLock()
	defer a.statsMu.RUnlock()
	return a.stats
}

// GetTotalEntries returns the total number of entries
func (a *AggregatorStats) GetTotalEntries() int64 {
	return a.TotalEntries
}

// GetDroppedEntries returns the number of dropped entries
func (a *AggregatorStats) GetDroppedEntries() int64 {
	return a.DroppedEntries
}

// GetCurrentEntries returns the current number of entries
func (a *AggregatorStats) GetCurrentEntries() int64 {
	return a.CurrentEntries
}

// GetNewestEntry returns the newest entry timestamp
func (a *AggregatorStats) GetNewestEntry() time.Time {
	return a.NewestEntry
}

// Clear clears all entries
func (a *MemoryAggregator) Clear() {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Reset entries
	for i := range a.entries {
		a.entries[i] = LogEntry{}
	}

	a.index = 0
	a.count = 0

	// Reset statistics
	a.statsMu.Lock()
	a.stats.CurrentEntries = 0
	a.stats.OldestEntry = time.Time{}
	a.stats.NewestEntry = time.Time{}
	a.statsMu.Unlock()

	a.logger.Info("Memory aggregator cleared", map[string]interface{}{})
}

// cleanupOldEntries removes entries older than maxAge
func (a *MemoryAggregator) cleanupOldEntries() {
	if a.maxAge <= 0 {
		return
	}

	cutoff := time.Now().Add(-a.maxAge)

	a.mu.Lock()
	defer a.mu.Unlock()

	dropped := 0
	for i := 0; i < a.count; i++ {
		idx := (a.index - a.count + i + a.maxEntries) % a.maxEntries
		if a.entries[idx].Timestamp.Before(cutoff) {
			a.entries[idx] = LogEntry{}
			dropped++
		}
	}

	if dropped > 0 {
		a.statsMu.Lock()
		a.stats.DroppedEntries += int64(dropped)
		a.stats.CurrentEntries -= int64(dropped)
		a.statsMu.Unlock()

		a.logger.Debug("Cleaned up old log entries", map[string]interface{}{
			"dropped": dropped,
			"cutoff":  cutoff,
		})
	}
}

// GetLevelCounts returns count of entries by level
func (a *MemoryAggregator) GetLevelCounts() map[LogLevel]int {
	a.mu.RLock()
	defer a.mu.RUnlock()

	counts := make(map[LogLevel]int)

	for i := 0; i < a.count; i++ {
		idx := (a.index - a.count + i + a.maxEntries) % a.maxEntries
		entry := a.entries[idx]

		if !entry.Timestamp.IsZero() {
			counts[entry.Level]++
		}
	}

	return counts
}

// Search searches for entries containing specific text
func (a *MemoryAggregator) Search(query string, limit int) []LogEntry {
	if query == "" {
		return a.GetRecentEntries(limit)
	}

	a.mu.RLock()
	defer a.mu.RUnlock()

	var result []LogEntry
	count := 0

	// Start from the most recent entry
	startIndex := (a.index - 1 + a.maxEntries) % a.maxEntries

	for i := 0; i < a.count && count < limit; i++ {
		idx := (startIndex - i + a.maxEntries) % a.maxEntries
		entry := a.entries[idx]

		// Skip if entry is zero (not yet filled)
		if entry.Timestamp.IsZero() {
			continue
		}

		// Simple string search in message
		if contains(entry.Message, query) {
			result = append(result, entry)
			count++
		}
	}

	return result
}

// generateLogID generates a unique ID for a log entry
func generateLogID(entry LogEntry) string {
	return entry.Timestamp.Format("20060102-150405.000000000")
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsSubstring(s, substr)))
}

// containsSubstring is a simple substring search
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
