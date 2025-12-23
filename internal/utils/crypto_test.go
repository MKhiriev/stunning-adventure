package utils

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	key := generateRSAKey(t)

	data := bytes.Repeat([]byte("hello-world-"), 1000) // гарантированно multi-chunk

	encrypted, err := EncryptData(data, &key.PublicKey)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	decrypted, err := DecryptData(encrypted.Bytes(), key)
	if err != nil {
		t.Fatalf("decrypt failed: %v", err)
	}

	if !bytes.Equal(decrypted.Bytes(), data) {
		t.Fatal("decrypted data does not match original")
	}
}

func TestEncryptDecrypt_SingleChunk(t *testing.T) {
	key := generateRSAKey(t)

	data := []byte("short message")

	enc, err := EncryptData(data, &key.PublicKey)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	dec, err := DecryptData(enc.Bytes(), key)
	if err != nil {
		t.Fatalf("decrypt failed: %v", err)
	}

	if !bytes.Equal(dec.Bytes(), data) {
		t.Fatal("round-trip failed")
	}
}

func TestGetPrivateKey_Success(t *testing.T) {
	key := generateRSAKey(t)
	path := writePrivateKeyPEM(t, key)

	loaded, err := GetPrivateKey(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if loaded.N.Cmp(key.N) != 0 {
		t.Fatal("loaded private key mismatch")
	}
}

func TestGetPrivateKey_EmptyPath(t *testing.T) {
	if _, err := GetPrivateKey(""); err == nil {
		t.Fatal("expected error")
	}
}

func TestGetPrivateKey_InvalidPEM(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "bad.pem")
	os.WriteFile(path, []byte("not pem"), 0644)

	if _, err := GetPrivateKey(path); err == nil {
		t.Fatal("expected error")
	}
}

func TestGetPublicKey_Success(t *testing.T) {
	key := generateRSAKey(t)
	path := writePublicCertPEM(t, key)

	pub, err := GetPublicKey(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pub.N.Cmp(key.PublicKey.N) != 0 {
		t.Fatal("public key mismatch")
	}
}

func TestGetPublicKey_EmptyPath(t *testing.T) {
	if _, err := GetPublicKey(""); err == nil {
		t.Fatal("expected error")
	}
}

func TestGetPublicKey_InvalidPEM(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "bad.pem")
	os.WriteFile(path, []byte("not pem"), 0644)

	if _, err := GetPublicKey(path); err == nil {
		t.Fatal("expected error")
	}
}

/* ############################ HELPER METHODS ############################ */

func generateRSAKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}
	return key
}

func writePrivateKeyPEM(t *testing.T, key *rsa.PrivateKey) string {
	t.Helper()

	tmp := t.TempDir()
	path := filepath.Join(tmp, "key.pem")

	bytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})

	if err := os.WriteFile(path, bytes, 0600); err != nil {
		t.Fatalf("failed to write private key: %v", err)
	}

	return path
}

func writePublicCertPEM(t *testing.T, key *rsa.PrivateKey) string {
	t.Helper()

	tmp := t.TempDir()
	path := filepath.Join(tmp, "cert.pem")

	template := &x509.Certificate{
		SerialNumber: bigIntOne(),
		NotBefore:    now(),
		NotAfter:     now().AddDate(1, 0, 0),
	}

	certDER, err := x509.CreateCertificate(
		rand.Reader,
		template,
		template,
		&key.PublicKey,
		key,
	)
	if err != nil {
		t.Fatalf("failed to create certificate: %v", err)
	}

	pemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	if err := os.WriteFile(path, pemBytes, 0644); err != nil {
		t.Fatalf("failed to write cert: %v", err)
	}

	return path
}

func bigIntOne() *big.Int {
	return big.NewInt(1)
}

func now() time.Time {
	return time.Now()
}
