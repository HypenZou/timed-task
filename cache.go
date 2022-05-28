package timedtask

import (
	"sync"
	"timedtask/errors"
)

const (
	PUT uint16 = iota
	DEL
)

type Cache struct {
	buf []*cacheNode
	mu  sync.Mutex
}

type cacheNode struct {
	key    []byte
	value  []byte
	method uint16
}

func NewCache() *Cache {
	return &Cache{}
}

func (ca *Cache) Put(key, value []byte, method uint16) {
	ca.mu.Lock()
	ca.buf = append(ca.buf, &cacheNode{key, value, method})
	ca.mu.Unlock()
}

func (ca *Cache) Get() (key, value []byte, method uint16, err error) {
	ca.mu.Lock()
	defer ca.mu.Unlock()
	if len(ca.buf) == 0 {
		err = errors.ErrNotFound
		return
	}
	key = ca.buf[0].key
	value = ca.buf[0].value
	method = ca.buf[0].method
	ca.buf = ca.buf[1:]
	return
}

func (ca *Cache) IsEmpty() bool {
	ca.mu.Lock()
	defer ca.mu.Unlock()
	return len(ca.buf) == 0
}
