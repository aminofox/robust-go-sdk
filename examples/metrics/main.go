package main

import (
	"fmt"
	"log"
	"time"

	"github.com/aminofox/robust-go-sdk/pkg/message"
	"github.com/aminofox/robust-go-sdk/pkg/metrics"
	"github.com/aminofox/robust-go-sdk/producer"
)

func main() {
	fmt.Println("📊 Producer Metrics Demo")
	fmt.Println("========================")
	fmt.Println()

	// Create producer configuration
	config := producer.DefaultConfig()
	config.Brokers = []string{"localhost:8686"}

	// Create producer
	p, err := producer.NewSyncProducer(config)
	if err != nil {
		log.Fatal("Failed to create producer:", err)
	}
	defer p.Close()

	// Create metrics collector
	producerMetrics := metrics.NewProducerMetrics()

	fmt.Printf("🔌 Connected to: %v\n", config.Brokers)
	fmt.Println()

	// Send messages and track metrics
	fmt.Println("📤 Sending 100 messages...")
	startTime := time.Now()

	for i := 0; i < 100; i++ {
		msg := message.NewMessage("demo-topic", []byte(fmt.Sprintf("Message #%d", i)))

		// Track send latency
		sendStart := time.Now()
		partition, offset, err := p.SendMessage(msg)
		sendLatency := time.Since(sendStart)

		if err != nil {
			producerMetrics.RecordFailed()
			fmt.Printf("❌ Failed to send message %d: %v\n", i+1, err)
			continue
		}

		// Record successful send
		producerMetrics.RecordSent(uint64(msg.Size()))

		// Print progress every 20 messages
		if (i+1)%20 == 0 {
			fmt.Printf("   Sent %d messages (last: partition=%d, offset=%d, latency=%v)\n",
				i+1, partition, offset, sendLatency)
		}
	}

	totalDuration := time.Since(startTime)

	// Get statistics
	stats := producerMetrics.GetStats()

	// Print detailed metrics
	fmt.Println()
	fmt.Println("📈 Producer Metrics:")
	fmt.Println("--------------------")
	fmt.Printf("Messages Sent:     %d\n", stats.MessagesSent)
	fmt.Printf("Messages Failed:   %d\n", stats.MessagesFailed)
	fmt.Printf("Bytes Sent:        %d bytes (%.2f KB)\n",
		stats.BytesSent, float64(stats.BytesSent)/1024)
	fmt.Printf("Total Duration:    %v\n", totalDuration)
	fmt.Println()

	// Calculate success rate
	total := stats.MessagesSent + stats.MessagesFailed
	var successRate float64
	if total > 0 {
		successRate = float64(stats.MessagesSent) / float64(total) * 100
	}

	// Calculate throughput
	seconds := totalDuration.Seconds()
	var messagesPerSec, bytesPerSec float64
	if seconds > 0 {
		messagesPerSec = float64(stats.MessagesSent) / seconds
		bytesPerSec = float64(stats.BytesSent) / seconds
	}

	fmt.Println("📊 Statistics:")
	fmt.Println("--------------")
	fmt.Printf("Success Rate:      %.2f%%\n", successRate)
	fmt.Printf("Throughput:        %.2f messages/sec\n", messagesPerSec)
	fmt.Printf("Throughput:        %.2f KB/sec\n", bytesPerSec/1024)
	fmt.Println()

	fmt.Println("✅ Metrics demo complete!")
}
