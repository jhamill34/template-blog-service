package session

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jhamill34/notion-provisioner/internal/database"
	"github.com/jhamill34/notion-provisioner/internal/models"
	"github.com/jhamill34/notion-provisioner/internal/services"
	"github.com/redis/go-redis/v9"
)

const PREFIX = "session:"

type KeyValueSessionStore struct {
	keyValueStore database.KeyValueStoreProvider
	ttl           time.Duration
	key           []byte
}

func NewRedisSessionStore(
	keyValueStore database.KeyValueStoreProvider,
	ttl time.Duration,
	key []byte,
) *KeyValueSessionStore {
	return &KeyValueSessionStore{
		keyValueStore: keyValueStore,
		ttl:           ttl,
		key:           key,
	}
}

// Create implements services.SessionService.
func (self *KeyValueSessionStore) Create(ctx context.Context, data *models.SessionData) string {
	var err error
	id := uuid.New().String()
	key := PREFIX + id

	b64Data := encodeData(data)

	salt := randomString(32)

	payload := b64Data + "/" + salt
	signature := signData([]byte(payload), self.key)
	value := fmt.Sprintf("%s.%s", payload, signature)

	err = self.keyValueStore.Get().Set(ctx, key, value, 0)
	if err != nil {
		panic(err)
	}

	err = self.keyValueStore.Get().Expire(ctx, key, self.ttl)
	if err != nil {
		panic(err)
	}

	return id
}

// Find implements services.SessionService.
func (self *KeyValueSessionStore) Find(
	ctx context.Context,
	id string,
	result *models.SessionData,
) models.Notifier {
	var err error
	key := PREFIX + id

	val, err := self.keyValueStore.Get().Get(ctx, key)
	if err == redis.Nil {
		return services.SessionNotFound
	}
	if err != nil {
		panic(err)
	}

	parts := strings.Split(val, ".")
	if len(parts) != 2 {
		return services.MalformedSession
	}

	b64Data := []byte(parts[0])
	signature := []byte(parts[1])

	if !verify(b64Data, signature, self.key) {
		return services.MalformedSession
	}

	saltIndex := strings.Index(string(b64Data), "/")
	if saltIndex == -1 {
		return services.MalformedSession
	}

	b64Data = b64Data[:saltIndex]
	decodeData(string(b64Data), result)

	err = self.keyValueStore.Get().Expire(ctx, key, self.ttl)
	if err != nil {
		panic(err)
	}

	return nil
}

func (self *KeyValueSessionStore) UpdateCsrf(
	ctx context.Context,
	id, csrfToken string,
) models.Notifier {
	var err error
	key := PREFIX + id

	val, err := self.keyValueStore.Get().Get(ctx, key)
	if err == redis.Nil {
		return services.SessionNotFound
	}
	if err != nil {
		panic(err)
	}

	parts := strings.Split(val, ".")
	if len(parts) != 2 {
		return services.MalformedSession
	}

	b64Data := []byte(parts[0])
	signature := []byte(parts[1])

	if !verify(b64Data, signature, self.key) {
		return services.MalformedSession
	}

	saltIndex := strings.Index(string(b64Data), "/")

	salt := b64Data[saltIndex+1:]

	data := models.SessionData{}
	decodeData(string(b64Data[:saltIndex]), &data)
	data.CsrfToken = csrfToken
	newData := encodeData(data)

	payload := newData + "/" + string(salt)
	newSignature := signData([]byte(payload), self.key)
	value := fmt.Sprintf("%s.%s", payload, newSignature)

	err = self.keyValueStore.Get().Set(ctx, key, value, 0)
	if err != nil {
		panic(err)
	}

	err = self.keyValueStore.Get().Expire(ctx, key, self.ttl)
	if err != nil {
		panic(err)
	}

	return nil
}

// Destroy implements services.SessionService.
func (self *KeyValueSessionStore) Destroy(ctx context.Context, id string) {
	err := self.keyValueStore.Get().Del(ctx, PREFIX+id)
	if err != nil {
		panic(err)
	}
}

// Create implements services.SessionService.
func randomString(n int) string {
	buffer := make([]byte, n)

	if _, err := rand.Read(buffer); err != nil {
		panic(err)
	}

	return base64.URLEncoding.EncodeToString(buffer)
}

func encodeData(data interface{}) string {
	encoded, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	return base64.URLEncoding.EncodeToString(encoded)
}

func decodeData(data string, result interface{}) {
	decoded, err := base64.URLEncoding.DecodeString(data)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(decoded, result)
	if err != nil {
		panic(err)
	}
}

func signData(data []byte, key []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	sig := mac.Sum(nil)

	return base64.URLEncoding.EncodeToString(sig)
}

func verify(data []byte, sig []byte, key []byte) bool {
	expectedSig := signData(data, key)
	return hmac.Equal(sig, []byte(expectedSig))
}

// var _ services.SessionService = (*RedisSessionStore)(nil)
