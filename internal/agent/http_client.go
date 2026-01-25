package agent

import (
	"bytes"
	"compress/gzip"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/MKhiriev/stunning-adventure/internal/utils"
	"github.com/MKhiriev/stunning-adventure/models"
	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog"
)

type HTTPMetricsSender struct {
	route  string        // server route for sending metrics
	client *resty.Client // HTTP client for sending metrics

	hasher    *utils.Hasher
	publicKey *rsa.PublicKey
	logger    *zerolog.Logger

	realIP string
}

func NewHTTPMetricsSender(serverAddress, route string, retryIntervals map[int]time.Duration, hasher *utils.Hasher, publicKey *rsa.PublicKey, logger *zerolog.Logger) (*HTTPMetricsSender, error) {
	httpClient := utils.NewHTTPClient(5 * time.Second)
	httpClient.SetRetryCount(3).
		SetRetryAfter(func(client *resty.Client, response *resty.Response) (time.Duration, error) {
			return retryIntervals[response.Request.Attempt], nil
		}).SetRetryMaxWaitTime(5 * time.Second)

	realIp, err := utils.GetLocalIP(serverAddress)
	if err != nil {
		return nil, fmt.Errorf("unable to get real agent IP: %w", err)
	}

	return &HTTPMetricsSender{
		route:     route,
		client:    httpClient,
		hasher:    hasher,
		publicKey: publicKey,
		logger:    logger,
		realIP:    realIp.String(),
	}, nil
}

func (m *HTTPMetricsSender) Send(metrics ...models.Metrics) error {
	return m.sendMetrics(metrics...)
}

func (m *HTTPMetricsSender) Close() error {
	m.logger.Info().Str("func", "*HTTPMetricsSender.Close").Msg("closing HTTP client...")
	return nil
}

func (m *HTTPMetricsSender) sendMetrics(metrics ...models.Metrics) error {
	if len(metrics) == 0 {
		m.logger.Error().Str("func", "*HTTPMetricsSender.sendMetrics").Msg("no metric was passed!")
		return errors.New("no metric was passed")
	}

	// construct headers
	headers := map[string]string{
		"Content-Type":     "application/json",
		"Content-Encoding": "gzip",
		"X-Real-IP":        m.realIP,
	}

	// include hash of the body
	if m.hasher != nil {
		hashedMetric, hashingError := m.hasher.HashMetrics(metrics...)
		if hashingError != nil {
			m.logger.Err(hashingError).Caller().Str("func", "*HTTPMetricsSender.sendMetrics").Msg("error occurred during hashing metric")
			return hashingError
		}
		headers["HashSHA256"] = fmt.Sprintf("%x", hashedMetric)
	}

	// gzip encode metric
	compressedMetric, compressionError := gzipCompress(metrics...)
	if compressionError != nil {
		m.logger.Err(compressionError).Caller().Str("func", "*HTTPMetricsSender.sendMetrics").Msg("error occurred during gzip compression")
		return compressionError
	}

	if m.publicKey != nil {
		encryptedMessage, err := utils.EncryptData(compressedMetric, m.publicKey)
		if err != nil {
			m.logger.Err(err).Str("func", "*HTTPMetricsSender.sendMetrics").Msg("error encrypting data")
			return err
		}
		compressedMetric = encryptedMessage.Bytes()
	}

	m.logger.Debug().Any("metric", metrics).Any("headers", headers).Send()

	var response models.Metrics
	_, sendMetricError := m.client.R().
		SetHeaders(headers).
		SetBody(compressedMetric).
		SetResult(&response).
		Post(m.route)

	if sendMetricError != nil {
		m.logger.Err(sendMetricError).Caller().Str("func", "*MetricsAgent.sendMetrics").Msg("error occurred during sending metric")
		return fmt.Errorf("error occurred during sending metric: %w", sendMetricError)
	}

	m.logger.Info().Caller().Str("func", "*MetricsAgent.sendMetrics").Any("request", compressedMetric).Any("response", response).Msg("metric is sent!")
	return nil
}

func gzipCompress(metrics ...models.Metrics) ([]byte, error) {
	if len(metrics) == 0 {
		return nil, errors.New("no metrics provided")
	}

	jsonData, err := json.Marshal(metrics)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metric(s): %w", err)
	}

	// создаем gzip-сжатие
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(jsonData); err != nil {
		return nil, fmt.Errorf("failed to gzip compress: %w", err)
	}
	if err := gz.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}

	return buf.Bytes(), nil
}
