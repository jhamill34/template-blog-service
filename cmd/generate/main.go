package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
)

const (
	iterations  = 3
	memory      = 32 * 1024
	parallelism = 4
	length      = 32
)

func main() {
	fmt.Printf("ID: %s\n", uuid.New().String())

	password := os.Args[1]
	password = strings.TrimSpace(password)

	salt, err  := randomBytes(16)
	if err != nil {
		panic(err)
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt, 
		iterations,
		memory,
		parallelism,
		length,
	)
	base64hash := base64.RawStdEncoding.EncodeToString(hash)
	base64salt := base64.RawStdEncoding.EncodeToString(salt)

	encodedPassword := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		memory,
		iterations,
		parallelism,
		base64salt,
		base64hash,
	)

	fmt.Println(encodedPassword)
}

func randomBytes(length int) ([]byte, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return nil, err
	}

	return bytes, nil
}
