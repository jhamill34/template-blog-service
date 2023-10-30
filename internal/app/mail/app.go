package mail

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"io"
	"os"
	"strings"

	"github.com/jhamill34/notion-provisioner/internal/config"
	e "github.com/jhamill34/notion-provisioner/internal/services/email"
	"github.com/jhamill34/notion-provisioner/internal/transport/email"
	"github.com/jhamill34/notion-provisioner/internal/transport/email/smtp"
)

type MailService struct {
	server *email.EmailServer
}

func Configure() *MailService {
	cfg, err := config.LoadMailConfig(os.Getenv("CONFIG_FILE"))
	if err != nil {
		panic(err)
	}
	cert, err := tls.LoadX509KeyPair(cfg.TLS.CertificatePath.String(), cfg.TLS.KeyPath.String())
	if err != nil {
		panic(err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	credentials := strings.Split(cfg.AuthCredentials.String(), ":")
	if len(credentials) != 2 {
		panic("Invalid auth credentials in configuration")
	}

	return &MailService{
		email.NewEmailServer(
			smtp.SmtpProtocolRouter(
				cfg.Dkim.Domain.String(),
				e.NewEmailForwarder(
					cfg.Forwarder.CommonPorts,
					e.SigningOptions{
						Selector:   cfg.Dkim.Selector,
						Domain:     cfg.Dkim.Domain.String(),
						Headers:    cfg.Dkim.Headers,
						PrivateKey: loadDKIMKey(cfg.Dkim.PrivateKeyPath.String()),
					},
				),
				e.NewAuthFromConfigHandler(credentials[0], credentials[1]),
			),
			tlsConfig,
			cfg,
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
