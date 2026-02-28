package connection

import (
	"fmt"
	"sync"
	"time"

	sdkerrors "github.com/aminofox/robust-go-sdk/pkg/errors"
)

// Pool manages a pool of connections to brokers
type Pool struct {
	brokers []string
	size    int
	timeout time.Duration

	mu     sync.RWMutex
	conns  []Conn
	index  int // Round-robin index
	closed bool
}

// PoolConfig configures a connection pool
type PoolConfig struct {
	Brokers        []string
	Size           int
	ConnectTimeout time.Duration
}

// NewPool creates a new connection pool
func NewPool(config *PoolConfig) (*Pool, error) {
	if len(config.Brokers) == 0 {
		return nil, fmt.Errorf("no brokers specified")
	}

	if config.Size <= 0 {
		config.Size = 5
	}

	if config.ConnectTimeout == 0 {
		config.ConnectTimeout = 10 * time.Second
	}

	pool := &Pool{
		brokers: config.Brokers,
		size:    config.Size,
		timeout: config.ConnectTimeout,
		conns:   make([]Conn, 0, config.Size),
	}

	return pool, nil
}

// Get returns a connection from the pool
func (p *Pool) Get() (Conn, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil, sdkerrors.ErrClosed
	}

	// Try to find an alive connection
	for _, conn := range p.conns {
		if conn.IsAlive() {
			return conn, nil
		}
	}

	// Create new connection if pool not full
	if len(p.conns) < p.size {
		conn, err := p.createConnection()
		if err != nil {
			return nil, err
		}
		p.conns = append(p.conns, conn)
		return conn, nil
	}

	// Replace dead connection
	conn, err := p.createConnection()
	if err != nil {
		return nil, err
	}

	// Replace oldest connection
	if len(p.conns) > 0 {
		oldConn := p.conns[0]
		oldConn.Close()
		p.conns[0] = conn
	}

	return conn, nil
}

// createConnection creates a new connection to a broker (round-robin)
func (p *Pool) createConnection() (Conn, error) {
	broker := p.brokers[p.index%len(p.brokers)]
	p.index++

	conn, err := Dial(broker, p.timeout)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// Close closes all connections in the pool
func (p *Pool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}

	p.closed = true

	for _, conn := range p.conns {
		conn.Close()
	}

	p.conns = nil
	return nil
}

// Size returns the current number of connections
func (p *Pool) Size() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.conns)
}

// ActiveConnections returns the number of alive connections
func (p *Pool) ActiveConnections() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	count := 0
	for _, conn := range p.conns {
		if conn.IsAlive() {
			count++
		}
	}
	return count
}
