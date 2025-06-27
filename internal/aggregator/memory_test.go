package aggregator

import (
	"fmt"
	"testing"
	"time"

	"github.com/kpblcaoo/sboxagent/internal/logger"
)

func TestNewMemoryAggregator(t *testing.T) {
	log, _ := logger.New("debug")
	aggregator := NewMemoryAggregator(log, 100, 1*time.Hour)

	if aggregator == nil {
		t.Fatal("Expected aggregator to be created")
	}

	if aggregator.logger != log {
		t.Error("Expected logger to be set")
	}

	if aggregator.maxEntries != 100 {
		t.Errorf("Expected maxEntries to be 100, got %d", aggregator.maxEntries)
	}

	if aggregator.maxAge != 1*time.Hour {
		t.Errorf("Expected maxAge to be 1 hour, got %v", aggregator.maxAge)
	}
}

func TestMemoryAggregator_Add(t *testing.T) {
	log, _ := logger.New("debug")
	aggregator := NewMemoryAggregator(log, 10, 0)

	// Add entries
	entry1 := LogEntry{
		Level:   LogLevelInfo,
		Message: "test message 1",
		Source:  "test",
	}

	entry2 := LogEntry{
		Level:   LogLevelError,
		Message: "test message 2",
		Source:  "test",
	}

	aggregator.Add(entry1)
	aggregator.Add(entry2)

	// Check stats
	stats := aggregator.GetStats()
	if stats.TotalEntries != 2 {
		t.Errorf("Expected 2 total entries, got %d", stats.TotalEntries)
	}

	if stats.CurrentEntries != 2 {
		t.Errorf("Expected 2 current entries, got %d", stats.CurrentEntries)
	}
}

func TestMemoryAggregator_CircularBuffer(t *testing.T) {
	log, _ := logger.New("debug")
	aggregator := NewMemoryAggregator(log, 3, 0)

	// Add more entries than buffer size
	for i := 0; i < 5; i++ {
		entry := LogEntry{
			Level:   LogLevelInfo,
			Message: fmt.Sprintf("test message %d", i),
			Source:  "test",
		}
		aggregator.Add(entry)
	}

	// Check stats
	stats := aggregator.GetStats()
	if stats.TotalEntries != 5 {
		t.Errorf("Expected 5 total entries, got %d", stats.TotalEntries)
	}

	if stats.CurrentEntries != 3 {
		t.Errorf("Expected 3 current entries, got %d", stats.CurrentEntries)
	}

	// Check that only the last 3 entries are kept
	entries := aggregator.GetRecentEntries(10)
	if len(entries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(entries))
	}

	// Check that the entries are in reverse chronological order (newest first)
	expectedMessages := []string{"test message 4", "test message 3", "test message 2"}
	for i, expected := range expectedMessages {
		if i < len(entries) && entries[i].Message != expected {
			t.Errorf("Expected entry %d to be '%s', got '%s'", i, expected, entries[i].Message)
		}
	}
}

func TestMemoryAggregator_GetEntriesByLevel(t *testing.T) {
	log, _ := logger.New("debug")
	aggregator := NewMemoryAggregator(log, 10, 0)

	// Add entries with different levels
	aggregator.Add(LogEntry{Level: LogLevelInfo, Message: "info 1", Source: "test"})
	aggregator.Add(LogEntry{Level: LogLevelError, Message: "error 1", Source: "test"})
	aggregator.Add(LogEntry{Level: LogLevelInfo, Message: "info 2", Source: "test"})
	aggregator.Add(LogEntry{Level: LogLevelWarn, Message: "warn 1", Source: "test"})

	// Get info entries
	infoEntries := aggregator.GetEntriesByLevel(LogLevelInfo, 10)
	if len(infoEntries) != 2 {
		t.Errorf("Expected 2 info entries, got %d", len(infoEntries))
	}

	// Get error entries
	errorEntries := aggregator.GetEntriesByLevel(LogLevelError, 10)
	if len(errorEntries) != 1 {
		t.Errorf("Expected 1 error entry, got %d", len(errorEntries))
	}

	// Get warn entries
	warnEntries := aggregator.GetEntriesByLevel(LogLevelWarn, 10)
	if len(warnEntries) != 1 {
		t.Errorf("Expected 1 warn entry, got %d", len(warnEntries))
	}
}

