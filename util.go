package jet

import (
	"math/rand"
)

var alpha = "abcdefghijkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ1234567890"

func newAlphanumericId(length int) string {
	buf := make([]byte, length)
	for i := 0; i < length; i++ {
		buf[i] = alpha[rand.Intn(len(alpha))]
	}
	return string(buf)
}
