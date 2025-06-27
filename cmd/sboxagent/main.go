package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/kpblcaoo/sboxagent/internal/socket"
)

func main() {
	// Parse command line flags
	socketPath := flag.String("socket", "/tmp/sboxagent.sock", "Unix socket path")
	flag.Parse()

	// Create logger
	logger := log.New(os.Stdout, "[sboxagent] ", log.LstdFlags)

	// Create server
	server := socket.NewServer(*socketPath, logger)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		logger.Printf("Received signal: %v", sig)
		cancel()
	}()

	// Start server
	logger.Printf("Starting sboxagent server on socket: %s", *socketPath)
	if err := server.Start(ctx); err != nil {
		logger.Fatalf("Server error: %v", err)
	}

	logger.Println("Server stopped")
}
