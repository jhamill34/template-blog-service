package routes

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jhamill34/notion-provisioner/internal/models"
	"github.com/jhamill34/notion-provisioner/internal/transport/utils"
)

type KeyRoutes struct {
	publicKey *rsa.PublicKey
}

func NewKeyRoutes(publicKey *rsa.PublicKey) *KeyRoutes {
	return &KeyRoutes{publicKey}
}

// Routes implements transport.Router.
func (self *KeyRoutes) Routes() (string, http.Handler) {
	router := chi.NewRouter()

	router.Get("/signer", self.GetSigningKey())

	return "/key", router
}

func (self *KeyRoutes) GetSigningKey() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bytes := x509.MarshalPKCS1PublicKey(self.publicKey)
		keyString := base64.StdEncoding.EncodeToString(bytes)

		publicKey := models.PublicKeyResponse{PublicKey: keyString}
		utils.RenderJSON(w, publicKey, http.StatusOK)
	}
}
