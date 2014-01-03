package jet

import (
	"container/list"
	"crypto/sha1"
	"database/sql"
	"sync"
)

type lru struct {
	m        sync.Mutex
	maxItems int
	keys     map[string]*list.Element
	list     *list.List
}

type lruItem struct {
	key  string
	stmt *sql.Stmt
}

func newLru() *lru {
	return &lru{
		maxItems: 500,
		keys:     make(map[string]*list.Element),
		list:     list.New(),
	}
}

func (c *lru) get(k string) (*sql.Stmt, bool) {
	c.m.Lock()
	defer c.m.Unlock()
	k = makeKey(k)
	e, ok := c.keys[k]
	if ok {
		return e.Value.(*lruItem).stmt, ok
	}
	return nil, false
}

func (c *lru) put(k string, stmt *sql.Stmt) {
	c.m.Lock()
	defer c.m.Unlock()
	k = makeKey(k)
	e, ok := c.keys[k]
	if ok {
		c.list.MoveToFront(e)
	} else {
		c.keys[k] = c.list.PushFront(&lruItem{
			key:  k,
			stmt: stmt,
		})
	}
	c.clean()
}

func (c *lru) del(k string) {
	c.m.Lock()
	defer c.m.Unlock()
	k = makeKey(k)
	e, ok := c.keys[k]
	if ok {
		c.delElem(e)
	}
}

func (c *lru) delElem(e *list.Element) {
	item := c.list.Remove(e).(*lruItem)
	defer item.stmt.Close()
	delete(c.keys, item.key)
}

func (c *lru) clean() {
	n := c.list.Len() - c.maxItems
	if n > 0 {
		for i := 0; i < n; i++ {
			c.delElem(c.list.Back())
		}
	}
}

// makeKey hashes the key to save some bytes
func makeKey(k string) string {
	buffer := sha1.New()
	buffer.Write([]byte(k))
	return string(buffer.Sum(nil))
}
