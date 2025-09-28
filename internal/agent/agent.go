package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/MKhiriev/stunning-adventure/internal/config"
	"github.com/MKhiriev/stunning-adventure/internal/utils"
	"github.com/MKhiriev/stunning-adventure/models"
	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog"
)

type MetricsAgent struct {
	serverAddress  string
	route          string
	client         *resty.Client
	memory         *AgentStorage
	pollCount      int64
	reportInterval int64
	pollInterval   int64
	mu             *sync.Mutex
	logger         *zerolog.Logger
	retryIntervals map[int]time.Duration
	hasher         *utils.Hasher
	rateLimit      int64
}

func NewMetricsAgent(route string, cfg *config.AgentConfig, logger *zerolog.Logger) *MetricsAgent {
	agent := &MetricsAgent{
		serverAddress:  "http://" + cfg.ServerAddress,
		route:          route,
		client:         newHTTPClient(),
		memory:         NewStorage(),
		pollCount:      0,
		reportInterval: cfg.ReportInterval,
		pollInterval:   cfg.PollInterval,
		logger:         logger,
		mu:             &sync.Mutex{},
		retryIntervals: map[int]time.Duration{
			1: 1 * time.Second,
			2: 3 * time.Second,
			3: 5 * time.Second,
		},
		hasher:    utils.NewHasher(cfg.HashKey),
		rateLimit: cfg.RateLimit,
	}

	// add retry mechanism
	agent.client.SetRetryCount(3).
		SetRetryAfter(func(client *resty.Client, response *resty.Response) (time.Duration, error) {
			return agent.retryIntervals[response.Request.Attempt], nil
		}).SetRetryMaxWaitTime(5 * time.Second)

	return agent
}

func (m *MetricsAgent) ReadMetrics() error {
	memStats := runtime.MemStats{}
	runtime.ReadMemStats(&memStats)

	allMetrics := m.getSliceOfMetrics(memStats)
	if len(allMetrics) == 0 {
		m.logger.Error().Caller().Str("func", "*MetricsAgent.ReadMetrics").Msg("error occurred during getting MemStats metrics: no metrics in MemStats")
		return errors.New("error occurred during getting MemStats metrics: no metrics in MemStats")
	}

	m.memory.RefreshAllMetrics(allMetrics...)
	return nil
}

func (m *MetricsAgent) SendBatchMetricsJSON() error {
	// get all metrics from memory
	allMetrics := m.memory.GetAllMetrics()
	if len(allMetrics) == 0 {
		m.logger.Error().Caller().Str("func", "*MetricsAgent.SendBatchMetricsJSON").Msg("no metrics retrieved from memory")
		return errors.New("no metrics passed")
	}

	route, pathJoinError := url.JoinPath(m.serverAddress, m.route, "/")
	if pathJoinError != nil {
		m.logger.Err(pathJoinError).Caller().Str("func", "*MetricsAgent.SendBatchMetricsJSON").Msg("url join error")
		return fmt.Errorf("url join error: %w", pathJoinError)
	}

	// gzip encode metrics
	compressedMetrics, compressionError := gzipCompressMultipleMetrics(allMetrics...)
	if compressionError != nil {
		m.logger.Err(compressionError).Caller().Str("func", "*MetricsAgent.SendBatchMetricsJSON").Msg("error occurred during gzip compression")
		return compressionError
	}

	// send all metrics batched retrieved from memory
	_, sendMetricError := m.client.R().
		SetHeaders(map[string]string{
			"Content-Type":     "application/json",
			"Content-Encoding": "gzip",
		}).
		SetBody(compressedMetrics).
		Post(route)
	if sendMetricError != nil {
		m.logger.Err(sendMetricError).Caller().Str("func", "*MetricsAgent.SendBatchMetricsJSON").Msg("error occurred during sending metric")
		return fmt.Errorf("error occurred during sending metric: %w", sendMetricError)
	}

	m.logger.Info().Caller().Str("func", "*MetricsAgent.SendBatchMetricsJSON").Any("request body", compressedMetrics).Msg("request from agent")

	// after sending metrics set poll count to 0
	m.pollCount = 0

	return nil
}

