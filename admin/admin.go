package admin

import (
	"encoding/json"
	"fmt"

	"github.com/aminofox/robust-go-sdk/internal/connection"
	"github.com/aminofox/robust-go-sdk/internal/protocol"
	sdkerrors "github.com/aminofox/robust-go-sdk/pkg/errors"
	"github.com/google/uuid"
)

// AdminClient provides administrative operations for Robust
type AdminClient interface {
	// CreateTopic creates a new topic
	CreateTopic(name string, partitions int32, config map[string]string) error

	// DeleteTopic deletes a topic
	DeleteTopic(name string) error

	// ListTopics returns a list of all topics
	ListTopics() ([]string, error)

	// GetMetadata returns metadata for the specified topics (empty = all topics)
	GetMetadata(topics []string) ([]protocol.TopicMetadata, error)

	// Close closes the admin client
	Close() error
}

// adminClient implements AdminClient
type adminClient struct {
	config *Config
	pool   *connection.Pool
	closed bool
}

// NewAdminClient creates a new admin client
func NewAdminClient(config *Config) (AdminClient, error) {
	if config == nil {
		config = DefaultConfig()
	}

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

	ac := &adminClient{
		config: config,
		pool:   pool,
	}

	return ac, nil
}

// CreateTopic creates a new topic
func (ac *adminClient) CreateTopic(name string, partitions int32, config map[string]string) error {
	if ac.closed {
		return sdkerrors.ErrClosed
	}

	conn, err := ac.pool.Get()
	if err != nil {
		return err
	}

	createReq := &protocol.CreateTopicReq{
		Name:       name,
		Partitions: partitions,
		Config:     config,
	}

	requestID := uuid.New().String()
	payloadBytes, err := json.Marshal(createReq)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req := &protocol.Request{
		Type:      protocol.CreateTopicRequest,
		RequestID: requestID,
		Payload:   payloadBytes,
	}

	resp, err := conn.Send(req)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("create topic failed: %s", resp.Error)
	}

	var createResp protocol.CreateTopicResp
	if err := json.Unmarshal(resp.Payload, &createResp); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	if createResp.Error != "" {
		return fmt.Errorf("create topic error: %s", createResp.Error)
	}

	return nil
}

// DeleteTopic deletes a topic
func (ac *adminClient) DeleteTopic(name string) error {
	if ac.closed {
		return sdkerrors.ErrClosed
	}

	conn, err := ac.pool.Get()
	if err != nil {
		return err
	}

	deleteReq := &protocol.DeleteTopicReq{
		Name: name,
	}

	requestID := uuid.New().String()
	payloadBytes, err := json.Marshal(deleteReq)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req := &protocol.Request{
		Type:      protocol.DeleteTopicRequest,
		RequestID: requestID,
		Payload:   payloadBytes,
	}

	resp, err := conn.Send(req)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("delete topic failed: %s", resp.Error)
	}

	var deleteResp protocol.DeleteTopicResp
	if err := json.Unmarshal(resp.Payload, &deleteResp); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	if deleteResp.Error != "" {
		return fmt.Errorf("delete topic error: %s", deleteResp.Error)
	}

	return nil
}

// ListTopics returns a list of all topics
func (ac *adminClient) ListTopics() ([]string, error) {
	if ac.closed {
		return nil, sdkerrors.ErrClosed
	}

	conn, err := ac.pool.Get()
	if err != nil {
		return nil, err
	}

	listReq := &protocol.ListTopicsReq{}
	requestID := uuid.New().String()
	payloadBytes, err := json.Marshal(listReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req := &protocol.Request{
		Type:      protocol.ListTopicsRequest,
		RequestID: requestID,
		Payload:   payloadBytes,
	}

	resp, err := conn.Send(req)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("list topics failed: %s", resp.Error)
	}

	var listResp protocol.ListTopicsResp
	if err := json.Unmarshal(resp.Payload, &listResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if listResp.Error != "" {
		return nil, fmt.Errorf("list topics error: %s", listResp.Error)
	}

	return listResp.Topics, nil
}

// GetMetadata returns topic metadata
func (ac *adminClient) GetMetadata(topics []string) ([]protocol.TopicMetadata, error) {
	if ac.closed {
		return nil, sdkerrors.ErrClosed
	}

	conn, err := ac.pool.Get()
	if err != nil {
		return nil, err
	}

	metadataReq := &protocol.MetadataReq{
		Topics: topics,
	}

	requestID := uuid.New().String()
	payloadBytes, err := json.Marshal(metadataReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req := &protocol.Request{
		Type:      protocol.MetadataRequest,
		RequestID: requestID,
		Payload:   payloadBytes,
	}

	resp, err := conn.Send(req)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("metadata failed: %s", resp.Error)
	}

	var metadataResp protocol.MetadataResp
	if err := json.Unmarshal(resp.Payload, &metadataResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if metadataResp.Error != "" {
		return nil, fmt.Errorf("metadata error: %s", metadataResp.Error)
	}

	return metadataResp.Topics, nil
}

// Close closes the admin client
func (ac *adminClient) Close() error {
	if ac.closed {
		return nil
	}

	ac.closed = true
	return ac.pool.Close()
}
