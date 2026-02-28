package main

import (
	"fmt"
	"log"
	"time"

	"github.com/aminofox/robust-go-sdk/consumer"
)

func main() {
	// Create consumer with default config
	config := consumer.DefaultConfig()
	config.Brokers = []string{"localhost:8686"}

	c, err := consumer.NewPartitionConsumer("demo-topic", 0, config)
	if err != nil {
		log.Fatal("Failed to create consumer:", err)
	}
	defer c.Close()

	fmt.Println("🔄 Consumer started, waiting for messages...")
	fmt.Println("   Topic: demo-topic, Partition: 0")
	fmt.Println("   Press Ctrl+C to stop")
	fmt.Println()

	// Consume messages
	msgCount := 0
	timeout := time.After(30 * time.Second)

	for {
		select {
		case msg, ok := <-c.Messages():
			if !ok {
				fmt.Println("\n❌ Consumer closed")
				return
			}

			msgCount++
			fmt.Printf("📨 Message #%d received\n", msgCount)
			fmt.Printf("   Offset: %d\n", msg.Offset)
			fmt.Printf("   Key: %s\n", string(msg.Key))
			fmt.Printf("   Value: %s\n", string(msg.Value))
			fmt.Printf("   Timestamp: %s\n", time.UnixMilli(msg.Timestamp).Format(time.RFC3339))
			fmt.Println()

		case err, ok := <-c.Errors():
			if !ok {
				return
			}
			if err != nil {
				log.Printf("⚠️  Consumer error: %v\n", err)
			}

		case <-timeout:
			fmt.Printf("\n✅ Consumer test complete! Consumed %d messages\n", msgCount)
			return
		}
	}
}
