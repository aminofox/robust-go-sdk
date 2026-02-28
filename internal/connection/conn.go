package connection

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/aminofox/robust-go-sdk/internal/protocol"
	sdkerrors "github.com/aminofox/robust-go-sdk/pkg/errors"
)

// Conn represents a connection to a Robust broker
type Conn interface {
	// Send sends a request and returns the response
	Send(req *protocol.Request) (*protocol.Response, error)

	// IsAlive checks if connection is alive
	IsAlive() bool

	// Close closes the connection
	Close() error

	// Addr returns the broker address
	Addr() string
}

// robustConn implements Conn interface
type robustConn struct {
	addr     string
	conn     net.Conn
	reader   *bufio.Reader
	writer   *bufio.Writer
	encoder  *json.Encoder
	decoder  *json.Decoder
	mu       sync.Mutex
	lastUsed time.Time
	closed   bool
}

// Dial creates a new connection to the broker
func Dial(addr string, timeout time.Duration) (Conn, error) {
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return nil, &sdkerrors.ConnectionError{
			Broker: addr,
			Err:    err,
		}
	}

	rc := &robustConn{
		addr:     addr,
		conn:     conn,
		reader:   bufio.NewReader(conn),
		writer:   bufio.NewWriter(conn),
		lastUsed: time.Now(),
	}

	rc.encoder = json.NewEncoder(rc.writer)
	rc.decoder = json.NewDecoder(rc.reader)

	return rc, nil
}

// Send sends a request and waits for response
func (c *robustConn) Send(req *protocol.Request) (*protocol.Response, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil, sdkerrors.ErrClosed
	}

	// Set write deadline
	if err := c.conn.SetWriteDeadline(time.Now().Add(30 * time.Second)); err != nil {
		return nil, err
	}

	// Encode and send request
	if err := c.encoder.Encode(req); err != nil {
		return nil, fmt.Errorf("encode request: %w", err)
	}

	// Flush writer
	if err := c.writer.Flush(); err != nil {
		return nil, fmt.Errorf("flush writer: %w", err)
	}

	// Set read deadline
	if err := c.conn.SetReadDeadline(time.Now().Add(30 * time.Second)); err != nil {
		return nil, err
	}

	// Read response
	var resp protocol.Response
	if err := c.decoder.Decode(&resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	// Verify request ID matches
	if resp.RequestID != req.RequestID {
		return nil, fmt.Errorf("request ID mismatch: expected %s, got %s",
			req.RequestID, resp.RequestID)
	}

	c.lastUsed = time.Now()
	return &resp, nil
}

// IsAlive checks if the connection is still alive
func (c *robustConn) IsAlive() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return false
	}

	// Set a short deadline for the test
	if err := c.conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond)); err != nil {
		return false
	}

	// Try to peek one byte
	_, err := c.reader.Peek(1)

	// Reset deadline
	c.conn.SetReadDeadline(time.Time{})

	return err == nil || err == bufio.ErrBufferFull
}

// Close closes the connection
func (c *robustConn) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true
	return c.conn.Close()
}

// Addr returns the broker address
func (c *robustConn) Addr() string {
	return c.addr
}
