package dispatcher

import (
	"context"
	"testing"
	"time"

	"github.com/kpblcaoo/sboxagent/internal/logger"
	"github.com/kpblcaoo/sboxagent/internal/services"
)

func TestNewDispatcher(t *testing.T) {
	log, _ := logger.New("debug")
	dispatcher := NewDispatcher(log)

	if dispatcher == nil {
		t.Fatal("Expected dispatcher to be created")
	}

	if dispatcher.logger != log {
		t.Error("Expected logger to be set")
	}

	if len(dispatcher.handlers) != 0 {
		t.Error("Expected empty handlers map")
	}
}

func TestDispatcher_StartStop(t *testing.T) {
	log, _ := logger.New("debug")
	dispatcher := NewDispatcher(log)
	ctx := context.Background()

	// Test start
	err := dispatcher.Start(ctx)
	if err != nil {
		t.Fatalf("Expected no error on start, got: %v", err)
	}

	// Test double start
	err = dispatcher.Start(ctx)
	if err == nil {
		t.Error("Expected error on double start")
	}

	// Test stop
	dispatcher.Stop()

	// Test stop again (should not panic)
	dispatcher.Stop()
}

func TestDispatcher_RegisterHandler(t *testing.T) {
	log, _ := logger.New("debug")
	dispatcher := NewDispatcher(log)

	// Create a test handler
	handler := &testHandler{
		name:  "test_handler",
		types: []EventType{EventTypeLog},
	}

	// Register handler
	err := dispatcher.RegisterHandler(handler)
	if err != nil {
		t.Fatalf("Expected no error on handler registration, got: %v", err)
	}

	// Check if handler is registered
	handlers := dispatcher.GetRegisteredHandlers()
	if len(handlers[EventTypeLog]) != 1 {
		t.Error("Expected handler to be registered")
	}

	if handlers[EventTypeLog][0] != "test_handler" {
		t.Error("Expected correct handler name")
	}
}

func TestDispatcher_RegisterNilHandler(t *testing.T) {
	log, _ := logger.New("debug")
	dispatcher := NewDispatcher(log)

	err := dispatcher.RegisterHandler(nil)
	if err == nil {
		t.Error("Expected error when registering nil handler")
	}
}

func TestDispatcher_UnregisterHandler(t *testing.T) {
	log, _ := logger.New("debug")
	dispatcher := NewDispatcher(log)

	// Create and register handler
	handler := &testHandler{
		name:  "test_handler",
		types: []EventType{EventTypeLog},
	}
	dispatcher.RegisterHandler(handler)

	// Unregister handler
	dispatcher.UnregisterHandler("test_handler")

	// Check if handler is unregistered
	handlers := dispatcher.GetRegisteredHandlers()
	if len(handlers[EventTypeLog]) != 0 {
		t.Error("Expected handler to be unregistered")
	}
}

func TestDispatcher_Dispatch(t *testing.T) {
	log, _ := logger.New("debug")
	dispatcher := NewDispatcher(log)
	ctx := context.Background()

	// Start dispatcher
	dispatcher.Start(ctx)
	defer dispatcher.Stop()

	// Create a test handler
	handler := &testHandler{
		name:  "test_handler",
		types: []EventType{EventTypeLog},
	}
	dispatcher.RegisterHandler(handler)

	// Create and dispatch event
	event := Event{
		Type:      EventTypeLog,
		Data:      map[string]interface{}{"message": "test"},
		Timestamp: time.Now(),
		Source:    "test",
		ID:        "test-1",
	}

	err := dispatcher.Dispatch(event)
	if err != nil {
		t.Fatalf("Expected no error on dispatch, got: %v", err)
	}

	// Wait a bit for processing
	time.Sleep(100 * time.Millisecond)

	// Check if handler was called
	if !handler.called {
		t.Error("Expected handler to be called")
	}
}

func TestDispatcher_GetStats(t *testing.T) {
	log, _ := logger.New("debug")
	dispatcher := NewDispatcher(log)
	ctx := context.Background()

	dispatcher.Start(ctx)
	defer dispatcher.Stop()

	// Dispatch some events
	event := Event{
		Type:      EventTypeLog,
		Data:      map[string]interface{}{"message": "test"},
		Timestamp: time.Now(),
		Source:    "test",
		ID:        "test-1",
	}

	dispatcher.Dispatch(event)
	dispatcher.Dispatch(event)

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Check stats
	stats := dispatcher.GetStats()
	if stats.EventsProcessed != 2 {
		t.Errorf("Expected 2 events processed, got %d", stats.EventsProcessed)
	}
}

func TestConvertSboxctlEvent(t *testing.T) {
	// Create a test sboxctl event
	sboxEvent := services.SboxctlEvent{
		Type:      "log",
		Data:      map[string]interface{}{"message": "test"},
		Timestamp: "2023-01-01T12:00:00Z",
		Version:   "1.0",
	}

	// Convert to generic event
	event := ConvertSboxctlEvent(sboxEvent)

	if event.Type != EventTypeLog {
		t.Errorf("Expected event type to be 'log', got %s", event.Type)
	}

	if event.Source != "sboxctl" {
		t.Errorf("Expected source to be 'sboxctl', got %s", event.Source)
	}

	if event.ID == "" {
		t.Error("Expected event ID to be generated")
	}
}

// testHandler is a test implementation of EventHandler
type testHandler struct {
	name   string
	types  []EventType
	called bool
}

func (h *testHandler) Handle(ctx context.Context, event Event) error {
	h.called = true
	return nil
}

func (h *testHandler) GetName() string {
	return h.name
}

func (h *testHandler) GetSupportedTypes() []EventType {
	return h.types
}
