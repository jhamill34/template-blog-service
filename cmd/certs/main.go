package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jhamill34/notion-provisioner/internal/config"
	"golang.org/x/crypto/acme"
	"gopkg.in/yaml.v3"
)

type SslConfig struct {
	AccountKeyPath config.StringFromEnv   `yaml:"account_key_path"`
	CertPath       config.StringFromEnv   `yaml:"cert_path"`
	KeyPath        config.StringFromEnv   `yaml:"key_path"`
	DnsDomains     []config.StringFromEnv `yaml:"dns_domains"`
}

type OwnerConfig struct {
	Email         config.StringFromEnv `yaml:"email"`
	Country       config.StringFromEnv `yaml:"country"`
	Organization  config.StringFromEnv `yaml:"organization"`
	Locality      config.StringFromEnv `yaml:"locality"`
	Province      config.StringFromEnv `yaml:"province"`
	StreetAddress config.StringFromEnv `yaml:"street_address"`
	PostalCode    config.StringFromEnv `yaml:"postal_code"`
}

func main() {
	inputScanner := bufio.NewScanner(os.Stdin)

	ownerInfoPath := os.Getenv("OWNER_INFO_FILE")
	ownerInfoFile, err := os.Open(ownerInfoPath)
	if err != nil {
		panic(err)
	}
	var owner OwnerConfig
	err = yaml.NewDecoder(ownerInfoFile).Decode(&owner)
	if err != nil {
		panic(err)
	}
	ownerInfoFile.Close()

	configPath := os.Getenv("CONFIG_FILE")
	inputConfigFile, err := os.Open(configPath)
	if err != nil {
		panic(err)
	}

	var config SslConfig
	err = yaml.NewDecoder(inputConfigFile).Decode(&config)
	if err != nil {
		panic(err)
	}
	inputConfigFile.Close()
	domainNames := make([]string, len(config.DnsDomains))
	for i, domain := range config.DnsDomains {
		domainNames[i] = domain.String()
	}

	var accountKey *rsa.PrivateKey
	if _, err := os.Stat(config.AccountKeyPath.String()); err == nil {
		accountFile, err := os.Open(config.AccountKeyPath.String())
		if err != nil {
			panic(err)
		}
		defer accountFile.Close()

		accountBytes, err := io.ReadAll(accountFile)
		if err != nil {
			panic(err)
		}

		block, _ := pem.Decode(accountBytes)
		maybeAccountKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)

		var ok bool
		accountKey, ok = maybeAccountKey.(*rsa.PrivateKey)
		if !ok {
			panic("Could not parse account key")
		}

		if err != nil {
			panic(err)
		}
	} else {
		panic(err)
	}

	var url string
	if os.Getenv("ACME_STAGING") == "true" {
		log.Printf("Using staging environment...\n")
		url = "https://acme-staging-v02.api.letsencrypt.org/directory"
	} else {
		url = acme.LetsEncryptURL
	}

	ctx := context.TODO()
	client := &acme.Client{
		Key:          accountKey,
		DirectoryURL: url,
	}

	_, err = client.GetReg(ctx, "ignored")
	if err == acme.ErrNoAccount {
		log.Printf("No account found on letsencrypt, trying to register...\n")
		account := &acme.Account{
			Contact: []string{"mailto:" + owner.Email.String()},
		}
		accountResponse, err := client.Register(ctx, account, func(tosUrl string) bool {
			log.Printf("Please accept the terms of service: %v\n", tosUrl)
			log.Printf("Enter 'yes' when you have accepted the terms of service.\n")
			inputScanner.Scan()
			return inputScanner.Text() == "yes"
		})

		if err != nil {
			panic(err)
		}
		log.Printf("Registered account: %v\n", accountResponse)
	}
	if err != nil && acme.ErrNoAccount != err {
		panic(err)
	}

	ids := []acme.AuthzID{
		{
			Type:  "dns",
			Value: domainNames[0],
		},
	}
	log.Printf("Authorizing order for domains: %v\n", ids)

	order, err := client.AuthorizeOrder(ctx, ids)
	if err != nil {
		panic(err)
	}

	for _, authzUrl := range order.AuthzURLs {
		authz, err := client.GetAuthorization(ctx, authzUrl)
		if err != nil {
			panic(err)
		}

		domain := authz.Identifier.Value
		if authz.Wildcard {
			domain = strings.TrimPrefix(domain, "*.")
		}

		for _, challenge := range authz.Challenges {
			switch challenge.Type {
			case "dns-01":
				handleDns01Challenge(ctx, client, inputScanner, domain, challenge)
			default:
				log.Printf("Unsupported challenge type: %v\n", challenge.Type)
			}
		}

		authz, err = client.WaitAuthorization(ctx, authz.URI)
		log.Printf("Authorization status: %v\n", authz.Status)
	}

	order, err = client.WaitOrder(ctx, order.URI)
	log.Printf("Order status: %v\n", order.Status)

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)

	template := pkix.Name{
		Country:       []string{owner.Country.String()},
		Organization:  []string{owner.Organization.String()},
		Locality:      []string{owner.Locality.String()},
		Province:      []string{owner.Province.String()},
		StreetAddress: []string{owner.StreetAddress.String()},
		PostalCode:    []string{owner.PostalCode.String()},
		CommonName:    domainNames[0],
	}
	csr, err := x509.CreateCertificateRequest(rand.Reader, &x509.CertificateRequest{
		SignatureAlgorithm: x509.SHA256WithRSA,
		Subject:            template,
		DNSNames:           domainNames,
	}, privateKey)
	if err != nil {
		panic(err)
	}

	ders, _, err := client.CreateOrderCert(ctx, order.FinalizeURL, csr, false)
	if err != nil {
		panic(err)
	}

	certFile, err := os.OpenFile(config.CertPath.String(), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	defer certFile.Close()

	for _, der := range ders {
		pem.Encode(certFile, &pem.Block{
			Type:    "CERTIFICATE",
			Headers: nil,
			Bytes:   der,
		})
	}

	privateKeyFile, err := os.OpenFile(config.KeyPath.String(), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	defer privateKeyFile.Close()
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		panic(err)
	}

	pem.Encode(privateKeyFile, &pem.Block{
		Type:    "PRIVATE KEY",
		Headers: nil,
		Bytes:   privateKeyBytes,
	})
}

func handleDns01Challenge(
	ctx context.Context,
	client *acme.Client,
	scanner *bufio.Scanner,
	domain string,
	challenge *acme.Challenge,
) {
	if challenge.Status == "pending" {
		record, err := client.DNS01ChallengeRecord(challenge.Token)
		if err != nil {
			panic(err)
		}

		log.Printf(
			"Please add the following TXT record to your DNS:\n    _acme-challenge.%s TXT %v\n",
			domain,
			record,
		)
		log.Printf("Press enter when you have added the record.")
		scanner.Scan()

		log.Printf("Validating challenge...")
		challenge, err = client.Accept(ctx, challenge)
		if err != nil {
			panic(err)
		}
	}

	attempts := 10
	for i := 0; i < attempts; i++ {
		challenge, err := client.GetChallenge(ctx, challenge.URI)
		if err != nil {
			panic(err)
		}

		switch challenge.Status {
		case "valid":
			log.Printf("Challenge valid!")
			return
		case "invalid":
			log.Printf("Challenge: %v\n", challenge)
			log.Fatal("Challenge invalid, exiting...")
			return
		default:
			// Pending status
			log.Printf("Challenge status: %v\n", challenge.Status)
		}

		log.Printf("Challenge not yet valid, waiting 10 seconds...")
		time.Sleep(10 * time.Second)
	}
}
