package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/aminofox/robust-go-sdk/internal/connection"
	"github.com/aminofox/robust-go-sdk/internal/protocol"
	sdkerrors "github.com/aminofox/robust-go-sdk/pkg/errors"
	"github.com/aminofox/robust-go-sdk/pkg/message"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ConsumerGroupHandler represents a consumer group handler
type ConsumerGroupHandler interface {
	// Setup is called at the beginning of a new session
	Setup(session ConsumerGroupSession) error

	// Cleanup is called at the end of a session
	Cleanup(session ConsumerGroupSession) error

	// ConsumeClaim processes messages from a single partition
	ConsumeClaim(session ConsumerGroupSession, claim ConsumerGroupClaim) error
}

// ConsumerGroupSession represents a consumer group session
type ConsumerGroupSession interface {
	// Claims returns the current partition claims
	Claims() map[string][]int32

	// MemberID returns the member ID
	MemberID() string

	// GenerationID returns the generation ID
	GenerationID() int32

	// MarkMessage marks a message as consumed
	MarkMessage(msg *message.Message, metadata string)

	// MarkOffset marks an offset as consumed
	MarkOffset(topic string, partition int32, offset int64, metadata string)

	// Commit commits the offsets
	Commit() error

	// Context returns the session context
	Context() context.Context
}

// ConsumerGroupClaim represents a partition claim
type ConsumerGroupClaim interface {
	// Topic returns the topic name
	Topic() string

	// Partition returns the partition ID
	Partition() int32

	// InitialOffset returns the initial offset
	InitialOffset() int64

	// HighWaterMarkOffset returns the high water mark offset
	HighWaterMarkOffset() int64

	// Messages returns the message channel
	Messages() <-chan *message.Message
}

// ConsumerGroup represents a consumer group
type ConsumerGroup interface {
	// Consume starts consuming from the specified topics
	Consume(ctx context.Context, topics []string, handler ConsumerGroupHandler) error

	// Errors returns the error channel
	Errors() <-chan error

	// Close closes the consumer group
	Close() error
}

// consumerGroup implements ConsumerGroup
type consumerGroup struct {
	groupID    string
	config     *Config
	pool       *connection.Pool
	handler    ConsumerGroupHandler
	memberID   string
	generation int32
	claims     map[string][]int32
	errors     chan error
	logger     *zap.Logger
	mu         sync.RWMutex
	closed     bool
	wg         sync.WaitGroup
}

// NewConsumerGroup creates a new consumer group
func NewConsumerGroup(groupID string, topics []string, config *Config) (ConsumerGroup, error) {
	if config == nil {
		config = DefaultConfig()
	}
	config.GroupID = groupID
	config.Topics = topics

	if err := config.Validate(); err != nil {
		return nil, err
	}

	pool, err := connection.NewPool(&connection.PoolConfig{
		Brokers:        config.Brokers,
		Size:           config.MaxConnections,
		ConnectTimeout: config.RequestTimeout,
	})
	if err != nil {
		return nil, err
	}

	logger, _ := zap.NewProduction()

	cg := &consumerGroup{
		groupID: groupID,
		config:  config,
		pool:    pool,
		errors:  make(chan error, 10),
		logger:  logger,
		claims:  make(map[string][]int32),
	}

	return cg, nil
}

// Consume starts consuming from the specified topics
func (cg *consumerGroup) Consume(ctx context.Context, topics []string, handler ConsumerGroupHandler) error {
	cg.mu.Lock()
	if cg.closed {
		cg.mu.Unlock()
		return sdkerrors.ErrClosed
	}
	cg.handler = handler
	cg.mu.Unlock()

	// Join the consumer group
	if err := cg.joinGroup(topics); err != nil {
		return fmt.Errorf("join group: %w", err)
	}

	// Sync group to get partition assignments
	if err := cg.syncGroup(); err != nil {
		return fmt.Errorf("sync group: %w", err)
	}

	// Create session
	session := &consumerGroupSession{
		cg:           cg,
		ctx:          ctx,
		claims:       cg.claims,
		memberID:     cg.memberID,
		generationID: cg.generation,
		offsets:      make(map[string]map[int32]int64),
	}

	// Setup handler
	if err := handler.Setup(session); err != nil {
		return fmt.Errorf("handler setup: %w", err)
	}

	// Start consuming
	errCh := make(chan error, len(cg.claims))
	for topic, partitions := range cg.claims {
		for _, partition := range partitions {
			cg.wg.Add(1)
			go func(t string, p int32) {
				defer cg.wg.Done()
				if err := cg.consumePartition(ctx, session, handler, t, p); err != nil {
					select {
					case errCh <- err:
					case <-ctx.Done():
					}
				}
			}(topic, partition)
		}
	}

	// Wait for context or error
	select {
	case <-ctx.Done():
		cg.wg.Wait()
	case err := <-errCh:
		cg.wg.Wait()
		return err
	}

	// Cleanup handler
	if err := handler.Cleanup(session); err != nil {
		return fmt.Errorf("handler cleanup: %w", err)
	}

	// Leave group
	if err := cg.leaveGroup(); err != nil {
		return fmt.Errorf("leave group: %w", err)
	}

	return nil
}

