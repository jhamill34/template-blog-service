package repositories

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jhamill34/notion-provisioner/internal/config"
	"github.com/jhamill34/notion-provisioner/internal/models"
	"github.com/jhamill34/notion-provisioner/internal/services"
	"github.com/redis/go-redis/v9"
)

type VerificationType string

const (
	VerificationTypeRegistration   VerificationType = "registration:"
	VerificationTypeForgotPassword VerificationType = "forgot_password:"
	VerificationTypeInvite         VerificationType = "invite:"
	VerificationTypeAuthCode       VerificationType = "auth_code:"
)

type HashedVerifyTokenRepository struct {
	redisClient    *redis.Client
	ttl            time.Duration
	prefix         VerificationType
	passwordParams *config.HashParams
}

func NewHashedVerifyTokenRepository(
	redisClient *redis.Client,
	ttl time.Duration,
	prefix VerificationType,
	passwordParams *config.HashParams,
) *HashedVerifyTokenRepository {
	return &HashedVerifyTokenRepository{
		redisClient:    redisClient,
		ttl:            ttl,
		prefix:         prefix,
		passwordParams: passwordParams,
	}
}

// Create implements services.VerifyTokenService.
func (self *HashedVerifyTokenRepository) Create(
	ctx context.Context,
	id string,
) string {
	publicToken := uuid.New().String()
	token, err := createHash(self.passwordParams, publicToken)

	err = self.redisClient.Set(ctx, string(self.prefix)+id, token, self.ttl).Err()
	if err != nil {
		panic(err)
	}

	return publicToken
}

// Verify implements services.VerifyTokenService.
func (self *HashedVerifyTokenRepository) Verify(
	ctx context.Context,
	id, token string,
) models.Notifier {
	hashedToken, err := self.redisClient.Get(ctx, string(self.prefix)+id).Result()
	if err == redis.Nil {
		return services.TokenNotFound
	}

	if err != nil {
		panic(err)
	}

	ok, err := comparePasswords(token, hashedToken)
	if err != nil {
		panic(err)
	}

	if ok {
		err = self.redisClient.Del(ctx, string(self.prefix)+id).Err()
		if err != nil {
			panic(err)
		}

		return nil
	}

	return services.InvalidToken
}

// CreateWithClaims implements services.TokenClaimsService.
func (self *HashedVerifyTokenRepository) CreateWithClaims(
	ctx context.Context,
	id string,
	data interface{},
) string {
	publicTokenBytes, err := randomBytes(32)
	if err != nil {
		panic(err)
	}
	publicToken := base64.RawURLEncoding.EncodeToString(publicTokenBytes)

	token, err := createHash(self.passwordParams, publicToken)
	encodedToken := base64.RawURLEncoding.EncodeToString([]byte(token))

	jsonBuffer := bytes.Buffer{}
	json.NewEncoder(&jsonBuffer).Encode(data)
	encodedClaims := base64.RawURLEncoding.EncodeToString(jsonBuffer.Bytes())

	output := encodedToken + "." + encodedClaims

	mac := hmac.New(sha256.New, []byte(self.passwordParams.Secret))
	mac.Write([]byte(output))
	signature := mac.Sum(nil)
	encodedSignaure := base64.RawURLEncoding.EncodeToString(signature)

	output += "." + encodedSignaure

	err = self.redisClient.Set(ctx, string(self.prefix)+id, output, self.ttl).Err()
	if err != nil {
		panic(err)
	}

	return publicToken
}

// VerifyWithClaims implements services.TokenClaimsService.
func (self *HashedVerifyTokenRepository) VerifyWithClaims(
	ctx context.Context,
	id string,
	token string,
	data interface{},
) models.Notifier {
	hashedToken, err := self.redisClient.Get(ctx, string(self.prefix)+id).Result()
	if err == redis.Nil {
		return services.TokenNotFound
	}

	if err != nil {
		panic(err)
	}

	parts := strings.Split(hashedToken, ".")
	if len(parts) != 3 {
		return services.InvalidToken
	}

	mac := hmac.New(sha256.New, []byte(self.passwordParams.Secret))
	mac.Write([]byte(parts[0] + "." + parts[1]))
	signature := mac.Sum(nil)

	decodedSignature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		panic(err)
	}

	if !hmac.Equal(signature, decodedSignature) {
		return services.InvalidToken
	}

	decodedToken, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		panic(err)
	}
	ok, err := comparePasswords(token, string(decodedToken))
	if err != nil {
		panic(err)
	}

	if !ok {
		return services.InvalidToken
	}

	decodedClaims, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		panic(err)
	}
	err = json.NewDecoder(bytes.NewBuffer(decodedClaims)).Decode(data)
	if err != nil {
		panic(err)
	}
	return nil
}

func (self *HashedVerifyTokenRepository) Destroy(ctx context.Context, id string) {
	err := self.redisClient.Del(ctx, string(self.prefix)+id).Err()
	if err != nil {
		panic(err)
	}
}

// var _ services.TokenClaimsService = (*HashedVerifyTokenRepository)(nil)
