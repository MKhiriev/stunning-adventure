package utils

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"testing"

	"github.com/MKhiriev/stunning-adventure/models"
)

func TestInitHasherPoolAndHash(t *testing.T) {
	key := "secret-key"
	InitHasherPool(key)

	data := []byte("test-data")

	sum1 := Hash(data)
	sum2 := Hash(data)

	if len(sum1) == 0 {
		t.Fatal("hash result is empty")
	}

	if !bytes.Equal(sum1, sum2) {
		t.Fatal("hash must be deterministic for the same input")
	}

	// verify against direct HMAC computation
	h := hmac.New(sha256.New, []byte(key))
	h.Write(data)
	expected := h.Sum(nil)

	if !bytes.Equal(sum1, expected) {
		t.Fatalf("unexpected hash value\nwant: %x\ngot:  %x", expected, sum1)
	}
}

func TestNewHasher(t *testing.T) {
	h := NewHasher("")
	if h != nil {
		t.Fatal("expected nil hasher for empty key")
	}

	h = NewHasher("key")
	if h == nil {
		t.Fatal("expected hasher instance")
	}
}

func TestHasher_HashMetrics_SingleMetric(t *testing.T) {
	h := NewHasher("metric-key")
	if h == nil {
		t.Fatal("hasher is nil")
	}

	val := float64(10)
	metric := models.Metrics{
		ID:    "cpu",
		MType: models.Gauge,
		Value: &val,
	}

	hash1, err := h.HashMetrics(metric)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hash2, err := h.HashMetrics(metric)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !bytes.Equal(hash1, hash2) {
		t.Fatal("hash must be deterministic for the same metric")
	}

	// validate against manual HMAC
	jsonData, _ := json.Marshal(metric)
	mac := hmac.New(sha256.New, []byte("metric-key"))
	mac.Write(jsonData)
	expected := mac.Sum(nil)

	if !bytes.Equal(hash1, expected) {
		t.Fatalf("unexpected hash\nwant: %x\ngot:  %x", expected, hash1)
	}
}

func TestHasher_HashMetrics_MultipleMetrics(t *testing.T) {
	h := NewHasher("metric-key")

	v1 := float64(1)
	v2 := float64(2)

	metrics := []models.Metrics{
		{
			ID:    "m1",
			MType: models.Gauge,
			Value: &v1,
		},
		{
			ID:    "m2",
			MType: models.Gauge,
			Value: &v2,
		},
	}

	hash1, err := h.HashMetrics(metrics...)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hash2, err := h.HashMetrics(metrics...)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !bytes.Equal(hash1, hash2) {
		t.Fatal("hash must be deterministic for the same metrics slice")
	}

	// ensure slice hashing differs from single metric hashing
	singleHash, err := h.HashMetrics(metrics[0])
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if bytes.Equal(hash1, singleHash) {
		t.Fatal("hash of metrics slice must differ from single metric hash")
	}
}

func TestHasher_HashMetrics_DifferentInputDifferentHash(t *testing.T) {
	h := NewHasher("metric-key")

	v1 := float64(1)
	v2 := float64(2)

	m1 := models.Metrics{
		ID:    "cpu",
		MType: models.Gauge,
		Value: &v1,
	}
	m2 := models.Metrics{
		ID:    "cpu",
		MType: models.Gauge,
		Value: &v2,
	}

	hash1, _ := h.HashMetrics(m1)
	hash2, _ := h.HashMetrics(m2)

	if bytes.Equal(hash1, hash2) {
		t.Fatal("different metric values must produce different hashes")
	}
}
