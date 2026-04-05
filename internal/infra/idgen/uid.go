package idgen

import (
	"fmt"
	"sync"
	"time"
)

type UIDGenerator struct {
	mu        sync.Mutex
	lastStamp int64
	sequence  int64
}

var Gen = &UIDGenerator{}

func (g *UIDGenerator) NextUID() string {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := time.Now().UnixMilli() % 1000000
	seq := g.sequence % 1000

	uid := now*1000 + seq
	g.sequence++

	return fmt.Sprintf("%09d", uid)
}
