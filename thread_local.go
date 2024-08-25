package vermouth

import (
    "runtime"
    "strconv"
    "strings"
    "sync"
)

type ThreadLocal struct {
    mu       sync.RWMutex
    store    map[uint64]interface{}
    parentID map[uint64]uint64
}

func NewThreadLocal() *ThreadLocal {
    return &ThreadLocal{
        store:    make(map[uint64]interface{}),
        parentID: make(map[uint64]uint64),
    }
}

func (tl *ThreadLocal) Set(value interface{}) {
    tl.mu.Lock()
    defer tl.mu.Unlock()
    id := goID()
    tl.store[id] = value
}

func (tl *ThreadLocal) Get() interface{} {
    tl.mu.RLock()
    defer tl.mu.RUnlock()
    id := goID()
    if value, ok := tl.store[id]; ok {
        return value
    }
    if parentID, ok := tl.parentID[id]; ok {
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