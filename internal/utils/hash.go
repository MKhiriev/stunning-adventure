package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"hash"

	"github.com/MKhiriev/stunning-adventure/models"
)

type Hasher struct {
	hash.Hash
}

func NewHasher(hashKey string) *Hasher {
	if hashKey != "" {
		return &Hasher{
			Hash: hmac.New(sha256.New, []byte(hashKey)),
		}
	}

	return nil
}

func (h *Hasher) HashByteSlice(slice []byte) ([]byte, error) {
	_, err := h.Write(slice)
	if err != nil {
		return nil, err
	}

	hashedSlice := h.Sum(nil)
	h.Reset()

	return hashedSlice, nil
}

func (h *Hasher) HashMetric(metric models.Metrics) ([]byte, error) {
	metricJSON, err := json.Marshal(metric)
	if err != nil {
		return nil, err
	}

	_, err = h.Write(metricJSON)
	if err != nil {
		return nil, err
	}

	hashedMetric := h.Sum(nil)
	h.Reset()

	return hashedMetric, nil
}
