package jet

import (
	"math/rand"
	"time"
)

var (
	gen = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func newAlphanumericId(length int) string {
	const alpha = "abcdefghijkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ1234567890"
	buf := make([]byte, length)
	for i := 0; i < length; i++ {
		buf[i] = alpha[gen.Intn(len(alpha))]
	}
	return string(buf)
}
