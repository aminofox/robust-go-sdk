package protocol

import "github.com/aminofox/robust-go-sdk/pkg/message"

// ProduceReq represents a produce request
type ProduceReq struct {
	Topic     string            `json:"topic"`
	Partition int32             `json:"partition"`
	Key       []byte            `json:"key,omitempty"`
	Value     []byte            `json:"value"`
	Headers   map[string]string `json:"headers,omitempty"`
	Timestamp int64             `json:"timestamp,omitempty"`
}

// FetchReq represents a fetch request
type FetchReq struct {
	Topic       string `json:"topic"`
	Partition   int32  `json:"partition"`
	Offset      int64  `json:"offset"`
	MaxBytes    int    `json:"max_bytes,omitempty"`
	MaxMessages int    `json:"max_messages,omitempty"`
}

// CreateTopicReq represents a create topic request
type CreateTopicReq struct {
	Name       string            `json:"name"`
	Partitions int32             `json:"partitions"`
	Config     map[string]string `json:"config,omitempty"`
}

// DeleteTopicReq represents a delete topic request
type DeleteTopicReq struct {
	Name string `json:"name"`
}

// ListTopicsReq represents a list topics request (empty struct)
type ListTopicsReq struct{}

// MetadataReq represents a metadata request
type MetadataReq struct {
	Topics []string `json:"topics,omitempty"` // Empty means all topics
}

// JoinGroupReq represents a join group request
type JoinGroupReq struct {
	GroupID  string   `json:"group_id"`
	MemberID string   `json:"member_id"`
	ClientID string   `json:"client_id"`
	Topics   []string `json:"topics"`
	Host     string   `json:"host"`
}

// SyncGroupReq represents a sync group request
type SyncGroupReq struct {
	GroupID    string `json:"group_id"`
	MemberID   string `json:"member_id"`
	Generation int32  `json:"generation"`
}

// LeaveGroupReq represents a leave group request
type LeaveGroupReq struct {
	GroupID  string `json:"group_id"`
	MemberID string `json:"member_id"`
}

// CommitOffsetReq represents a commit offset request
type CommitOffsetReq struct {
	GroupID   string `json:"group_id"`
	Topic     string `json:"topic"`
	Partition int32  `json:"partition"`
	Offset    int64  `json:"offset"`
	Metadata  string `json:"metadata,omitempty"`
}

// FetchOffsetReq represents a fetch offset request
type FetchOffsetReq struct {
	GroupID   string `json:"group_id"`
	Topic     string `json:"topic"`
	Partition int32  `json:"partition"`
}

// ProduceResp represents a produce response
type ProduceResp struct {
	Topic     string `json:"topic"`
	Partition int32  `json:"partition"`
	Offset    int64  `json:"offset"`
	Timestamp int64  `json:"timestamp"`
	Error     string `json:"error,omitempty"`
}

// FetchResp represents a fetch response
type FetchResp struct {
	Topic     string             `json:"topic"`
	Partition int32              `json:"partition"`
	Messages  []*message.Message `json:"messages"`
	Error     string             `json:"error,omitempty"`
}

// CreateTopicResp represents a create topic response
type CreateTopicResp struct {
	Name  string `json:"name"`
	Error string `json:"error,omitempty"`
}

// DeleteTopicResp represents a delete topic response
type DeleteTopicResp struct {
	Name  string `json:"name"`
	Error string `json:"error,omitempty"`
}

// ListTopicsResp represents a list topics response
type ListTopicsResp struct {
	Topics []string `json:"topics"`
	Error  string   `json:"error,omitempty"`
}

// TopicMetadata represents metadata for a topic
type TopicMetadata struct {
	Name       string              `json:"name"`
	Partitions []PartitionMetadata `json:"partitions"`
}

// PartitionMetadata represents metadata for a partition
type PartitionMetadata struct {
	ID     int32  `json:"id"`
	Leader string `json:"leader"`
}

// MetadataResp represents a metadata response
type MetadataResp struct {
	Topics []TopicMetadata `json:"topics"`
	Error  string          `json:"error,omitempty"`
}

// JoinGroupResp represents a join group response
type JoinGroupResp struct {
	GroupID    string  `json:"group_id"`
	MemberID   string  `json:"member_id"`
	Generation int64   `json:"generation"`
	Partitions []int32 `json:"partitions"`
	Error      string  `json:"error,omitempty"`
}

// SyncGroupResp represents a sync group response
type SyncGroupResp struct {
	Assignment map[string][]int32 `json:"assignment"`
	Error      string             `json:"error,omitempty"`
}

// LeaveGroupResp represents a leave group response
type LeaveGroupResp struct {
	Error string `json:"error,omitempty"`
}

// CommitOffsetResp represents a commit offset response
type CommitOffsetResp struct {
	Error string `json:"error,omitempty"`
}

// FetchOffsetResp represents a fetch offset response
type FetchOffsetResp struct {
	Offset int64  `json:"offset"`
	Error  string `json:"error,omitempty"`
}
