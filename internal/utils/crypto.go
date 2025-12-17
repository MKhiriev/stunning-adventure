package utils

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func GetPrivateKey(fPath string) (*rsa.PrivateKey, error) {
	if fPath == "" {
		return nil, errors.New("empty private key file path was given")
	}

	privateKeyBytes, err := os.ReadFile(fPath)
	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %w", filepath.Base(fPath), err)
	}

	privateKeyPemBlock, _ := pem.Decode(privateKeyBytes)
	if privateKeyPemBlock == nil {
		return nil, errors.New("private key not found")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyPemBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("error parsing private key: %w", err)
	}

	return privateKey, nil
}

func GetPublicKey(fPath string) (*rsa.PublicKey, error) {
	if fPath == "" {
		return nil, errors.New("empty certificate (public key) file path was given")
	}

	certificateBytes, err := os.ReadFile(fPath)
	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %w", filepath.Base(fPath), err)
	}

	certificatePemBlock, _ := pem.Decode(certificateBytes)
	if certificatePemBlock == nil {
		return nil, errors.New("certificate (public key) key not found")
	}

	certificate, err := x509.ParseCertificate(certificatePemBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("error parsing certificate (public key): %w", err)
	}

	publicKey, ok := certificate.PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("error getting public key from certificate: %w", err)
	}

	return publicKey, nil
}