// joinGroup joins the consumer group
func (cg *consumerGroup) joinGroup(topics []string) error {
	conn, err := cg.pool.Get()
	if err != nil {
		return err
	}

	joinReq := &protocol.JoinGroupReq{
		GroupID:  cg.groupID,
		MemberID: cg.memberID,
		Topics:   topics,
	}

	requestID := uuid.New().String()
	payloadBytes, err := json.Marshal(joinReq)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req := &protocol.Request{
		Type:      protocol.JoinGroupRequest,
		RequestID: requestID,
		Payload:   payloadBytes,
	}

	resp, err := conn.Send(req)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("join group failed: %s", resp.Error)
	}

	var joinResp protocol.JoinGroupResp
	if err := json.Unmarshal(resp.Payload, &joinResp); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	if joinResp.Error != "" {
		return fmt.Errorf("join group error: %s", joinResp.Error)
	}

	cg.memberID = joinResp.MemberID
	cg.generation = int32(joinResp.Generation) // Cast from int64 to int32

	return nil
}

// syncGroup syncs the consumer group
func (cg *consumerGroup) syncGroup() error {
	conn, err := cg.pool.Get()
	if err != nil {
		return err
	}

	syncReq := &protocol.SyncGroupReq{
		GroupID:    cg.groupID,
		MemberID:   cg.memberID,
		Generation: cg.generation,
	}

	requestID := uuid.New().String()
	payloadBytes, err := json.Marshal(syncReq)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req := &protocol.Request{
		Type:      protocol.SyncGroupRequest,
		RequestID: requestID,
		Payload:   payloadBytes,
	}

	resp, err := conn.Send(req)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("sync group failed: %s", resp.Error)
	}

	var syncResp protocol.SyncGroupResp
	if err := json.Unmarshal(resp.Payload, &syncResp); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	if syncResp.Error != "" {
		return fmt.Errorf("sync group error: %s", syncResp.Error)
	}

	cg.claims = syncResp.Assignment

	return nil
}

// leaveGroup leaves the consumer group
func (cg *consumerGroup) leaveGroup() error {
	conn, err := cg.pool.Get()
	if err != nil {
		return err
	}

	leaveReq := &protocol.LeaveGroupReq{
		GroupID:  cg.groupID,
		MemberID: cg.memberID,
	}

	requestID := uuid.New().String()
	payloadBytes, err := json.Marshal(leaveReq)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req := &protocol.Request{
		Type:      protocol.LeaveGroupRequest,
		RequestID: requestID,
		Payload:   payloadBytes,
	}

	resp, err := conn.Send(req)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("leave group failed: %s", resp.Error)
	}

	return nil
}

