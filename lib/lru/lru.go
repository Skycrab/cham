package lru

import (
	"container/list"
	"sync"
)

type ExpireHandler func(Key, Value)
type Key interface{}
type Value interface{}

type entry struct {
	key   Key
	value Value
}

type Cache struct {
	sync.Mutex
	MaxEntries int           // 0 -> no limit
	OnExpire   ExpireHandler // invoke when expire delete
	ll         *list.List
	cache      map[Key]*list.Element
}

func New(maxEntries int, handler ExpireHandler) *Cache {
	return &Cache{
		MaxEntries: maxEntries,
		OnExpire:   handler,
		ll:         list.New(),
		cache:      make(map[Key]*list.Element),
	}
}

func (c *Cache) Add(key Key, value Value) (ok bool) {
	c.Lock()
	defer c.Unlock()
	if e, ok := c.cache[key]; ok {
		c.ll.MoveToFront(e)
		e.Value.(*entry).value = value
	} else {
		e = c.ll.PushFront(&entry{key, value})
		c.cache[key] = e
	}
	if c.MaxEntries != 0 && c.ll.Len() > c.MaxEntries {
		c.RemoveOldest()
	}
	return
}

func (c *Cache) Get(key Key) (Value, bool) {
	c.Lock()
	defer c.Unlock()
	if e, ok := c.cache[key]; ok {
		c.ll.MoveToFront(e)
		return e.Value.(*entry).value, true
	} else {
		return nil, false
	}
}

func (c *Cache) Len() int {
	c.Lock()
	defer c.Unlock()
	return c.ll.Len()
}

func (c *Cache) RemoveOldest() {
	e := c.ll.Back()
	c.ll.Remove(e)
	kv := e.Value.(*entry)
	delete(c.cache, kv.key)
	if c.OnExpire != nil {
		c.OnExpire(kv.key, kv.value)
	}
}
