package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	_ "github.com/warpstreamlabs/bento/public/components/io"
	_ "github.com/warpstreamlabs/bento/public/components/prometheus"
	_ "github.com/warpstreamlabs/bento/public/components/pure"
	"github.com/warpstreamlabs/bento/public/service"
)

func runStream(ctx context.Context, configFile, streamName string, wg *sync.WaitGroup) {
	defer wg.Done()

	// Read YAML configuration from file
	configYAML, err := os.ReadFile(configFile)
	if err != nil {
		log.Printf("Failed to read %s config file: %v", streamName, err)
		return
	}

	builder := service.NewStreamBuilder()

	// Parse the configuration
	err = builder.SetYAML(string(configYAML))
	if err != nil {
		log.Printf("Failed to create %s: %v", streamName, err)
		return
	}

	stream, err := builder.Build()
	if err != nil {
		log.Printf("Failed to create %s: %v", streamName, err)
		return
	}

	log.Printf("Starting %s...", streamName)

	// Run the stream
	if err := stream.Run(ctx); err != nil {
		log.Printf("%s finished with error: %v", streamName, err)
	} else {
		log.Printf("%s finished cleanly", streamName)
	}
}

func monitorCache(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if data, err := os.ReadFile("cache"); err == nil {
				fmt.Printf("=== Cache content at %s ===\n", time.Now().Format("15:04:05"))
				fmt.Printf("%s\n\n", string(data))
			} else {
				fmt.Println("Cache file not yet created...")
			}
		}
	}
}

func main() {
	fmt.Println("Starting Bento cache test...")
	fmt.Println("Press Ctrl+C to stop both streams")
	fmt.Println()

	// Clean up any existing cache files
	_ = os.Remove("cache")

	// Create context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	var wg sync.WaitGroup

	// Start monitoring cache in a separate goroutine
	go monitorCache(ctx)

	// Start both streams
	wg.Add(2)
	go runStream(ctx, "stream_a.yaml", "Stream A", &wg)
	go runStream(ctx, "stream_b.yaml", "Stream B", &wg)

	// Wait for interrupt signal
	<-sigChan
	fmt.Println("\nReceived interrupt signal, stopping streams...")

	// Cancel context to stop all operations
	cancel()

	// Wait for all goroutines to finish
	wg.Wait()

	fmt.Println("Test completed.")
}
