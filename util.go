package jet

import (
	"math/rand"
	"time"
)

var (
	gen   = rand.New(rand.NewSource(time.Now().UnixNano()))
	alpha = "abcdefghijkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ1234567890"
)

func newAlphanumericId(length int) string {
	buf := make([]byte, length)
	for i := 0; i < length; i++ {
		buf[i] = alpha[gen.Intn(len(alpha))]
	}
	return string(buf)
}
