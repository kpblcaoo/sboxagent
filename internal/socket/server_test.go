package socket

import (
	"context"
	"log"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestServer_Echo(t *testing.T) {
	dir := t.TempDir()
	socketPath := filepath.Join(dir, "test.sock")
	logger := log.New(os.Stdout, "[test-server] ", log.LstdFlags)
	server := NewServer(socketPath, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server in background
	go func() {
		err := server.Start(ctx)
		require.NoError(t, err)
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Connect as client
	conn, err := net.Dial("unix", socketPath)
	require.NoError(t, err)
	defer conn.Close()

	// Send message
	msg := NewEventMessage(map[string]interface{}{
		"type":    "test_echo",
		"payload": "hello",
	})
	err = WriteMessage(conn, msg)
	require.NoError(t, err)

	// Read echo response
	echo, err := ReadMessage(conn)
	require.NoError(t, err)
	require.Equal(t, msg.Type, echo.Type)
	require.Equal(t, msg.Event.Event, echo.Event.Event)

	// Stop server
	cancel()
	_ = server.Stop()
}
