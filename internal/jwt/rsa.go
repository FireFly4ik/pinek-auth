package jwt

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

func LoadRSAPrivateKey() (*rsa.PrivateKey, error) {
	data, err := os.ReadFile("rsa_private.pem")
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to parse PEM block from private key file")
	}

	var parsedKey any
	parsedKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		parsedKeyAny, err2 := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err2 != nil {
			return nil, fmt.Errorf("failed to parse RSA private key: %v / %v", err, err2)
		}
		parsedKey = parsedKeyAny
	}

	rsaKey, ok := parsedKey.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("key is not RSA private key")
	}

	return rsaKey, nil
}

func LoadRSAPublicKey() (*rsa.PublicKey, error) {
	data, err := os.ReadFile("rsa_public.pem")
	if err != nil {
		return nil, fmt.Errorf("failed to read public key file: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to parse PEM block from public key file")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSA public key: %w", err)
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("key is not RSA public key")
	}

	return rsaPub, nil
}
