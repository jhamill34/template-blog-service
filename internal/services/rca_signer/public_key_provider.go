package rca_signer

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jhamill34/notion-provisioner/internal/models"
)

type StaticPublicKeyProvider struct {
	publicKey *rsa.PublicKey
}

func NewStaticPublicKeyProvider(publicKey *rsa.PublicKey) *StaticPublicKeyProvider {
	return &StaticPublicKeyProvider{publicKey}
}

func (self *StaticPublicKeyProvider) GetKey() *rsa.PublicKey {
	return self.publicKey
}

//==================================================

type RemotePublicKeyProvider struct {
	publicKeyUrl string
	publicKey    *rsa.PublicKey
	nextFetch    int64
}

func NewRemotePublicKeyProvider(publicKeyUrl string) *RemotePublicKeyProvider {
	return &RemotePublicKeyProvider{publicKeyUrl, nil, 0}
}

func (self *RemotePublicKeyProvider) GetKey() *rsa.PublicKey {
	if self.publicKey == nil || time.Now().Unix() > self.nextFetch {
		self.fetchPublicKey()
	}

	return self.publicKey
}

func (self *RemotePublicKeyProvider) fetchPublicKey() {
	var publicKeyResponse models.PublicKeyResponse
	resp, err := http.Get(self.publicKeyUrl)
	if err != nil {
		panic(err)
	}
	json.NewDecoder(resp.Body).Decode(&publicKeyResponse)

	keyBytes, err := base64.StdEncoding.DecodeString(publicKeyResponse.PublicKey)
	if err != nil {
		panic(err)
	}
	publicKey, err := x509.ParsePKCS1PublicKey(keyBytes)
	if err != nil {
		panic(err)
	}

	self.publicKey = publicKey
	self.nextFetch = time.Now().Unix() + publicKeyResponse.TTL
}
