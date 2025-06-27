package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/kpblcaoo/sboxagent/internal/agent"
	"github.com/kpblcaoo/sboxagent/internal/config"
)

var (
	configFile = flag.String("config", "", "Path to configuration file")
	debug      = flag.Bool("debug", false, "Enable debug mode")
	logLevel   = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
)

func main() {
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Override config with command line flags
	if *debug {
		cfg.Agent.LogLevel = "debug"
	}
	if *logLevel != "info" {
		cfg.Agent.LogLevel = *logLevel
	}

	// Create agent instance
	agent, err := agent.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		fmt.Printf("\nReceived signal %v, shutting down...\n", sig)
		cancel()
	}()

	// Start agent
	fmt.Printf("Starting Subbox Agent (sboxagent) v%s\n", cfg.Agent.Version)
	fmt.Printf("Agent name: %s\n", cfg.Agent.Name)
	fmt.Printf("Log level: %s\n", cfg.Agent.LogLevel)

	if err := agent.Start(ctx); err != nil {
		log.Fatalf("Agent failed: %v", err)
	}

	fmt.Println("Agent stopped gracefully")
} 