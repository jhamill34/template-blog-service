package repositories

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/jhamill34/notion-provisioner/internal/config"
	"golang.org/x/crypto/argon2"
)

func comparePasswords(password, encodedPassword string) (bool, error) {
	params, salt, storedHash, err := decodeHash(encodedPassword)

	if err != nil {
		return false, err
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		params.Iterations,
		params.Memory,
		params.Parallelism,
		params.HashLength,
	)

	if subtle.ConstantTimeCompare(storedHash, hash) == 1 {
		return true, nil
	}

	return false, nil
}

// "$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",

func createHash(params *config.HashParams, password string) (string, error) {
	salt, err := randomBytes(params.SaltLength)
	if err != nil {
		return "", err
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		params.Iterations,
		params.Memory,
		params.Parallelism,
		params.HashLength,
	)
	base64hash := base64.RawStdEncoding.EncodeToString(hash)
	base64salt := base64.RawStdEncoding.EncodeToString(salt)

	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		params.Memory,
		params.Iterations,
		params.Parallelism,
		base64salt,
		base64hash,
	), nil
}

func decodeHash(
	encodedPassword string,
) (params *config.HashParams, salt []byte, hash []byte, err error) {
	vals := strings.Split(encodedPassword, "$")

	if len(vals) != 6 || vals[0] != "" || vals[1] != "argon2id" {
		return nil, nil, nil, fmt.Errorf("Invalid hash format")
	}

	var version int
	_, err = fmt.Sscanf(vals[2], "v=%d", &version)
	if err != nil {
		return nil, nil, nil, err
	}

	if version != argon2.Version {
		return nil, nil, nil, fmt.Errorf("Invalid hash version")
	}

	params = &config.HashParams{}
	_, err = fmt.Sscanf(
		vals[3],
		"m=%d,t=%d,p=%d",
		&params.Memory,
		&params.Iterations,
		&params.Parallelism,
	)
	if err != nil {
		return nil, nil, nil, err
	}

	salt, err = base64.RawStdEncoding.Strict().DecodeString(vals[4])
	if err != nil {
		return nil, nil, nil, err
	}
	params.SaltLength = uint32(len(salt))

	hash, err = base64.RawStdEncoding.Strict().DecodeString(vals[5])
	if err != nil {
		return nil, nil, nil, err
	}
	params.HashLength = uint32(len(hash))

	return params, salt, hash, nil
}

func randomBytes(length uint32) ([]byte, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return nil, err
	}

	return bytes, nil
}
