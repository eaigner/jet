package jet

import (
	"database/sql"
	"sort"
	"sync"
	"time"
)

type lruCache struct {
	m   map[string]*lruItem
	max int
	mtx sync.Mutex
}

type lruItem struct {
	key        string
	stmt       *sql.Stmt
	lastAccess time.Time
}

type lruList []*lruItem

func (l lruList) Len() int {
	return len(l)
}

func (l lruList) Less(i, j int) bool {
	return l[i].lastAccess.Before(l[j].lastAccess)
}

func (l lruList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func newLRUCache(max int) *lruCache {
	c := &lruCache{max: max}
	c.reset()
	return c
}

func (c *lruCache) set(key string, stmt *sql.Stmt) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	c.m[key] = &lruItem{
		key:        key,
		stmt:       stmt,
		lastAccess: time.Now(),
	}
	c.cleanIfNeeded()
}

func (c *lruCache) get(key string) *sql.Stmt {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if v, ok := c.m[key]; ok {
		return v.stmt
	}
	return nil
}

func (c *lruCache) cleanIfNeeded() {
	if c.max == 0 {
		c.max = 20
	}
	if len(c.m) > c.max {
		c.clean()
	}
}

func (c *lruCache) clean() {
	a := make(lruList, 0, len(c.m))
	for _, v := range c.m {
		a = append(a, v)
	}
	sort.Sort(sort.Reverse(a))
	for _, v := range a[c.max:] {
		delete(c.m, v.key)
	}
}

func (c *lruCache) reset() {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	c.m = make(map[string]*lruItem, c.max)
}
