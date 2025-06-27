package socket

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMessage(t *testing.T) {
	msg := NewMessage(MessageTypeEvent)

	assert.NotEmpty(t, msg.ID)
	assert.Equal(t, "event", msg.Type)
	assert.NotEmpty(t, msg.Timestamp)

	// Check timestamp format
	_, err := time.Parse(time.RFC3339, msg.Timestamp)
	assert.NoError(t, err)
}

func TestNewEventMessage(t *testing.T) {
	event := map[string]interface{}{
		"type": "config_updated",
		"data": map[string]interface{}{
			"config_id": "test-123",
		},
	}

	msg := NewEventMessage(event)

	assert.Equal(t, "event", msg.Type)
	assert.NotNil(t, msg.Event)
	assert.Equal(t, event, msg.Event.Event)
}

func TestNewCommandMessage(t *testing.T) {
	command := "get_status"
	params := map[string]interface{}{
		"include_details": true,
	}

	msg := NewCommandMessage(command, params)

	assert.Equal(t, "command", msg.Type)
	assert.NotNil(t, msg.Command)
	assert.Equal(t, command, msg.Command.Command)
	assert.Equal(t, params, msg.Command.Params)
}

func TestNewResponseMessage(t *testing.T) {
	requestID := "req-123"
	status := "success"
	data := map[string]interface{}{
		"status": "running",
		"uptime": 3600,
	}

	msg := NewResponseMessage(requestID, status, data, nil)

	assert.Equal(t, "response", msg.Type)
	assert.NotNil(t, msg.Response)
	assert.Equal(t, status, msg.Response.Status)
	assert.Equal(t, requestID, msg.Response.RequestID)
	assert.Equal(t, data, msg.Response.Data)
	assert.Nil(t, msg.Response.Error)
}

func TestNewResponseMessageWithError(t *testing.T) {
	requestID := "req-123"
	status := "error"
	err := &ErrorMessage{
		Code:    "INVALID_CONFIG",
		Message: "Configuration is invalid",
		Details: map[string]interface{}{
			"field": "proxy_url",
		},
	}

	msg := NewResponseMessage(requestID, status, nil, err)

	assert.Equal(t, "response", msg.Type)
	assert.NotNil(t, msg.Response)
	assert.Equal(t, status, msg.Response.Status)
	assert.Equal(t, requestID, msg.Response.RequestID)
	assert.Equal(t, err, msg.Response.Error)
}

func TestNewHeartbeatMessage(t *testing.T) {
	agentID := "agent-123"
	status := "healthy"
	uptime := 3600.5
	version := "1.0.0"

	msg := NewHeartbeatMessage(agentID, status, uptime, version)

	assert.Equal(t, "heartbeat", msg.Type)
	assert.NotNil(t, msg.Heartbeat)
	assert.Equal(t, agentID, msg.Heartbeat.AgentID)
	assert.Equal(t, status, msg.Heartbeat.Status)
	assert.Equal(t, uptime, msg.Heartbeat.UptimeSeconds)
	assert.Equal(t, version, msg.Heartbeat.Version)
}

func TestEncodeMessage(t *testing.T) {
	msg := NewEventMessage(map[string]interface{}{
		"type": "test_event",
		"data": "test_data",
	})

	encoded, err := EncodeMessage(msg)
	require.NoError(t, err)

	// Check frame header
	assert.Len(t, encoded, FrameHeaderSize+len(encoded)-FrameHeaderSize)

	// Extract length and version from header
	length := binary.BigEndian.Uint32(encoded[0:4])
	version := binary.BigEndian.Uint32(encoded[4:8])

	assert.Equal(t, uint32(ProtocolVersion), version)
	assert.Greater(t, length, uint32(0))
}

func TestEncodeMessageNil(t *testing.T) {
	_, err := EncodeMessage(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "message cannot be nil")
}

func TestDecodeMessage(t *testing.T) {
	original := NewEventMessage(map[string]interface{}{
		"type": "test_event",
		"data": "test_data",
	})

	encoded, err := EncodeMessage(original)
	require.NoError(t, err)

	// Create reader from encoded data
	reader := bytes.NewReader(encoded)

	decoded, err := DecodeMessage(reader)
	require.NoError(t, err)

	assert.Equal(t, original.ID, decoded.ID)
	assert.Equal(t, original.Type, decoded.Type)
	assert.Equal(t, original.Timestamp, decoded.Timestamp)
	assert.Equal(t, original.Event.Event, decoded.Event.Event)
}

