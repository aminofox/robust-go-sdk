package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/aminofox/robust-go-sdk/consumer"
)

// ConsumerHandler implements the ConsumerGroupHandler interface
type ConsumerHandler struct {
	messageCount int
	mu           sync.Mutex
}

// Setup is called at the beginning of a new session
func (h *ConsumerHandler) Setup(session consumer.ConsumerGroupSession) error {
	fmt.Println("🔧 Consumer group session started")
	fmt.Printf("   Member ID: %s\n", session.MemberID())
	fmt.Printf("   Generation ID: %d\n", session.GenerationID())
	fmt.Printf("   Claims: %v\n", session.Claims())
	fmt.Println()
	return nil
}

// Cleanup is called at the end of a session
func (h *ConsumerHandler) Cleanup(session consumer.ConsumerGroupSession) error {
	fmt.Println("\n🧹 Consumer group session cleanup")
	fmt.Printf("   Total messages processed: %d\n", h.messageCount)
	return nil
}

// ConsumeClaim processes messages from a single partition
func (h *ConsumerHandler) ConsumeClaim(session consumer.ConsumerGroupSession, claim consumer.ConsumerGroupClaim) error {
	fmt.Printf("📥 Started consuming topic=%s partition=%d offset=%d\n",
		claim.Topic(), claim.Partition(), claim.InitialOffset())

	for {
		select {
		case msg, ok := <-claim.Messages():
			if !ok {
				return nil
			}

			// Process the message
			h.mu.Lock()
			h.messageCount++
			count := h.messageCount
			h.mu.Unlock()

			fmt.Printf("✅ Message #%d: topic=%s partition=%d offset=%d value=%s\n",
				count, msg.Topic, msg.Partition, msg.Offset, string(msg.Value))

			// Mark message as consumed
			session.MarkMessage(msg, "")

			// Commit offsets every 10 messages
			if count%10 == 0 {
				if err := session.Commit(); err != nil {
					log.Printf("❌ Failed to commit offset: %v\n", err)
				} else {
					fmt.Printf("💾 Committed offset: topic=%s partition=%d offset=%d\n",
						msg.Topic, msg.Partition, msg.Offset)
				}
			}

		case <-session.Context().Done():
			return nil
		}
	}
}

func main() {
	fmt.Println("🚀 Consumer Group Demo")
	fmt.Println("======================")
	fmt.Println()

	// Create consumer group configuration
	config := consumer.DefaultConfig()
	config.Brokers = []string{"localhost:8686"}

	// Create consumer group
	groupID := "test-consumer-group"
	topics := []string{"demo-topic"}

	cg, err := consumer.NewConsumerGroup(groupID, topics, config)
	if err != nil {
		log.Fatal("Failed to create consumer group:", err)
	}
	defer cg.Close()

	fmt.Printf("📋 Consumer Group: %s\n", groupID)
	fmt.Printf("📌 Topics: %v\n", topics)
	fmt.Printf("🔌 Brokers: %v\n", config.Brokers)
	fmt.Println()

	// Setup signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Create handler
	handler := &ConsumerHandler{}

	// Start consuming in a goroutine
	consumeErr := make(chan error, 1)
	go func() {
		consumeErr <- cg.Consume(ctx, topics, handler)
	}()

	// Handle errors
	go func() {
		for err := range cg.Errors() {
			log.Printf("⚠️  Consumer error: %v\n", err)
		}
	}()

	fmt.Println("👂 Listening for messages... (Press Ctrl+C to stop)")
	fmt.Println()

	// Wait for signal or error
	select {
	case <-sigChan:
		fmt.Println("\n\n🛑 Received interrupt signal, shutting down...")
		cancel()
	case err := <-consumeErr:
		if err != nil {
			log.Printf("❌ Consume error: %v\n", err)
		}
		cancel()
	}

	fmt.Println("✅ Consumer group shut down gracefully")
}
