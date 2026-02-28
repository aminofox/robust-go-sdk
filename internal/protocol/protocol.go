package protocol

// RequestType defines the type of request
type RequestType int32

const (
	// ProduceRequest sends messages to broker
	ProduceRequest RequestType = iota
	// FetchRequest fetches messages from broker
	FetchRequest
	// CreateTopicRequest creates a new topic
	CreateTopicRequest
	// DeleteTopicRequest deletes a topic
	DeleteTopicRequest
	// ListTopicsRequest lists all topics
	ListTopicsRequest
	// MetadataRequest requests topic metadata
	MetadataRequest
	// HeartbeatRequest sends heartbeat
	HeartbeatRequest
	// JoinGroupRequest joins a consumer group
	JoinGroupRequest
	// LeaveGroupRequest leaves a consumer group
	LeaveGroupRequest
	// SyncGroupRequest syncs consumer group state
	SyncGroupRequest
	// CommitOffsetRequest commits consumer offset
	CommitOffsetRequest
	// FetchOffsetRequest fetches committed offset
	FetchOffsetRequest
	// ListGroupsRequest lists consumer groups
	ListGroupsRequest
	// ClusterInfoRequest requests cluster information
	ClusterInfoRequest
)

// Request represents a client request to the broker
type Request struct {
	Type      RequestType `json:"type"`       // Request type
	RequestID string      `json:"request_id"` // Unique request ID
	Payload   []byte      `json:"payload"`    // Request payload
}

// Response represents a broker response to client
type Response struct {
	RequestID string `json:"request_id"`        // Matches request ID
	Success   bool   `json:"success"`           // Whether request succeeded
	Error     string `json:"error,omitempty"`   // Error message if any
	Payload   []byte `json:"payload,omitempty"` // Response payload
}

// String returns string representation of request type
func (r RequestType) String() string {
	switch r {
	case ProduceRequest:
		return "ProduceRequest"
	case FetchRequest:
		return "FetchRequest"
	case CreateTopicRequest:
		return "CreateTopicRequest"
	case DeleteTopicRequest:
		return "DeleteTopicRequest"
	case ListTopicsRequest:
		return "ListTopicsRequest"
	case MetadataRequest:
		return "MetadataRequest"
	case HeartbeatRequest:
		return "HeartbeatRequest"
	case JoinGroupRequest:
		return "JoinGroupRequest"
	case LeaveGroupRequest:
		return "LeaveGroupRequest"
	case SyncGroupRequest:
		return "SyncGroupRequest"
	case CommitOffsetRequest:
		return "CommitOffsetRequest"
	case FetchOffsetRequest:
		return "FetchOffsetRequest"
	case ListGroupsRequest:
		return "ListGroupsRequest"
	case ClusterInfoRequest:
		return "ClusterInfoRequest"
	default:
		return "Unknown"
	}
}
