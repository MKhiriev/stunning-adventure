package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os/signal"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/MKhiriev/stunning-adventure/internal/config"
	"github.com/MKhiriev/stunning-adventure/internal/utils"
	"github.com/MKhiriev/stunning-adventure/models"
	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog"
)

// MetricsAgent provides a concrete implementation of a local metrics collection agent.
// It reads metrics from the host, caches them in memory, and sends them to a remote server.
type MetricsAgent struct {
	serverAddress  string        // full server address including http://
	route          string        // server route for sending metrics
	client         *resty.Client // HTTP client for sending metrics
	memory         *AgentStorage // in-memory storage for metrics
	pollCount      int64         // number of polls since last report
	reportInterval int64         // seconds between metric reports
	pollInterval   int64         // seconds between metric reads
	mu             *sync.Mutex   // mutex for thread safety
	logger         *zerolog.Logger
	retryIntervals map[int]time.Duration // mapping retry attempt to delay
	hasher         *utils.Hasher
	rateLimit      int64
	publicKey      *rsa.PublicKey // public key to encrypt message
}

// NewMetricsAgent initializes and returns a new MetricsAgent with configuration values.
func NewMetricsAgent(route string, publicKey *rsa.PublicKey, cfg *config.AgentConfig, logger *zerolog.Logger) *MetricsAgent {
	agent := &MetricsAgent{
		serverAddress:  "http://" + cfg.ServerAddress,
		route:          route,
		client:         utils.NewHTTPClient(5 * time.Second),
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
		publicKey: publicKey,
	}

	// add retry mechanism
	agent.client.SetRetryCount(3).
		SetRetryAfter(func(client *resty.Client, response *resty.Response) (time.Duration, error) {
			return agent.retryIntervals[response.Request.Attempt], nil
		}).SetRetryMaxWaitTime(5 * time.Second)

	return agent
}

// ReadMetrics reads runtime memory metrics and refreshes the in-memory cache.
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

// SendBatchMetricsJSON sends all metrics from memory as a single gzip-compressed JSON batch.
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

// SendMetricsJSON sends all metrics from memory individually as JSON.
func (m *MetricsAgent) SendMetricsJSON() error {
	// get all metrics from memory
	allMetrics := m.memory.GetAllMetrics()
	if len(allMetrics) == 0 {
		return errors.New("no metrics passed")
	}

	// send every metric retrieved from memory
	for _, metric := range allMetrics {
		err := m.sendMetrics(metric)
		if err != nil {
			return err
		}
	}

	// after sending metrics set poll count to 0
	m.pollCount = 0

	return nil
}

func (m *MetricsAgent) sendMetrics(metric ...models.Metrics) error {
	if len(metric) == 0 {
		m.logger.Error().Caller().Str("func", "*MetricsAgent.sendMetrics").Msg("no metric was passed!")
		return errors.New("no metric was passed")
	}
	// construct a route
	route, pathJoinError := url.JoinPath(m.serverAddress, m.route, "/")
	if pathJoinError != nil {
		m.logger.Err(pathJoinError).Caller().Str("func", "*MetricsAgent.sendMetrics").Msg("url join error")
		return fmt.Errorf("url join error: %w", pathJoinError)
	}

	// construct headers
	headers := map[string]string{
		"Content-Type":     "application/json",
		"Content-Encoding": "gzip",
	}

	// include hash of the body
	if m.hasher != nil {
		hashedMetric, hashingError := m.hasher.HashMetrics(metric...)
		if hashingError != nil {
			m.logger.Err(hashingError).Caller().Str("func", "*MetricsAgent.sendMetrics").Msg("error occurred during hashing metric")
			return hashingError
		}
		headers["HashSHA256"] = fmt.Sprintf("%x", hashedMetric)
	}

	// gzip encode metric
	compressedMetric, compressionError := gzipCompress(metric...)
	if compressionError != nil {
		m.logger.Err(compressionError).Caller().Str("func", "*MetricsAgent.sendMetrics").Msg("error occurred during gzip compression")
		return compressionError
	}

	if m.publicKey != nil {
		encryptedMessage, err := utils.EncryptData(compressedMetric, m.publicKey)
		if err != nil {
			m.logger.Err(err).Str("func", "*MetricsAgent.sendMetrics").Msg("error encrypting data")
			return err
		}
		compressedMetric = encryptedMessage.Bytes()
	}

	m.logger.Debug().Any("metric", metric).Any("hash", headers).Send()

	var response models.Metrics
	_, sendMetricError := m.client.R().
		SetHeaders(headers).
		SetBody(compressedMetric).
		SetResult(&response).
		Post(route)
	if sendMetricError != nil {
		m.logger.Err(sendMetricError).Caller().Str("func", "*MetricsAgent.sendMetrics").Msg("error occurred during sending metric")
		return fmt.Errorf("error occurred during sending metric: %w", sendMetricError)
	}

	m.logger.Info().Caller().Str("func", "*MetricsAgent.sendMetrics").Any("request", compressedMetric).Any("response", response).Msg("metric is sent!")
	return nil
}

