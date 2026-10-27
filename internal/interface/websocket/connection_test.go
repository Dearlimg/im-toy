package websocket

import (
	"sync"
	"testing"
)

func TestConcurrentSendAndClose(t *testing.T) {
	conn := &Conn{UserID: "test-user", RoomID: "test-room", SendChan: make(chan []byte, 1)}
	mu.Lock()
	rooms[conn.RoomID] = map[*Conn]bool{conn: true}
	userConns[conn.UserID] = conn
	mu.Unlock()

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				sendToConn(conn, []byte("message"))
				SendBytesToUser(conn.UserID, []byte("message"))
			}
		}()
	}
	mu.Lock()
	closeConnLocked(conn)
	mu.Unlock()
	wg.Wait()
	if sendToConn(conn, []byte("late")) {
		t.Fatal("closed connection accepted a message")
	}
	mu.RLock()
	defer mu.RUnlock()
	if _, ok := userConns[conn.UserID]; ok {
		t.Fatal("closed connection remains registered")
	}
	if _, ok := rooms[conn.RoomID]; ok {
		t.Fatal("empty room remains registered")
	}
}

func TestClosingOldConnectionPreservesReplacement(t *testing.T) {
	old := &Conn{UserID: "same-user", RoomID: "old-room", SendChan: make(chan []byte, 1)}
	replacement := &Conn{UserID: old.UserID, RoomID: "new-room", SendChan: make(chan []byte, 1)}
	mu.Lock()
	rooms[old.RoomID] = map[*Conn]bool{old: true}
	rooms[replacement.RoomID] = map[*Conn]bool{replacement: true}
	userConns[old.UserID] = replacement
	closeConnLocked(old)
	if userConns[old.UserID] != replacement {
		mu.Unlock()
		t.Fatal("closing an old connection removed its replacement")
	}
	closeConnLocked(replacement)
	mu.Unlock()
}