func (m *MetricsAgent) SendMetricsJSON() error {
	// get all metrics from memory
	allMetrics := m.memory.GetAllMetrics()
	if len(allMetrics) == 0 {
		return errors.New("no metrics passed")
	}

	// send every metric retrieved from memory
	for _, metric := range allMetrics {
		err := m.sendMetric(metric)
		if err != nil {
			return err
		}
	}

	// after sending metrics set poll count to 0
	m.pollCount = 0

	return nil
}

func (m *MetricsAgent) sendMetric(metric models.Metrics) error {
	// construct a route
	route, pathJoinError := url.JoinPath(m.serverAddress, m.route, "/")
	if pathJoinError != nil {
		m.logger.Err(pathJoinError).Caller().Str("func", "*MetricsAgent.sendMetric").Msg("url join error")
		return fmt.Errorf("url join error: %w", pathJoinError)
	}

	// construct headers
	headers := map[string]string{
		"Content-Type":     "application/json",
		"Content-Encoding": "gzip",
	}

	// include hash of the body
	if m.hasher != nil {
		hashedMetric, hashingError := m.hasher.HashMetric(metric)
		if hashingError != nil {
			m.logger.Err(hashingError).Caller().Str("func", "*MetricsAgent.sendMetric").Msg("error occurred during hashing metric")
			return hashingError
		}
		headers["HashSHA256"] = fmt.Sprintf("%x", hashedMetric)
	}

	// gzip encode metric
	compressedMetric, compressionError := gzipCompress(metric)
	if compressionError != nil {
		m.logger.Err(compressionError).Caller().Str("func", "*MetricsAgent.sendMetric").Msg("error occurred during gzip compression")
		return compressionError
	}

	m.logger.Debug().Any("metric", metric).Any("hash", headers).Msg("")

	var response models.Metrics
	_, sendMetricError := m.client.R().
		SetHeaders(headers).
		SetBody(compressedMetric).
		SetResult(&response).
		Post(route)
	if sendMetricError != nil {
		m.logger.Err(sendMetricError).Caller().Str("func", "*MetricsAgent.sendMetric").Msg("error occurred during sending metric")
		return fmt.Errorf("error occurred during sending metric: %w", sendMetricError)
	}

	m.logger.Info().Caller().Str("func", "*MetricsAgent.sendMetric").Any("request", compressedMetric).Any("response", response).Msg("metric is sent!")
	return nil
}

func (m *MetricsAgent) SendMetrics() error {
	// get all metrics from memory
	allMetrics := m.memory.GetAllMetrics()
	if len(allMetrics) == 0 {
		m.logger.Error().Caller().Str("func", "*MetricsAgent.SendMetrics").Msg("no metrics passed")
		return errors.New("no metrics passed")
	}

	// send every metric retrieved from memory
	for _, metric := range allMetrics {
		// construct route based on metric's type
		route, err := m.getRoute(metric)
		if err != nil {
			m.logger.Err(err).Caller().Str("func", "*MetricsAgent.SendMetrics")
			return err
		}

		response, sendMetricError := m.client.R().
			SetHeader("Content-Type", "text/plain").
			Post(route)
		if sendMetricError != nil {
			m.logger.Err(sendMetricError).Caller().Str("func", "*MetricsAgent.SendMetrics").Msg("error occurred during sending metric")
			return fmt.Errorf("error occurred during sending metric: %w", sendMetricError)
		}

		if response.StatusCode() != http.StatusOK {
			m.logger.Error().Caller().Str("func", "*MetricsAgent.SendMetrics").Bool("response.StatusCode == 200", false).Str("response.Status", response.Status()).Msg("error occurred during sending metric")
			return fmt.Errorf("error during metrics sending: %s", response.Status())
		}
	}
	// after sending metrics set poll count to 0
	m.pollCount = 0
	return nil
}