// SendMetrics sends all cached metrics individually using HTTP POST requests.
// It constructs a route based on metric type and value.
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
func (m *MetricsAgent) ReadMetricsGenerator(ctx context.Context, pollInterval *time.Ticker, reportInterval *time.Ticker) chan []models.Metrics {
	metricsChannel := make(chan []models.Metrics)

	go func() {
		defer close(metricsChannel)
		for {
			select {
			case <-ctx.Done():
				m.logger.Debug().Msg("read metrics job generator stopped")
				return
			case <-pollInterval.C:
				m.logger.Debug().Str("func", "ReadMetricsGenerator").Msg("time to READ metrics")
				_ = m.ReadMetrics()
			case <-reportInterval.C:
				m.logger.Debug().Str("func", "ReadMetricsGenerator").Msg("time to SEND metrics")
				allMetrics := m.memory.GetAllMetrics()
				metricsChannel <- allMetrics
			}
		}
	}()

	return metricsChannel
}

// SendMetricsWorker consumes metric batches from a channel and sends them.
func (m *MetricsAgent) SendMetricsWorker(metricBatches <-chan []models.Metrics) {
	for batch := range metricBatches {
		m.logger.Debug().Any("batch", batch).Msg("worker is called")
		_ = m.sendMetrics(batch...)
		m.pollCount = 0
	}
	m.logger.Debug().Msg("worker stopped working")
}

// Run starts the MetricsAgent lifecycle, including reading metrics and sending
// them with worker goroutines.
func (m *MetricsAgent) Run() error {
	// creating ctx to end jobs generator and sender workers
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT,
	)
	defer stop()

	// reading metrics part
	pollTicker, reportTicker := getTickers(time.Duration(m.pollInterval)*time.Second, time.Duration(m.reportInterval)*time.Second)
	m.logger.Debug().Str("func", "Run").Msg("preparing to run goroutine for reading metrics")
	jobs := m.ReadMetricsGenerator(ctx, pollTicker, reportTicker)

	// creating wait groups to cancel workers after finishing
	wg := new(sync.WaitGroup)

	// creating workers
	m.logger.Debug().Str("func", "Run").Msg("creating workers")
	m.withWorkers(wg, func() {
		defer wg.Done() // runs only when jobs channel is closed
		m.SendMetricsWorker(jobs)
	}, m.rateLimit)
	m.logger.Debug().Str("func", "Run").Msg("workers are created")

	// wait for the shutdown signal
	<-ctx.Done()
	m.logger.Info().Msg("shutdown signal received")

	// stop all the tickers
	pollTicker.Stop()
	reportTicker.Stop()

	// wait till all the workers are finished
	wg.Wait()

	m.logger.Info().Msg("agent shutdown completed")
	return nil
}

func getTickers(pollIntervalDuration time.Duration, reportIntervalDuration time.Duration) (*time.Ticker, *time.Ticker) {
	return time.NewTicker(pollIntervalDuration), time.NewTicker(reportIntervalDuration)
}

func (m *MetricsAgent) withWorkers(wg *sync.WaitGroup, fn func(), count int64) {
	for i := range count {
		wg.Add(1)
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

func gzipCompress(metric ...models.Metrics) ([]byte, error) {
	var jsonData []byte
	var err error

	if len(metric) == 1 {
		jsonData, err = json.Marshal(metric[0])
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metric: %w", err)
		}
	} else {
		// сериализуем metric в JSON
		jsonData, err = json.Marshal(metric)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metrics: %w", err)
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