func TestDecodeMessageEOF(t *testing.T) {
	reader := bytes.NewReader([]byte{})

	_, err := DecodeMessage(reader)
	assert.Equal(t, io.EOF, err)
}

func TestDecodeMessageInvalidHeader(t *testing.T) {
	// Create invalid header (too short)
	reader := bytes.NewReader([]byte{1, 2, 3})

	_, err := DecodeMessage(reader)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read frame header")
}

func TestDecodeMessageInvalidVersion(t *testing.T) {
	// Create header with wrong version
	header := make([]byte, FrameHeaderSize)
	binary.BigEndian.PutUint32(header[0:4], 10)  // length
	binary.BigEndian.PutUint32(header[4:8], 999) // wrong version

	reader := bytes.NewReader(header)

	_, err := DecodeMessage(reader)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported protocol version")
}

func TestDecodeMessageTooLarge(t *testing.T) {
	// Create header with too large message size
	header := make([]byte, FrameHeaderSize)
	binary.BigEndian.PutUint32(header[0:4], MaxMessageSize+1)
	binary.BigEndian.PutUint32(header[4:8], ProtocolVersion)

	reader := bytes.NewReader(header)

	_, err := DecodeMessage(reader)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "message too large")
}

func TestWriteAndReadMessage(t *testing.T) {
	original := NewCommandMessage("test_command", map[string]interface{}{
		"param1": "value1",
		"param2": 42,
	})

	var buf bytes.Buffer

	// Write message
	err := WriteMessage(&buf, original)
	require.NoError(t, err)

	// Read message back
	reader := bytes.NewReader(buf.Bytes())
	decoded, err := ReadMessage(reader)
	require.NoError(t, err)

	assert.Equal(t, original.ID, decoded.ID)
	assert.Equal(t, original.Type, decoded.Type)
	assert.Equal(t, original.Command.Command, decoded.Command.Command)

	// Compare JSON representation to handle type differences
	originalJSON, err := json.Marshal(original.Command.Params)
	require.NoError(t, err)

	decodedJSON, err := json.Marshal(decoded.Command.Params)
	require.NoError(t, err)

	assert.JSONEq(t, string(originalJSON), string(decodedJSON))
}

func TestMessageRoundTrip(t *testing.T) {
	testCases := []struct {
		name string
		msg  *Message
	}{
		{
			name: "event message",
			msg: NewEventMessage(map[string]interface{}{
				"type": "config_updated",
				"data": map[string]interface{}{
					"config_id": "test-123",
				},
			}),
		},
		{
			name: "command message",
			msg: NewCommandMessage("get_status", map[string]interface{}{
				"include_details": true,
			}),
		},
		{
			name: "response message",
			msg: NewResponseMessage("req-123", "success", map[string]interface{}{
				"status": "running",
			}, nil),
		},
		{
			name: "response with error",
			msg: NewResponseMessage("req-123", "error", nil, &ErrorMessage{
				Code:    "INVALID_CONFIG",
				Message: "Configuration is invalid",
			}),
		},
		{
			name: "heartbeat message",
			msg:  NewHeartbeatMessage("agent-123", "healthy", 3600.5, "1.0.0"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			encoded, err := EncodeMessage(tc.msg)
			require.NoError(t, err)

			reader := bytes.NewReader(encoded)
			decoded, err := DecodeMessage(reader)
			require.NoError(t, err)

			// Compare JSON representation for deep equality
			originalJSON, err := json.Marshal(tc.msg)
			require.NoError(t, err)

			decodedJSON, err := json.Marshal(decoded)
			require.NoError(t, err)

			assert.JSONEq(t, string(originalJSON), string(decodedJSON))
		})
	}
}

func TestMessageWithCorrelationID(t *testing.T) {
	msg := NewMessage(MessageTypeCommand)
	msg.CorrelationID = "corr-123"
	msg.Metadata = map[string]interface{}{
		"source":   "test",
		"priority": "high",
	}

	encoded, err := EncodeMessage(msg)
	require.NoError(t, err)

	reader := bytes.NewReader(encoded)
	decoded, err := DecodeMessage(reader)
	require.NoError(t, err)

	assert.Equal(t, "corr-123", decoded.CorrelationID)
	assert.Equal(t, "test", decoded.Metadata["source"])
	assert.Equal(t, "high", decoded.Metadata["priority"])
}
