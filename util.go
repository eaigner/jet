package jet

import (
	"math/rand"
	"sync"
	"time"
)

var (
	gen = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func newQueryId() string {
	return newAlphanumericId(7)
}

var rndMtx sync.Mutex

func newAlphanumericId(length int) string {
	rndMtx.Lock()
	defer rndMtx.Unlock()

	const alpha = "abcdefghijkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ1234567890"
	buf := make([]byte, length)
	for i := 0; i < length; i++ {
		buf[i] = alpha[gen.Intn(len(alpha))]
	}
	return string(buf)
}
