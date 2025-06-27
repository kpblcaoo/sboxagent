package dispatcher

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kpblcaoo/sboxagent/internal/logger"
	"github.com/kpblcaoo/sboxagent/internal/services"
)

// EventType represents the type of event
type EventType string

const (
	EventTypeLog    EventType = "log"
	EventTypeConfig EventType = "config"
	EventTypeError  EventType = "error"
	EventTypeStatus EventType = "status"
	EventTypeHealth EventType = "health"
)

// Event represents a generic event that can be dispatched
type Event struct {
	Type      EventType              `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	Source    string                 `json:"source"`
	ID        string                 `json:"id,omitempty"`
}

// EventHandler defines the interface for event handlers
type EventHandler interface {
	Handle(ctx context.Context, event Event) error
	GetName() string
	GetSupportedTypes() []EventType
}

// Dispatcher represents the event dispatcher
type Dispatcher struct {
	logger *logger.Logger

	// Handlers registry
	mu       sync.RWMutex
	handlers map[EventType][]EventHandler

	// Event processing
	eventChan chan Event
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup

	// Statistics
	statsMu sync.RWMutex
	stats   DispatcherStats
}

// DispatcherStats holds dispatcher statistics
type DispatcherStats struct {
	EventsProcessed int64
	EventsDropped   int64
	Errors          int64
	LastEventTime   time.Time
	StartTime       time.Time
}

// NewDispatcher creates a new event dispatcher
func NewDispatcher(log *logger.Logger) *Dispatcher {
	return &Dispatcher{
		logger:    log,
		handlers:  make(map[EventType][]EventHandler),
		eventChan: make(chan Event, 1000), // Large buffer for high throughput
		stats: DispatcherStats{
			StartTime: time.Now(),
		},
	}
}

// Start starts the dispatcher
func (d *Dispatcher) Start(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.ctx != nil {
		return fmt.Errorf("dispatcher is already running")
	}

	d.ctx, d.cancel = context.WithCancel(ctx)
	d.stats.StartTime = time.Now()

	d.logger.Info("Event dispatcher starting", map[string]interface{}{
		"bufferSize": cap(d.eventChan),
	})

	// Start event processing loop
	d.wg.Add(1)
	go d.processEvents()

	return nil
}

// Stop stops the dispatcher
func (d *Dispatcher) Stop() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.ctx == nil {
		return
	}

	d.logger.Info("Event dispatcher stopping", map[string]interface{}{})
	d.cancel()

	// Wait for processing to complete
	d.wg.Wait()

	d.ctx = nil
	d.cancel = nil
}

// RegisterHandler registers an event handler
func (d *Dispatcher) RegisterHandler(handler EventHandler) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	supportedTypes := handler.GetSupportedTypes()
	if len(supportedTypes) == 0 {
		return fmt.Errorf("handler must support at least one event type")
	}

	for _, eventType := range supportedTypes {
		d.handlers[eventType] = append(d.handlers[eventType], handler)
	}

	d.logger.Info("Event handler registered", map[string]interface{}{
		"handler":        handler.GetName(),
		"supportedTypes": supportedTypes,
	})

	return nil
}

// UnregisterHandler unregisters an event handler
func (d *Dispatcher) UnregisterHandler(handlerName string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	for eventType, handlers := range d.handlers {
		for i, handler := range handlers {
			if handler.GetName() == handlerName {
				// Remove handler from slice
				d.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
				d.logger.Info("Event handler unregistered", map[string]interface{}{
					"handler":   handlerName,
					"eventType": eventType,
				})
				break
			}
		}
	}
}

// Dispatch dispatches an event to registered handlers
func (d *Dispatcher) Dispatch(event Event) error {
	// Set timestamp if not set
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Update statistics
	d.statsMu.Lock()
	d.stats.EventsProcessed++
	d.stats.LastEventTime = event.Timestamp
	d.statsMu.Unlock()

	// Send to processing channel
	select {
	case d.eventChan <- event:
		return nil
	default:
		// Channel is full, drop event
		d.statsMu.Lock()
		d.stats.EventsDropped++
		d.statsMu.Unlock()

		d.logger.Warn("Event channel is full, dropping event", map[string]interface{}{
			"type": event.Type,
			"id":   event.ID,
		})
		return fmt.Errorf("event channel is full")
	}
}

// processEvents processes events from the channel
func (d *Dispatcher) processEvents() {
	defer d.wg.Done()

	for {
		select {
		case <-d.ctx.Done():
			d.logger.Info("Event processing loop stopped", map[string]interface{}{})
			return
		case event := <-d.eventChan:
			d.handleEvent(event)
		}
	}
}

// handleEvent handles a single event
func (d *Dispatcher) handleEvent(event Event) {
	d.mu.RLock()
	handlers := d.handlers[event.Type]
	d.mu.RUnlock()

	if len(handlers) == 0 {
		d.logger.Debug("No handlers registered for event type", map[string]interface{}{
			"type": event.Type,
		})
		return
	}

	d.logger.Debug("Processing event", map[string]interface{}{
		"type":     event.Type,
		"id":       event.ID,
		"source":   event.Source,
		"handlers": len(handlers),
	})

	// Process with all registered handlers
	var wg sync.WaitGroup
	for _, handler := range handlers {
		wg.Add(1)
		go func(h EventHandler) {
			defer wg.Done()
			if err := h.Handle(d.ctx, event); err != nil {
				d.statsMu.Lock()
				d.stats.Errors++
				d.statsMu.Unlock()

				d.logger.Error("Handler failed to process event", map[string]interface{}{
					"handler": h.GetName(),
					"type":    event.Type,
					"id":      event.ID,
					"error":   err.Error(),
				})
			}
		}(handler)
	}

	// Wait for all handlers to complete
	wg.Wait()
}

// GetStats returns dispatcher statistics
func (d *Dispatcher) GetStats() DispatcherStats {
	d.statsMu.RLock()
	defer d.statsMu.RUnlock()
	return d.stats
}

// GetEventsProcessed returns the number of events processed
func (d *DispatcherStats) GetEventsProcessed() int64 {
	return d.EventsProcessed
}

// GetEventsDropped returns the number of events dropped
func (d *DispatcherStats) GetEventsDropped() int64 {
	return d.EventsDropped
}

// GetErrors returns the number of errors
func (d *DispatcherStats) GetErrors() int64 {
	return d.Errors
}

// GetLastEventTime returns the last event time
func (d *DispatcherStats) GetLastEventTime() time.Time {
	return d.LastEventTime
}

// GetRegisteredHandlers returns information about registered handlers
func (d *Dispatcher) GetRegisteredHandlers() map[EventType][]string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	result := make(map[EventType][]string)
	for eventType, handlers := range d.handlers {
		handlerNames := make([]string, len(handlers))
		for i, handler := range handlers {
			handlerNames[i] = handler.GetName()
		}
		result[eventType] = handlerNames
	}
	return result
}

// ConvertSboxctlEvent converts a SboxctlEvent to a generic Event
func ConvertSboxctlEvent(sboxEvent services.SboxctlEvent) Event {
	event := Event{
		Type:      EventType(sboxEvent.Type),
		Data:      sboxEvent.Data,
		Source:    "sboxctl",
		Timestamp: time.Now(), // Will be overridden if timestamp is provided
	}

	// Try to parse timestamp if provided
	if sboxEvent.Timestamp != "" {
		if t, err := time.Parse(time.RFC3339, sboxEvent.Timestamp); err == nil {
			event.Timestamp = t
		}
	}

	// Generate ID if not provided
	if event.ID == "" {
		event.ID = fmt.Sprintf("%s-%d", event.Type, time.Now().UnixNano())
	}

	return event
}
