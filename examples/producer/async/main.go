package main

import (
	"fmt"
	"log"
	"time"

	"github.com/aminofox/robust-go-sdk/pkg/message"
	"github.com/aminofox/robust-go-sdk/producer"
)

func main() {
	// Create async producer with batching
	config := producer.DefaultConfig()
	config.Brokers = []string{"localhost:8686"}
	config.BatchSize = 10
	config.BatchTimeout = 100 * time.Millisecond

	p, err := producer.NewAsyncProducer(config)
	if err != nil {
		log.Fatal("Failed to create async producer:", err)
	}
	defer p.Close()

	fmt.Println("🚀 Async Producer Demo")
	fmt.Println("======================")
	fmt.Printf("Batch Size: %d\n", config.BatchSize)
	fmt.Printf("Batch Timeout: %v\n", config.BatchTimeout)
	fmt.Println()

	// Handle successes
	successCount := 0
	go func() {
		for msg := range p.Successes() {
			successCount++
			fmt.Printf("✅ Success #%d: partition=%d, offset=%d\n",
				successCount, msg.Partition, msg.Offset)
		}
	}()

	// Handle errors
	errorCount := 0
	go func() {
		for err := range p.Errors() {
			errorCount++
			fmt.Printf("❌ Error #%d: %v\n", errorCount, err.Err)
		}
	}()

	// Send messages asynchronously
	fmt.Println("📤 Sending 50 messages...")
	for i := 0; i < 50; i++ {
		msg := message.NewMessage("demo-topic", []byte(fmt.Sprintf("Async message #%d", i)))

		select {
		case p.Input() <- msg:
			if (i+1)%10 == 0 {
				fmt.Printf("   Queued %d messages...\n", i+1)
			}
		case <-time.After(5 * time.Second):
			log.Printf("⚠️  Timeout sending message %d\n", i)
		}
	}

	fmt.Println("\n⏳ Waiting for all messages to be sent...")
	time.Sleep(2 * time.Second)

	fmt.Printf("\n📊 Results:\n")
	fmt.Printf("   Successes: %d\n", successCount)
	fmt.Printf("   Errors: %d\n", errorCount)
	fmt.Println("\n✅ Async producer demo complete!")
}
