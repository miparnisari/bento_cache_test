package main

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"
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

var trafficLightIDs = []string{"one", "two"}

func produceSomeEvents(producer service.MessageHandlerFunc) {
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		// producing blocks until the message can be accepted, so do it in a goroutine
		go func() {
			defer wg.Done()
			tid := trafficLightIDs[rand.Int64N(int64(len(trafficLightIDs)))]
			msg := fmt.Sprintf("{ \"traffic_light\": \"%s\",  \"created_at\": \"%s\", \"passengers\": %d}", tid, time.Now().Add(-24*time.Hour).Format(time.RFC3339), rand.Int64N(10))
			fmt.Println("sending message", msg)
			err := producer(context.Background(), service.NewMessage([]byte(msg)))
			if err != nil {
				log.Printf("Failed to produce message to stream B: %v", err)
				return
			}
		}()
	}

	wg.Wait()
}

func createStreamB(ctx context.Context, streamName string) (*service.StreamBuilder, service.MessageHandlerFunc, error) {
	streamBBuilder := service.NewStreamBuilder()
	producerFunc, err := streamBBuilder.AddProducerFunc()
	if err != nil {
		log.Printf("Failed to add producer function to %s: %v", streamName, err)
		return nil, nil, err
	}
	err = streamBBuilder.SetBufferYAML(`
system_window:
  timestamp_mapping: root = this.created_at
  size: 10s
`)
	if err != nil {
		log.Printf("Failed to set buffer for %s: %v", streamName, err)
		return nil, nil, err
	}
	err = streamBBuilder.AddProcessorYAML(`
  # Group messages of the window into batches of common traffic light IDs
  group_by_value:
    value: '${! json("traffic_light") }'`)
	if err != nil {
		log.Printf("Failed to add processor 1 to %s: %v", streamName, err)
		return nil, nil, err
	}
	err = streamBBuilder.AddProcessorYAML(`
  # Reduce each batch to a single message by deleting indexes > 0, and
  # aggregate the car and passenger counts.
mapping: |
    root = if batch_index() == 0 {
      {
        "traffic_light": this.traffic_light,
        "created_at": metadata("window_end_timestamp"),
        "passengers": json("passengers").from_all().sum(),
      }
    } else { deleted() }
`)
	if err != nil {
		log.Printf("Failed to add procesor 2 to %s: %v", streamName, err)
		return nil, nil, err
	}
	err = streamBBuilder.AddConsumerFunc(func(ctx context.Context, message *service.Message) error {
		bytes, err := message.AsBytes()
		if err != nil {
			return err
		}
		fmt.Println("-> consuming message", string(bytes))
		return nil
	})
	if err != nil {
		log.Printf("Failed to add consumer function to %s: %v", streamName, err)
		return nil, nil, err
	}

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

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	var wg sync.WaitGroup

	streamBBuilder, streamBproducer, err := createStreamB(ctx, "streamB")
	if err != nil {
		log.Fatalf("Failed to create streamB: %v", err)
	}

	streamB, err := streamBBuilder.Build()
	if err != nil {
		log.Fatalf("Failed to build streamB: %v", err)
	}

	// receiver of messages
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := streamB.Run(ctx)
		if err != nil {
			log.Printf("error running: %v", err)
		}
		streamB.Stop(context.Background())
	}()

	// sender of messages
Loop:
	for {
		select {
		case <-sigChan:
			fmt.Println("\nReceived interrupt signal, stopping streams...")
			break Loop
		case <-time.After(1 * time.Second):
			produceSomeEvents(streamBproducer)
			goto Loop
		}
	}

	// Cancel context to stop all operations
	cancel()

	// Wait for all goroutines to finish
	wg.Wait()

	fmt.Println("Test completed.")
}
