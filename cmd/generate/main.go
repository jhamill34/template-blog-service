package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <name>")
		os.Exit(1)
	}

	keysDir := os.Getenv("KEYS_DIR")

	name := os.Args[1]

	privateFile, err := os.OpenFile(keysDir+"/"+name+"-key.pem", os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer privateFile.Close()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		panic(err)
	}

	pem.Encode(privateFile, &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	publicFile, err := os.OpenFile(keysDir+"/"+name+".pem", os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer publicFile.Close()

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		panic(err)
	}

	pem.Encode(publicFile, &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	if name == "dkim" {
		dkimRecord, err := os.OpenFile(keysDir+"/dkim-record.txt", os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			panic(err)
		}
		defer dkimRecord.Close()

		publicKeyString := base64.StdEncoding.EncodeToString(publicKeyBytes)

		dkimRecord.WriteString(formatDNSRecord(publicKeyString))
	}
}

func formatDNSRecord(key string) string {
	return fmt.Sprintf("v=DKIM1;k=rsa;h=sha256;p=%s", key)
}
