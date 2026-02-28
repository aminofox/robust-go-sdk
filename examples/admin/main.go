package main

import (
	"fmt"
	"log"

	"github.com/aminofox/robust-go-sdk/admin"
)

func main() {
	// Create admin client with default config
	config := admin.DefaultConfig()
	config.Brokers = []string{"localhost:8686"}

	adm, err := admin.NewAdminClient(config)
	if err != nil {
		log.Fatal("Failed to create admin client:", err)
	}
	defer adm.Close()

	fmt.Println("🔧 Robust Admin Client Demo")
	fmt.Println("============================")
	fmt.Println()

	// List existing topics
	fmt.Println("1️⃣  Listing existing topics...")
	topics, err := adm.ListTopics()
	if err != nil {
		log.Printf("Error listing topics: %v\n", err)
	} else {
		fmt.Printf("   Found %d topics: %v\n\n", len(topics), topics)
	}

	// Create a new topic
	topicName := "demo-topic"
	fmt.Printf("2️⃣  Creating topic '%s' with 3 partitions...\n", topicName)
	err = adm.CreateTopic(topicName, 3, map[string]string{
		"retention.hours": "168",
	})
	if err != nil {
		log.Printf("Error creating topic: %v\n", err)
	} else {
		fmt.Println("   ✅ Topic created successfully!")
		fmt.Println()
	}

	// List topics again
	fmt.Println("3️⃣  Listing topics after creation...")
	topics, err = adm.ListTopics()
	if err != nil {
		log.Printf("Error listing topics: %v\n", err)
	} else {
		fmt.Printf("   Found %d topics: %v\n\n", len(topics), topics)
	}

	// Get metadata
	fmt.Printf("4️⃣  Getting metadata for '%s'...\n", topicName)
	metadata, err := adm.GetMetadata([]string{topicName})
	if err != nil {
		log.Printf("Error getting metadata: %v\n", err)
	} else {
		for _, topic := range metadata {
			fmt.Printf("   Topic: %s\n", topic.Name)
			fmt.Printf("   Partitions: %d\n", len(topic.Partitions))
			for _, p := range topic.Partitions {
				fmt.Printf("      - Partition %d (Leader: %s)\n", p.ID, p.Leader)
			}
		}
		fmt.Println()
	}

	fmt.Println("✅ Admin demo complete!")
}
