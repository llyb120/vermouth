package vermouth

import (
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ThreadLocal struct {
	mu       sync.RWMutex
	store    map[uint64]interface{}
	parentID map[uint64]uint64
	expiry   map[uint64]time.Time
	ttl      time.Duration
}

func NewThreadLocalWithTTL(ttl time.Duration) *ThreadLocal {
	tl := &ThreadLocal{
		store:    make(map[uint64]interface{}),
		parentID: make(map[uint64]uint64),
		expiry:   make(map[uint64]time.Time),
		ttl:      ttl,
	}
	go tl.cleanup()
	return tl
}

func NewThreadLocal() *ThreadLocal {
	// 默认30分钟
	return NewThreadLocalWithTTL(30 * time.Minute)
}

func (tl *ThreadLocal) Set(value interface{}) {
	tl.mu.Lock()
	defer tl.mu.Unlock()
	id := goID()
	tl.store[id] = value
	tl.expiry[id] = time.Now().Add(tl.ttl)
}

func (tl *ThreadLocal) Get() interface{} {
	tl.mu.RLock()
	defer tl.mu.RUnlock()
	id := goID()
	if value, ok := tl.store[id]; ok {
		tl.expiry[id] = time.Now().Add(tl.ttl)
		return value
	}
	if parentID, ok := tl.parentID[id]; ok {
		tl.expiry[parentID] = time.Now().Add(tl.ttl)
		return tl.store[parentID]
	}
	return nil
}

func (tl *ThreadLocal) Go(f func()) {
	parentID := goID()
	go func() {
		tl.mu.Lock()
		tl.parentID[goID()] = parentID
		tl.mu.Unlock()
		f()
	}()
}

func (tl *ThreadLocal) cleanup() {
	for {
		time.Sleep(tl.ttl)
		tl.mu.Lock()
		now := time.Now()
		for id, expiry := range tl.expiry {
			if now.After(expiry) {
				delete(tl.store, id)
				delete(tl.expiry, id)
				delete(tl.parentID, id)
			}
		}
		tl.mu.Unlock()
	}
}

// 获取当前 goroutine 的 ID
func goID() uint64 {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, err := strconv.ParseUint(idField, 10, 64)
	if err != nil {
		return 0
	}
	return id
}
