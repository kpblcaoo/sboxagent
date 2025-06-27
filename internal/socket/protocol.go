// internal/socket/protocol.go
// sboxagent: реализация framed JSON protocol_v1
//
// TODO: Реализовать encode/decode framed JSON, валидацию, интеграцию с event handler.

// Package socket provides Unix socket communication with framed JSON protocol.
package socket

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
)

const (
	// FrameHeaderSize is the size of the frame header in bytes.
	// 4 bytes for message length + 4 bytes for protocol version.
	FrameHeaderSize = 8

	// ProtocolVersion is the current protocol version.
	ProtocolVersion = 1

	// MaxMessageSize is the maximum allowed message size in bytes.
	MaxMessageSize = 1024 * 1024 // 1MB
)

// MessageType represents the type of message.
type MessageType string

const (
	MessageTypeEvent     MessageType = "event"
	MessageTypeCommand   MessageType = "command"
	MessageTypeResponse  MessageType = "response"
	MessageTypeHeartbeat MessageType = "heartbeat"
)

// Message represents a framed JSON message according to protocol_v1.schema.json.
type Message struct {
	ID            string                 `json:"id"`
	Type          string                 `json:"type"`
	Timestamp     string                 `json:"timestamp"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	Event         *EventMessage          `json:"event,omitempty"`
	Command       *CommandMessage        `json:"command,omitempty"`
	Response      *ResponseMessage       `json:"response,omitempty"`
	Heartbeat     *HeartbeatMessage      `json:"heartbeat,omitempty"`
}

// EventMessage represents an event message.
type EventMessage struct {
	Event map[string]interface{} `json:"event"`
}

// CommandMessage represents a command message.
type CommandMessage struct {
	Command string                 `json:"command"`
	Params  map[string]interface{} `json:"params"`
}

// ResponseMessage represents a response message.
type ResponseMessage struct {
	Status    string                 `json:"status"`
	RequestID string                 `json:"request_id"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Error     *ErrorMessage          `json:"error,omitempty"`
}

// ErrorMessage represents an error in a response.
type ErrorMessage struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// HeartbeatMessage represents a heartbeat message.
type HeartbeatMessage struct {
	AgentID       string  `json:"agent_id"`
	Status        string  `json:"status"`
	UptimeSeconds float64 `json:"uptime_seconds,omitempty"`
	Version       string  `json:"version,omitempty"`
}

// NewMessage creates a new message with the given type.
func NewMessage(msgType MessageType) *Message {
	return &Message{
		ID:        uuid.New().String(),
		Type:      string(msgType),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

// NewEventMessage creates a new event message.
func NewEventMessage(event map[string]interface{}) *Message {
	msg := NewMessage(MessageTypeEvent)
	msg.Event = &EventMessage{Event: event}
	return msg
}

// NewCommandMessage creates a new command message.
func NewCommandMessage(command string, params map[string]interface{}) *Message {
	msg := NewMessage(MessageTypeCommand)
	msg.Command = &CommandMessage{
		Command: command,
		Params:  params,
	}
	return msg
}

// NewResponseMessage creates a new response message.
func NewResponseMessage(requestID, status string, data map[string]interface{}, err *ErrorMessage) *Message {
	msg := NewMessage(MessageTypeResponse)
	msg.Response = &ResponseMessage{
		Status:    status,
		RequestID: requestID,
		Data:      data,
		Error:     err,
	}
	return msg
}

// NewHeartbeatMessage creates a new heartbeat message.
func NewHeartbeatMessage(agentID, status string, uptimeSeconds float64, version string) *Message {
	msg := NewMessage(MessageTypeHeartbeat)
	msg.Heartbeat = &HeartbeatMessage{
		AgentID:       agentID,
		Status:        status,
		UptimeSeconds: uptimeSeconds,
		Version:       version,
	}
	return msg
}

// EncodeMessage encodes a message to framed JSON bytes.
func EncodeMessage(msg *Message) ([]byte, error) {
	if msg == nil {
		return nil, fmt.Errorf("message cannot be nil")
	}

	// Marshal message to JSON
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	// Check message size
	if len(data) > MaxMessageSize {
		return nil, fmt.Errorf("message too large: %d bytes (max: %d)", len(data), MaxMessageSize)
	}

	// Create frame header: 4 bytes length + 4 bytes version
	header := make([]byte, FrameHeaderSize)
	binary.BigEndian.PutUint32(header[0:4], uint32(len(data)))
	binary.BigEndian.PutUint32(header[4:8], ProtocolVersion)

	// Combine header and data
	return append(header, data...), nil
}

// DecodeMessage reads and decodes a framed JSON message from io.Reader.
func DecodeMessage(r io.Reader) (*Message, error) {
	// Read frame header
	header := make([]byte, FrameHeaderSize)
	if _, err := io.ReadFull(r, header); err != nil {
		if err == io.EOF {
			return nil, io.EOF
		}
		return nil, fmt.Errorf("failed to read frame header: %w", err)
	}

	// Extract length and version
	length := binary.BigEndian.Uint32(header[0:4])
	version := binary.BigEndian.Uint32(header[4:8])

	// Validate protocol version
	if version != ProtocolVersion {
		return nil, fmt.Errorf("unsupported protocol version: %d (expected: %d)", version, ProtocolVersion)
	}

	// Validate message size
	if length > MaxMessageSize {
		return nil, fmt.Errorf("message too large: %d bytes (max: %d)", length, MaxMessageSize)
	}

	// Read message data
	data := make([]byte, length)
	if _, err := io.ReadFull(r, data); err != nil {
		return nil, fmt.Errorf("failed to read message data: %w", err)
	}

	// Parse JSON
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	return &msg, nil
}

// WriteMessage writes a complete message to io.Writer.
func WriteMessage(w io.Writer, msg *Message) error {
	encoded, err := EncodeMessage(msg)
	if err != nil {
		return err
	}

	_, err = w.Write(encoded)
	return err
}

// ReadMessage reads a complete message from io.Reader.
func ReadMessage(r io.Reader) (*Message, error) {
	return DecodeMessage(r)
}
