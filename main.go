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

func produceee(ctx context.Context, producer service.MessageHandlerFunc) {
	producer(ctx, service.NewMessage([]byte("Hello from stream A!")))
	producer(ctx, service.NewMessage([]byte("Another hello from stream A!")))
}

func createStreamB(ctx context.Context, streamName string) (*service.StreamBuilder, service.MessageHandlerFunc, error) {
	streamBBuilder := service.NewStreamBuilder()
	producerFunc, err := streamBBuilder.AddProducerFunc()
	if err != nil {
		log.Printf("Failed to add producer function to %s: %v", streamName, err)
		return nil, nil, err
	}
	err = streamBBuilder.AddProcessorYAML(`
mapping: |
  root = content().string().uppercase()
`)
	if err != nil {
		log.Printf("Failed to add procesor to %s: %v", streamName, err)
		return nil, nil, err
	}
	streamBBuilder.AddConsumerFunc(func(ctx context.Context, message *service.Message) error {
		bytes, err := message.AsBytes()
		if err != nil {
			return err
		}
		fmt.Println("consuming message", string(bytes))
		return nil
	})

	return streamBBuilder, producerFunc, nil
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

	streamBBuilder, streamBproducer, err := createStreamB(ctx, "streamB")
	if err != nil {
		log.Fatalf("Failed to build streamB: %v", err)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		interval, err := time.ParseDuration("1s")
		if err != nil {
			return
		}
		for {
			fmt.Println("building stream")
			stream, err := streamBBuilder.Build()
			if err != nil {
				log.Printf("Failed to build: %v", err)
				return
			}
			streamDone := make(chan error, 1)
			go func() {
				streamDone <- stream.Run(ctx)
			}()

			select {
			case <-ctx.Done():
				<-streamDone
				fmt.Println("streamB context done, exiting")
				return
			case <-time.After(interval):
				fmt.Println("stopping streamB gracefully...")
				err := stream.Stop(ctx)
				if err != nil {
					fmt.Println("failed to stop streamB gracefully:", err)
					return
				}
				// ... and run again
			}
		}
	}()

	log.Println("Producing messages to stream B...")

	time.Sleep(3 * time.Second)
	produceee(ctx, streamBproducer)

	// Wait for interrupt signal
	<-sigChan
	fmt.Println("\nReceived interrupt signal, stopping streams...")

	// Cancel context to stop all operations
	cancel()

	// Wait for all goroutines to finish
	wg.Wait()

	fmt.Println("Test completed.")
}
