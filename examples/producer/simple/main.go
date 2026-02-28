package main

import (
	"fmt"
	"log"

	"github.com/aminofox/robust-go-sdk/pkg/message"
	"github.com/aminofox/robust-go-sdk/producer"
)

func main() {
	// Create producer with default config
	config := producer.DefaultConfig()
	config.Brokers = []string{"localhost:8686"}

	p, err := producer.NewSyncProducer(config)
	if err != nil {
		log.Fatal("Failed to create producer:", err)
	}
	defer p.Close()

	// Send a simple message
	msg := message.NewMessage("demo-topic", []byte("Hello Robust!"))

	partition, offset, err := p.SendMessage(msg)
	if err != nil {
		log.Fatal("Failed to send message:", err)
	}

	fmt.Printf("✅ Message sent successfully!\n")
	fmt.Printf("   Topic: %s\n", msg.Topic)
	fmt.Printf("   Partition: %d\n", partition)
	fmt.Printf("   Offset: %d\n", offset)
	fmt.Printf("   Value: %s\n", string(msg.Value))

	// Send message with key
	msg2 := message.NewMessageWithKey("demo-topic", []byte("key1"), []byte("Hello with key!"))
	partition2, offset2, err := p.SendMessage(msg2)
	if err != nil {
		log.Fatal("Failed to send message:", err)
	}

	fmt.Printf("\n✅ Message with key sent successfully!\n")
	fmt.Printf("   Partition: %d, Offset: %d\n", partition2, offset2)
}
