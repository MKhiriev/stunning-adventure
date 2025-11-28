package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"hash"
	"sync"

	"github.com/MKhiriev/stunning-adventure/models"
)

var hasherPool sync.Pool

// Hasher provides keyed HMAC-SHA256 hashing for metrics.
// It stores the hash key as a byte slice to avoid repeated conversions.
type Hasher struct {
	hashKey []byte
}

// InitHasherPool initializes a sync.Pool of HMAC-SHA256 hashers.
// Each hasher in the pool is configured with the provided hash key.
//
// Purpose:
//   - Avoid repeated allocations of new hash.Hash instances
//   - Reduce GC pressure in high-throughput hashing paths
//
// Parameters:
//
//	hashKey - key used for all HMAC operations
func InitHasherPool(hashKey string) {
	hasherPool = sync.Pool{
		New: func() any {
			return hmac.New(sha256.New, []byte(hashKey))
		},
	}
}

// Hash computes an HMAC-SHA256 signature over the given byte slice
// using a hasher pulled from the global hasher pool.
//
// Behavior:
//   - Retrieves a hash.Hash instance from sync.Pool
//   - Resets it, writes the data, computes the sum
//   - Resets again and returns it to the pool
//
// Parameters:
//
//	data - arbitrary byte slice to be hashed
//
// Returns:
//
//	[]byte - HMAC-SHA256 digest
func Hash(data []byte) []byte {
	h := hasherPool.Get().(hash.Hash)
	h.Reset()

	h.Write(data)
	sum := h.Sum(nil)

	h.Reset()
	hasherPool.Put(h)

	return sum
}

// NewHasher constructs a Hasher for manual metric hashing workflows.
// Returns nil if the hashKey is an empty string.
//
// Parameters:
//
//	hashKey - HMAC key used for hashing operations
//
// Returns:
//
//	*Hasher - configured hasher instance or nil
func NewHasher(hashKey string) *Hasher {
	if hashKey != "" {
		return &Hasher{
			hashKey: []byte(hashKey),
		}
	}

	return nil
}

// HashMetrics computes an HMAC-SHA256 digest for one or more Metrics objects.
//
// Behavior:
//   - Marshals either a single metric (metrics[0]) or the full slice (metrics)
//   - Constructs a fresh HMAC instance with the stored key
//   - Writes the JSON representation into the HMAC
//   - Returns the resulting digest
//
// Parameters:
//
//	metrics - one or more models.Metrics values to hash
//
// Returns:
//
//	[]byte - HMAC digest
//	error  - marshalling or hashing error
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
