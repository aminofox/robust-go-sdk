# Quick Start Guide

This guide will help you get started with Robust Go SDK by running the examples with a real Robust server.

## Prerequisites

- Docker and Docker Compose installed
- Go 1.23 or higher
- Port 8686 available

## Step 1: Start Robust Server

From the `examples` directory, start the Robust server using Docker:

```bash
cd examples
docker-compose up -d
```

Check that the server is running and healthy:

```bash
docker-compose ps
```

You should see:
```
NAME            IMAGE                   STATUS
robust-server   aminofox/robust:0.1.0   Up (healthy)   0.0.0.0:8686->8686/tcp
```

The server is now running at `localhost:8686`.

## Step 2: Run Admin Example (Create Topic)

First, create a topic using the admin client:

```bash
cd admin
go run main.go
```

Expected output:
```
🔧 Robust Admin Client Demo
============================

1️⃣  Listing existing topics...
   Found 0 topics: []

2️⃣  Creating topic 'demo-topic' with 3 partitions...
   ✅ Topic created successfully!

3️⃣  Listing topics after creation...
   Found 1 topics: [demo-topic]

✅ Admin demo complete!
```

## Step 3: Run Producer Example

### Sync Producer (Simple)

```bash
cd ../producer/simple
go run main.go
```

Expected output:
```
✅ Message sent successfully!
   Topic: demo-topic
   Partition: 1
   Offset: 0
   Value: Hello Robust!

✅ Message with key sent successfully!
   Partition: 0, Offset: 0
```

### Async Producer (High Throughput)

```bash
cd ../async
go run main.go
```

Expected output:
```
🚀 Async Producer Demo
======================
Batch Size: 10
Batch Timeout: 100ms

📤 Sending 50 messages...
   Queued 10 messages...
   Queued 20 messages...
   Queued 30 messages...
   Queued 40 messages...
   Queued 50 messages...

⏳ Waiting for all messages to be sent...
✅ Success #1: partition=2, offset=0
✅ Success #2: partition=0, offset=1
... (50 messages total)

📊 Results:
   Successes: 50
   Errors: 0

✅ Async producer demo complete!
```

## Step 4: Run Consumer Example

### Simple Consumer (Single Partition)

Open a new terminal and run:

```bash
cd consumer/simple
go run main.go
```

Expected output:
```
🔄 Consumer started, waiting for messages...
   Topic: demo-topic, Partition: 0
   Press Ctrl+C to stop

📨 Message #1 received
   Offset: 0
   Key: key1
   Value: Hello with key!
   Timestamp: 2026-02-28T09:22:41+07:00
```

Keep this running and send more messages from another terminal to see them consumed in real-time!

### Consumer Group (Multiple Consumers)

```bash
cd ../group
go run main.go
```

Expected output:
```
👥 Consumer Group Demo
======================
📋 Consumer Group: test-consumer-group
📌 Topics: [demo-topic]
🔌 Brokers: [localhost:8686]

🔧 Consumer group session started
   Member ID: consumer-xxx
   Generation ID: 1
   Claims: map[demo-topic:[0 1 2]]

📥 Started consuming topic=demo-topic partition=0 offset=0
📥 Started consuming topic=demo-topic partition=1 offset=0
📥 Started consuming topic=demo-topic partition=2 offset=0
```

## Step 5: Run Metrics Example

Monitor producer performance with metrics:

```bash
cd ../metrics
go run main.go
```

Expected output:
```
📊 Producer Metrics Demo
========================

🔌 Connected to: [localhost:8686]

📤 Sending 100 messages...
   Sent 20 messages...
   Sent 40 messages...
   Sent 60 messages...
   Sent 80 messages...
   Sent 100 messages...

✅ All messages sent!

📊 Producer Metrics:
   Messages Sent: 100
   Messages Failed: 0
   Bytes Sent: 1234
   Success Rate: 100.00%
   Average Latency: 2.5ms
   Throughput: 40000 messages/second
```

## Complete Test Flow

Run this sequence to test the full SDK end-to-end:

### Terminal 1: Start Server and Admin
```bash
# Start server
cd examples
docker-compose up -d

# Create topic
cd admin
go run main.go
```

### Terminal 2: Start Consumer
```bash
cd consumer/simple
go run main.go
```

### Terminal 3: Send Messages
```bash
cd producer/simple
go run main.go
```

You should see messages appear in Terminal 2 as they're sent from Terminal 3!

## Stop the Server

When you're done:

```bash
cd examples
docker-compose down
```

To remove the data volume as well:

```bash
docker-compose down -v
```

## Troubleshooting

### Connection Refused

If you see "connection refused" errors:
1. Check that Docker is running: `docker ps`
2. Check that the server is healthy: `docker-compose ps`
3. Wait a few seconds for the server to fully start
4. Check logs: `docker-compose logs robust`

### Topic Not Found

If you see "topic not found" errors:
1. Run the admin example first to create the topic
2. Check that the topic exists: `cd admin && go run main.go`

### Port Already in Use

If port 8686 is already in use:
1. Stop the conflicting service or change the port in `docker-compose.yml`
2. Update the broker address in example code to match

## Next Steps

- Read the [examples README](README.md) for detailed API documentation
- See the main [README](../README.md) for complete SDK documentation and API reference
- Explore advanced examples in the examples directory

## Need Help?

- Check server logs: `docker-compose logs -f robust`
- Verify SDK version: `go list -m github.com/aminofox/robust-go-sdk`
- Review example source code in each subdirectory
- See main [README.md](../README.md) for complete implementation guide
