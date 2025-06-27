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

// Version information - will be set during build
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

var (
	configFile = flag.String("config", "", "Path to configuration file")
	debug      = flag.Bool("debug", false, "Enable debug mode")
	logLevel   = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	version    = flag.Bool("version", false, "Show version information")
)

func main() {
	flag.Parse()

	// Show version information if requested
	if *version {
		fmt.Printf("SboxAgent v%s\n", Version)
		fmt.Printf("Build time: %s\n", BuildTime)
		fmt.Printf("Git commit: %s\n", GitCommit)
		os.Exit(0)
	}

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

	// Update agent version with build information
	cfg.Agent.Version = Version

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
	fmt.Printf("Starting SboxAgent v%s\n", Version)
	fmt.Printf("Build time: %s\n", BuildTime)
	fmt.Printf("Git commit: %s\n", GitCommit)
	fmt.Printf("Agent name: %s\n", cfg.Agent.Name)
	fmt.Printf("Log level: %s\n", cfg.Agent.LogLevel)

	if err := agent.Start(ctx); err != nil {
		log.Fatalf("Agent failed: %v", err)
	}

	fmt.Println("Agent stopped gracefully")
}