// ReadMetricsGenerator reads metrics and returns a channel that will feed the worker metrics for sending
func (m *MetricsAgent) ReadMetricsGenerator(pollInterval *time.Ticker, reportInterval *time.Ticker) chan models.Metrics {
	metricsChannel := make(chan models.Metrics)

	go func() {
		for {
			select {
			case <-pollInterval.C:
				m.logger.Debug().Str("func", "ReadMetricsGenerator").Msg("time to READ metrics")
				_ = m.ReadMetrics()
			case <-reportInterval.C:
				m.logger.Debug().Str("func", "ReadMetricsGenerator").Msg("time to SEND metrics")
				allMetrics := m.memory.GetAllMetrics()
				for _, metric := range allMetrics {
					m.logger.Debug().Str("func", "ReadMetricsGenerator").Any("metric", metric).Msg("metric is sent to the channel")
					metricsChannel <- metric
				}
			}
		}
	}()

	return metricsChannel
}

// SendMetricsWorker this func we run in separate goroutines
func (m *MetricsAgent) SendMetricsWorker(metrics <-chan models.Metrics) {
	for {
		select {
		case metric := <-metrics:
			m.logger.Debug().Any("metric", metric).Msg("worker is called")
			_ = m.sendMetric(metric)

		}
	}
}

func (m *MetricsAgent) Run() error {
	// reading metrics part
	pollTicker, reportTicker := getTickers(time.Duration(m.pollInterval)*time.Second, time.Duration(m.reportInterval)*time.Second)
	m.logger.Debug().Str("func", "Run").Msg("preparing to run goroutine for reading metrics")
	jobs := m.ReadMetricsGenerator(pollTicker, reportTicker)

	// creating workers
	m.logger.Debug().Str("func", "Run").Msg("creating workers")
	m.withWorkers(func() {
		m.SendMetricsWorker(jobs)
	}, m.rateLimit)
	m.logger.Debug().Str("func", "Run").Msg("workers are created")

	select {} // block main routine forever
}

func newHTTPClient() *resty.Client {
	return resty.New().SetDebug(true)
}

func getTickers(pollIntervalDuration time.Duration, reportIntervalDuration time.Duration) (*time.Ticker, *time.Ticker) {
	return time.NewTicker(pollIntervalDuration), time.NewTicker(reportIntervalDuration)
}

func (m *MetricsAgent) withWorkers(fn func(), count int64) {
	for i := range count {
		m.logger.Debug().Str("func", "withWorkers").Msgf("creating worker #%d", i)
		go fn()
		m.logger.Debug().Msgf("worker#%d is created", i)
	}
}

func (m *MetricsAgent) getRoute(metric models.Metrics) (string, error) {
	if metric.MType == models.Counter {
		// check if Counter's Delta is not nil
		if metric.Delta == nil {
			m.logger.Error().Caller().Str("func", "*MetricsAgent.getRoute").Msg("no metric's data has been passed: field Delta is nil")
			return "", errors.New("no metric's data has been passed: field Delta is nil")
		}

		return fmt.Sprintf("%s/%s/%s/%s/%d", m.serverAddress, m.route,
			metric.MType, metric.ID, *metric.Delta), nil
	}

	if metric.MType == models.Gauge {
		// check if Gauge's Value is not nil
		if metric.Value == nil {
			m.logger.Error().Caller().Str("func", "*MetricsAgent.getRoute").Msg("no metric's data has been passed: field Value in nil")
			return "", errors.New("no metric's data has been passed: field Value in nil")
		}

		return fmt.Sprintf("%s/%s/%s/%s/%s", m.serverAddress, m.route,
			metric.MType, metric.ID, strconv.FormatFloat(*metric.Value, 'f', -1, 64)), nil
	}

	return "", errors.New("error occurred during route construction")
}

func gzipCompress(metric models.Metrics) ([]byte, error) {
	// сериализуем metric в JSON
	jsonData, err := json.Marshal(metric)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metric: %w", err)
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

func gzipCompressMultipleMetrics(metrics ...models.Metrics) ([]byte, error) {
	var jsonData []byte
	var err error
	// сериализуем metrics в JSON
	if len(metrics) == 1 {
		jsonData, err = json.Marshal(metrics[0])
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metric: %w", err)
		}
	} else {
		jsonData, err = json.Marshal(metrics)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metric: %w", err)
		}
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
