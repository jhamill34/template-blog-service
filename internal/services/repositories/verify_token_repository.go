package repositories

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jhamill34/notion-provisioner/internal/config"
	"github.com/redis/go-redis/v9"
)

type VerificationType string

const (
	VerificationTypeRegistration   VerificationType = "registration:"
	VerificationTypeForgotPassword VerificationType = "forgot_password:"
	VerificationTypeInvite         VerificationType = "invite:"
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
) (string, error) {
	publicToken := uuid.New().String()
	token, err := createHash(self.passwordParams, publicToken)

	err = self.redisClient.Set(ctx, string(self.prefix)+id, token, self.ttl).Err()
	if err != nil {
		return "", err
	}

	return publicToken, nil
}

// Verify implements services.VerifyTokenService.
func (self *HashedVerifyTokenRepository) Verify(
	ctx context.Context,
	id, token string,
) error {
	hashedToken, err := self.redisClient.Get(ctx, string(self.prefix)+id).Result()
	if err != nil {
		return err
	}

	ok, err := comparePasswords(token, hashedToken)
	if err != nil {
		return err
	}

	if ok {
		err = self.redisClient.Del(ctx, string(self.prefix)+id).Err()
		if err != nil {
			return err
		}

		return nil
	}

	return fmt.Errorf("Invalid Token")
}

// CreateWithClaims implements services.TokenClaimsService.
func (self *HashedVerifyTokenRepository) CreateWithClaims(
	ctx context.Context,
	id string,
	data interface{},
) (string, error) {
	publicToken := uuid.New().String()

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
		return "", err
	}

	return publicToken, nil
}

// VerifyWithClaims implements services.TokenClaimsService.
func (self *HashedVerifyTokenRepository) VerifyWithClaims(
	ctx context.Context,
	id string,
	token string,
	data interface{},
) error {
	hashedToken, err := self.redisClient.Get(ctx, string(self.prefix)+id).Result()
	if err != nil {
		return err
	}

	parts := strings.Split(hashedToken, ".")
	if len(parts) != 3 {
		return fmt.Errorf("Invalid Token")
	}

	mac := hmac.New(sha256.New, []byte(self.passwordParams.Secret))
	mac.Write([]byte(parts[0] + "." + parts[1]))
	signature := mac.Sum(nil)

	decodedSignature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return err
	}

	if !hmac.Equal(signature, decodedSignature) {
		return fmt.Errorf("Invalid Token")
	}

	decodedToken, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return err
	}
	ok, err := comparePasswords(token, string(decodedToken))
	if err != nil {
		return err
	}

	if !ok {
		return fmt.Errorf("Invalid Token")
	}

	decodedClaims, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return err
	}
	err = json.NewDecoder(bytes.NewBuffer(decodedClaims)).Decode(data)
	if err != nil {
		return err
	}
	return nil 
}

func (self *HashedVerifyTokenRepository) Destroy(ctx context.Context, id string) error {
	err := self.redisClient.Del(ctx, string(self.prefix)+id).Err()
	if err != nil {
		return err
	}

	return nil
}

// var _ services.TokenClaimsService = (*HashedVerifyTokenRepository)(nil)
