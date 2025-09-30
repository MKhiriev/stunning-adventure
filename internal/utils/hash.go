package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"

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

func (h *Hasher) HashMetrics(metrics ...models.Metrics) ([]byte, error) {
	var metricJSON []byte
	var err error

	if len(metrics) == 1 {
		metricJSON, err = json.Marshal(metrics[0])
		if err != nil {
			return nil, fmt.Errorf("error during json marshalling metric: %w", err)
		}
	} else {
		metricJSON, err = json.Marshal(metrics)
		if err != nil {
			return nil, fmt.Errorf("error during json marshalling metrics: %w", err)
		}
	}

	hasher := hmac.New(sha256.New, h.hashKey)
	_, err = hasher.Write(metricJSON)
	if err != nil {
		return nil, fmt.Errorf("error during hashing metric(s): %w", err)
	}

	hashedMetric := hasher.Sum(nil)

	return hashedMetric, nil
}

func Hash(data []byte, hashKey string) []byte {
	hasher := hmac.New(sha256.New, []byte(hashKey))
	hasher.Write(data)
	return hasher.Sum(nil)
}
