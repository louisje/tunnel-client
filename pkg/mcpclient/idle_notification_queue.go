package mcpclient

import (
	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"sync"
)

// idleNotificationQueue is a bounded buffer for unsolicited notifications.
// Responses and server requests must remain on the normal request path.
type idleNotificationQueue struct {
	mu         sync.Mutex
	pending    []*jsonrpc.Request
	limit      int
	generation uint64
}

func newIdleNotificationQueue(limit int) *idleNotificationQueue {
	if limit < 1 {
		limit = 1
	}
	return &idleNotificationQueue{limit: limit}
}

func (q *idleNotificationQueue) offer(msg jsonrpc.Message) bool {
	notification, ok := msg.(*jsonrpc.Request)
	if !ok || notification.ID.IsValid() {
		return false
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.pending) >= q.limit {
		return false
	}
	q.pending = append(q.pending, notification)
	return true
}

// reset discards buffered notifications when the physical MCP session changes.
func (q *idleNotificationQueue) reset(generation uint64) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.generation = generation
	q.pending = nil
}

// offerForGeneration prevents a delayed notification from an older session
// from entering the current session queue.
func (q *idleNotificationQueue) offerForGeneration(generation uint64, msg jsonrpc.Message) bool {
	notification, ok := msg.(*jsonrpc.Request)
	if !ok || notification.ID.IsValid() {
		return false
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	if generation != q.generation || len(q.pending) >= q.limit {
		return false
	}
	q.pending = append(q.pending, notification)
	return true
}

func (q *idleNotificationQueue) drain() []*jsonrpc.Request {
	q.mu.Lock()
	defer q.mu.Unlock()
	pending := q.pending
	q.pending = nil
	return pending
}