func TestMemoryAggregator_GetEntriesSince(t *testing.T) {
	log, _ := logger.New("debug")
	aggregator := NewMemoryAggregator(log, 10, 0)

	// Add entries with different timestamps
	now := time.Now()
	oldTime := now.Add(-2 * time.Hour)
	recentTime := now.Add(-30 * time.Minute)

	aggregator.Add(LogEntry{
		Level:     LogLevelInfo,
		Message:   "old entry",
		Source:    "test",
		Timestamp: oldTime,
	})

	aggregator.Add(LogEntry{
		Level:     LogLevelInfo,
		Message:   "recent entry",
		Source:    "test",
		Timestamp: recentTime,
	})

	// Get entries since 1 hour ago
	since := now.Add(-1 * time.Hour)
	entries := aggregator.GetEntriesSince(since, 10)
	if len(entries) != 1 {
		t.Errorf("Expected 1 recent entry, got %d", len(entries))
	}

	if entries[0].Message != "recent entry" {
		t.Errorf("Expected 'recent entry', got %s", entries[0].Message)
	}
}

func TestMemoryAggregator_GetLevelCounts(t *testing.T) {
	log, _ := logger.New("debug")
	aggregator := NewMemoryAggregator(log, 10, 0)

	// Add entries with different levels
	aggregator.Add(LogEntry{Level: LogLevelInfo, Message: "info 1", Source: "test"})
	aggregator.Add(LogEntry{Level: LogLevelInfo, Message: "info 2", Source: "test"})
	aggregator.Add(LogEntry{Level: LogLevelError, Message: "error 1", Source: "test"})
	aggregator.Add(LogEntry{Level: LogLevelWarn, Message: "warn 1", Source: "test"})

	// Get level counts
	counts := aggregator.GetLevelCounts()
	if counts[LogLevelInfo] != 2 {
		t.Errorf("Expected 2 info entries, got %d", counts[LogLevelInfo])
	}

	if counts[LogLevelError] != 1 {
		t.Errorf("Expected 1 error entry, got %d", counts[LogLevelError])
	}

	if counts[LogLevelWarn] != 1 {
		t.Errorf("Expected 1 warn entry, got %d", counts[LogLevelWarn])
	}
}

func TestMemoryAggregator_Search(t *testing.T) {
	log, _ := logger.New("debug")
	aggregator := NewMemoryAggregator(log, 10, 0)

	// Add entries with different messages
	aggregator.Add(LogEntry{Level: LogLevelInfo, Message: "hello world", Source: "test"})
	aggregator.Add(LogEntry{Level: LogLevelInfo, Message: "goodbye world", Source: "test"})
	aggregator.Add(LogEntry{Level: LogLevelInfo, Message: "hello there", Source: "test"})

	// Search for "hello"
	entries := aggregator.Search("hello", 10)
	if len(entries) != 2 {
		t.Errorf("Expected 2 entries with 'hello', got %d", len(entries))
	}

	// Search for "world"
	entries = aggregator.Search("world", 10)
	if len(entries) != 2 {
		t.Errorf("Expected 2 entries with 'world', got %d", len(entries))
	}

	// Search for non-existent term
	entries = aggregator.Search("nonexistent", 10)
	if len(entries) != 0 {
		t.Errorf("Expected 0 entries for 'nonexistent', got %d", len(entries))
	}
}

func TestMemoryAggregator_Clear(t *testing.T) {
	log, _ := logger.New("debug")
	aggregator := NewMemoryAggregator(log, 10, 0)

	// Add some entries
	aggregator.Add(LogEntry{Level: LogLevelInfo, Message: "test", Source: "test"})
	aggregator.Add(LogEntry{Level: LogLevelError, Message: "test", Source: "test"})

	// Clear
	aggregator.Clear()

	// Check stats
	stats := aggregator.GetStats()
	if stats.CurrentEntries != 0 {
		t.Errorf("Expected 0 current entries after clear, got %d", stats.CurrentEntries)
	}

	// Check entries
	entries := aggregator.GetRecentEntries(10)
	if len(entries) != 0 {
		t.Errorf("Expected 0 entries after clear, got %d", len(entries))
	}
}

func TestMemoryAggregator_MaxAge(t *testing.T) {
	log, _ := logger.New("debug")
	aggregator := NewMemoryAggregator(log, 10, 1*time.Hour)

	// Add an old entry
	oldTime := time.Now().Add(-2 * time.Hour)
	aggregator.Add(LogEntry{
		Level:     LogLevelInfo,
		Message:   "old entry",
		Source:    "test",
		Timestamp: oldTime,
	})

	// Add a recent entry
	aggregator.Add(LogEntry{
		Level:   LogLevelInfo,
		Message: "recent entry",
		Source:  "test",
	})

	// Trigger cleanup
	aggregator.cleanupOldEntries()

	// Check that only recent entry remains
	entries := aggregator.GetRecentEntries(10)
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry after cleanup, got %d", len(entries))
	}

	if entries[0].Message != "recent entry" {
		t.Errorf("Expected 'recent entry', got %s", entries[0].Message)
	}
}
