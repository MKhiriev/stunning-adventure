package utils

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func EncryptData(data []byte, publicKey *rsa.PublicKey) (*bytes.Buffer, error) {
	chunkSize := publicKey.Size() - 11                     // size of bytes that can be encrypted in one swoop
	chunksCount := (len(data) + chunkSize - 1) / chunkSize // count of how many slices of data need to be encrypted
	encryptedLen := publicKey.Size() * chunksCount         // exact size of encrypted data

	result := bytes.NewBuffer([]byte{})
	result.Grow(encryptedLen) // set fixed size to prevent reallocations

	// iterating over chunks
	for currentChunk := range chunksCount {
		chunkStartIndex := currentChunk * chunkSize // start position
		var chunkEndIndex int                       // end position (not included)
		if currentChunk == chunksCount-1 {
			chunkEndIndex = len(data)
		} else {
			chunkEndIndex = chunkStartIndex + chunkSize
		}

		if err := encryptData(data[chunkStartIndex:chunkEndIndex], publicKey, result); err != nil {
			return nil, err
		}
	}

	return result, nil
}

func DecryptData(encryptedData []byte, privateKey *rsa.PrivateKey) (*bytes.Buffer, error) {
	chunkSize := privateKey.Size()                                            // chunk size of encrypted data
	chunksCount := len(encryptedData) / chunkSize                             // encrypted parts count
	decryptedLen := (privateKey.Size() - 11) * len(encryptedData) / chunkSize // decrypted message length

	result := bytes.NewBuffer([]byte{})
	result.Grow(decryptedLen) // set fixed size to prevent reallocations

	for currentChunk := range chunksCount {
		currentIndex := currentChunk * chunkSize
		endIndex := currentIndex + chunkSize

		if err := decryptData(encryptedData[currentIndex:endIndex], privateKey, result); err != nil {
			return nil, err
		}

	}

	return result, nil
}

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

func decryptData(encryptedData []byte, privateKey *rsa.PrivateKey, result *bytes.Buffer) error {
	decryptedPart, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, encryptedData)
	if err != nil {
		return err
	}

	_, err = result.Write(decryptedPart)
	if err != nil {
		return err
	}

	return nil
}

func encryptData(data []byte, publicKey *rsa.PublicKey, result *bytes.Buffer) error {
	encryptedMessage, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, data) // encrypt portion of data
	if err != nil {
		return fmt.Errorf("error encrypting data portion: %w", err)
	}

	_, err = result.Write(encryptedMessage) // write encrypted portion of data to the buffer
	if err != nil {
		return fmt.Errorf("error writing encrypted data portion to a buffer: %w", err)
	}
	return nil
}
