package internal

import (
	"encoding/base64"
	"math/rand"
)

func GenerateId(n int) string {
	num_bytes := (n >> 3)

	data := make([]byte, num_bytes)

	for i := 0; i < num_bytes; i++ {
		data[i] = byte(rand.Intn(256))
	}

	return base64.StdEncoding.EncodeToString(data)
}
