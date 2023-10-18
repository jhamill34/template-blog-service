package app

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io"
	"os"

	"github.com/jhamill34/notion-provisioner/internal/config"
	e "github.com/jhamill34/notion-provisioner/internal/services/email"
	"github.com/jhamill34/notion-provisioner/internal/transport/email"
	"github.com/jhamill34/notion-provisioner/internal/transport/email/smtp"
)

type MailService struct {
	server *email.EmailServer
}

func ConfigureMail() *MailService {
	cfg, err := config.LoadMailConfig("configs/mail.yaml")
	if err != nil {
		panic(err)
	}

	return &MailService{
		email.NewEmailServer(
			smtp.SmtpProtocolRouter(
				e.NewEmailForwarder(
					cfg.Forwarder.CommonPorts,
					e.SigningOptions{
						Selector:   cfg.Dkim.Selector,
						Domain:     cfg.Dkim.Domain,
						Headers:    cfg.Dkim.Headers,
						PrivateKey: loadDKIMKey(cfg.Dkim.PrivateKeyPath),
					},
				),
			),
			nil,
			cfg.Protocol,
			cfg.Port,
		),
	}
}

func (m *MailService) Start(ctx context.Context) {
	m.server.Start(ctx)
}

func loadDKIMKey(path string) *rsa.PrivateKey {
	privateFile, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer privateFile.Close()

	privateKeyBytes, err := io.ReadAll(privateFile)
	if err != nil {
		panic(err)
	}

	block, _ := pem.Decode(privateKeyBytes)
	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		panic(err)
	}

	if privateKey, ok := privateKey.(*rsa.PrivateKey); ok {
		return privateKey
	}

	panic("Not an RSA private key")
}
