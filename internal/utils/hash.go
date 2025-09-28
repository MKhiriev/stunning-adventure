package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"

	"github.com/MKhiriev/stunning-adventure/models"
)

type Hasher struct {
	hashKey []byte
}

func NewHasher(hashKey string) *Hasher {
	if hashKey != "" {
		return &Hasher{
			hashKey: []byte(hashKey),
		}
	}

	return nil
}

func (h *Hasher) HashMetric(metric models.Metrics) ([]byte, error) {
	metricJSON, err := json.Marshal(metric)
	if err != nil {
		return nil, err
	}

	hasher := hmac.New(sha256.New, h.hashKey)
	_, err = hasher.Write(metricJSON)
	if err != nil {
		return nil, err
	}

	hashedMetric := hasher.Sum(nil)

	return hashedMetric, nil
}

func Hash(data []byte, hashKey string) []byte {
	hasher := hmac.New(sha256.New, []byte(hashKey))
	hasher.Write(data)
	return hasher.Sum(nil)
}
