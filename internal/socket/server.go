// internal/socket/server.go
// sboxagent: Unix socket server (framed JSON protocol_v1)
//
// TODO: Реализовать запуск Unix socket сервера, чтение/запись framed JSON сообщений,
// обработку команд и событий, интеграцию с event handler.

// Package socket implements a Unix socket server for framed JSON protocol.
package socket

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
)

// Server represents a Unix socket server for framed JSON protocol.
type Server struct {
	SocketPath string
	listener   net.Listener
	Logger     *log.Logger
}

// NewServer creates a new Server instance.
func NewServer(socketPath string, logger *log.Logger) *Server {
	return &Server{
		SocketPath: socketPath,
		Logger:     logger,
	}
}

// Start launches the Unix socket server and accepts connections.
// Each connection is handled in a separate goroutine.
func (s *Server) Start(ctx context.Context) error {
	if s.Logger == nil {
		s.Logger = log.New(os.Stdout, "[socket-server] ", log.LstdFlags)
	}

	// Remove old socket if exists
	if err := os.RemoveAll(s.SocketPath); err != nil {
		return fmt.Errorf("failed to remove old socket: %w", err)
	}

	ln, err := net.Listen("unix", s.SocketPath)
	if err != nil {
		return fmt.Errorf("failed to listen on unix socket: %w", err)
	}
	s.listener = ln
	s.Logger.Printf("Listening on unix socket: %s", s.SocketPath)

	go func() {
		<-ctx.Done()
		s.listener.Close()
		s.Logger.Println("Server stopped")
	}()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return nil // graceful shutdown
			default:
				s.Logger.Printf("Accept error: %v", err)
				return err
			}
		}
		go s.handleConnection(conn)
	}
}

// handleConnection processes a single client connection.
func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	s.Logger.Printf("Accepted connection from %v", conn.RemoteAddr())

	for {
		msg, err := ReadMessage(conn)
		if err != nil {
			if err.Error() != "EOF" {
				s.Logger.Printf("Read error: %v", err)
			}
			break
		}

		s.Logger.Printf("Received message: type=%s id=%s", msg.Type, msg.ID)

		// Echo back the same message (for test/demo)
		err = WriteMessage(conn, msg)
		if err != nil {
			s.Logger.Printf("Write error: %v", err)
			break
		}
	}

	s.Logger.Printf("Connection closed: %v", conn.RemoteAddr())
}

// Stop stops the server and closes the listener.
func (s *Server) Stop() error {
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}
