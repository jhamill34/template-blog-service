package rca_signer

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
)

type RcaSigner struct {
	privateKey *rsa.PrivateKey
}

func NewRcaSigner(privateKey *rsa.PrivateKey) *RcaSigner {
	return &RcaSigner{privateKey}
}

// Sign implements services.Signer.
func (self *RcaSigner) Sign(data []byte) (string, error) {
	dataHash := sha256.New()
	if _, err := dataHash.Write(data); err != nil {
		return "", err
	}

	signature, err := rsa.SignPSS(rand.Reader, self.privateKey, crypto.SHA256, dataHash.Sum(nil), nil)
	if err != nil {
		return "", err
	}

	encodedSignature := base64.RawURLEncoding.EncodeToString(signature)

	return encodedSignature, nil
}

// Verify implements services.Signer.
func (self *RcaSigner) Verify(data []byte, signature string) error {
	decodedSignature, err := base64.RawURLEncoding.DecodeString(signature)
	if err != nil {
		return err
	}

	dataHash := sha256.New()
	if _, err := dataHash.Write(data); err != nil {
		return err 
	}

	return rsa.VerifyPSS(&self.privateKey.PublicKey, crypto.SHA256, dataHash.Sum(nil), decodedSignature, nil)
}

// var _ services.Signer = (*RcaSigner)(nil)