// consumePartition consumes from a partition
func (cg *consumerGroup) consumePartition(ctx context.Context, session ConsumerGroupSession, handler ConsumerGroupHandler, topic string, partition int32) error {
	// Get initial offset
	offset, err := cg.fetchOffset(topic, partition)
	if err != nil {
		offset = 0 // Start from beginning
	}

	claim := &consumerGroupClaim{
		topic:         topic,
		partition:     partition,
		initialOffset: offset,
		messages:      make(chan *message.Message, 100),
	}

	// Start fetch loop
	go func() {
		defer close(claim.messages)
		ticker := time.NewTicker(cg.config.PollInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				msgs, err := cg.fetchMessages(topic, partition, offset)
				if err != nil {
					cg.errors <- err
					continue
				}

				for _, msg := range msgs {
					select {
					case claim.messages <- msg:
						offset = msg.Offset + 1
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	return handler.ConsumeClaim(session, claim)
}

// fetchOffset fetches the committed offset
func (cg *consumerGroup) fetchOffset(topic string, partition int32) (int64, error) {
	conn, err := cg.pool.Get()
	if err != nil {
		return 0, err
	}

	fetchOffsetReq := &protocol.FetchOffsetReq{
		GroupID:   cg.groupID,
		Topic:     topic,
		Partition: partition,
	}

	requestID := uuid.New().String()
	payloadBytes, err := json.Marshal(fetchOffsetReq)
	if err != nil {
		return 0, err
	}

	req := &protocol.Request{
		Type:      protocol.FetchOffsetRequest,
		RequestID: requestID,
		Payload:   payloadBytes,
	}

	resp, err := conn.Send(req)
	if err != nil {
		return 0, err
	}

	if !resp.Success {
		return 0, fmt.Errorf("fetch offset failed: %s", resp.Error)
	}

	var fetchOffsetResp protocol.FetchOffsetResp
	if err := json.Unmarshal(resp.Payload, &fetchOffsetResp); err != nil {
		return 0, err
	}

	return fetchOffsetResp.Offset, nil
}

// fetchMessages fetches messages from a partition
func (cg *consumerGroup) fetchMessages(topic string, partition int32, offset int64) ([]*message.Message, error) {
	conn, err := cg.pool.Get()
	if err != nil {
		return nil, err
	}

	fetchReq := &protocol.FetchReq{
		Topic:       topic,
		Partition:   partition,
		Offset:      offset,
		MaxMessages: cg.config.FetchMaxMessages,
		MaxBytes:    cg.config.FetchMaxBytes,
	}

	requestID := uuid.New().String()
	payloadBytes, err := json.Marshal(fetchReq)
	if err != nil {
		return nil, err
	}

	req := &protocol.Request{
		Type:      protocol.FetchRequest,
		RequestID: requestID,
		Payload:   payloadBytes,
	}

	resp, err := conn.Send(req)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("fetch failed: %s", resp.Error)
	}

	var fetchResp protocol.FetchResp
	if err := json.Unmarshal(resp.Payload, &fetchResp); err != nil {
		return nil, err
	}

	return fetchResp.Messages, nil
}

// Errors returns the error channel
func (cg *consumerGroup) Errors() <-chan error {
	return cg.errors
}

// Close closes the consumer group
func (cg *consumerGroup) Close() error {
	cg.mu.Lock()
	defer cg.mu.Unlock()

	if cg.closed {
		return nil
	}

	cg.closed = true
	cg.wg.Wait()
	close(cg.errors)

	return cg.pool.Close()
}

// consumerGroupSession implements ConsumerGroupSession
type consumerGroupSession struct {
	cg           *consumerGroup
	ctx          context.Context
	claims       map[string][]int32
	memberID     string
	generationID int32
	offsets      map[string]map[int32]int64
	mu           sync.Mutex
}

func (s *consumerGroupSession) Claims() map[string][]int32 {
	return s.claims
}

func (s *consumerGroupSession) MemberID() string {
	return s.memberID
}

func (s *consumerGroupSession) GenerationID() int32 {
	return s.generationID
}

func (s *consumerGroupSession) MarkMessage(msg *message.Message, metadata string) {
	s.MarkOffset(msg.Topic, msg.Partition, msg.Offset, metadata)
}

func (s *consumerGroupSession) MarkOffset(topic string, partition int32, offset int64, metadata string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.offsets[topic] == nil {
		s.offsets[topic] = make(map[int32]int64)
	}
	s.offsets[topic][partition] = offset
}

func (s *consumerGroupSession) Commit() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for topic, partitions := range s.offsets {
		for partition, offset := range partitions {
			if err := s.commitOffset(topic, partition, offset); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *consumerGroupSession) commitOffset(topic string, partition int32, offset int64) error {
	conn, err := s.cg.pool.Get()
	if err != nil {
		return err
	}

	commitReq := &protocol.CommitOffsetReq{
		GroupID:   s.cg.groupID,
		Topic:     topic,
		Partition: partition,
		Offset:    offset,
	}

	requestID := uuid.New().String()
	payloadBytes, err := json.Marshal(commitReq)
	if err != nil {
		return err
	}

	req := &protocol.Request{
		Type:      protocol.CommitOffsetRequest,
		RequestID: requestID,
		Payload:   payloadBytes,
	}

	resp, err := conn.Send(req)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("commit offset failed: %s", resp.Error)
	}

	return nil
}

func (s *consumerGroupSession) Context() context.Context {
	return s.ctx
}

// consumerGroupClaim implements ConsumerGroupClaim
type consumerGroupClaim struct {
	topic               string
	partition           int32
	initialOffset       int64
	highWaterMarkOffset int64
	messages            chan *message.Message
}

func (c *consumerGroupClaim) Topic() string {
	return c.topic
}

func (c *consumerGroupClaim) Partition() int32 {
	return c.partition
}

func (c *consumerGroupClaim) InitialOffset() int64 {
	return c.initialOffset
}

func (c *consumerGroupClaim) HighWaterMarkOffset() int64 {
	return c.highWaterMarkOffset
}

func (c *consumerGroupClaim) Messages() <-chan *message.Message {
	return c.messages
}
