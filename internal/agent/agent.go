package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/MKhiriev/stunning-adventure/internal/config"
	"github.com/MKhiriev/stunning-adventure/internal/utils"
	"github.com/MKhiriev/stunning-adventure/models"
	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog"
	"math/rand/v2"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"sync"
	"time"
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
}

func NewMetricsAgent(route string, cfg *config.AgentConfig, logger *zerolog.Logger) *MetricsAgent {
	return &MetricsAgent{
		serverAddress:  "http://" + cfg.ServerAddress,
		route:          route,
		client:         resty.New(),
		memory:         NewStorage(),
		pollCount:      0,
		reportInterval: cfg.ReportInterval,
		pollInterval:   cfg.PollInterval,
		logger:         logger,
		mu:             &sync.Mutex{},
	}
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

func (m *MetricsAgent) SendMetricsJSON() error {
	// get all metrics from memory
	allMetrics := m.memory.GetAllMetrics()
	if len(allMetrics) == 0 {
		return errors.New("no metrics passed")
	}

	route, pathJoinError := url.JoinPath(m.serverAddress, m.route, "/")
	if pathJoinError != nil {
		m.logger.Err(pathJoinError).Caller().Str("func", "*MetricsAgent.SendMetricsJSON").Msg("url join error")
		return fmt.Errorf("url join error: %w", pathJoinError)
	}
	// send every metric retrieved from memory
	for _, metric := range allMetrics {
		var response models.Metrics

		// gzip encode metric
		compressedMetric, compressionError := gzipCompress(metric)
		if compressionError != nil {
			m.logger.Err(compressionError).Caller().Str("func", "*MetricsAgent.SendMetricsJSON").Msg("error occurred during gzip compression")
			return compressionError
		}

		_, sendMetricError := resty.New().R().
			SetHeaders(map[string]string{
				"Content-Type":     "application/json",
				"Content-Encoding": "gzip",
			}).
			SetBody(compressedMetric).
			SetResult(&response).
			Post(route)
		if sendMetricError != nil {
			m.logger.Err(sendMetricError).Caller().Str("func", "*MetricsAgent.SendMetricsJSON").Msg("error occurred during sending metric")
			return fmt.Errorf("error occurred during sending metric: %w", sendMetricError)
		}

		m.logger.Info().Caller().Str("func", "*MetricsAgent.SendMetricsJSON").Any("request", compressedMetric).Any("response", response).Msg("request & response from server")
	}
	// after sending metrics set poll count to 0
	m.pollCount = 0

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
	return nil
}

func (m *MetricsAgent) Run() error {
	var err error

	utils.RunWithTicker(func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		err = m.ReadMetrics()
	}, time.Duration(m.pollInterval)*time.Second)

	utils.RunWithTicker(func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		err = m.SendMetricsJSON()
	}, time.Duration(m.reportInterval)*time.Second)
	if err != nil {
		m.logger.Err(err).Caller().Str("func", "*MetricsAgent.Run").Msg("error occurred during running metrics-agent")
		return err
	}

	select {} // block main routine forever
}

func (m *MetricsAgent) getSliceOfMetrics(memStats runtime.MemStats) []models.Metrics {
	m.pollCount++
	return []models.Metrics{
		gaugeMetric("Alloc", float64(memStats.Alloc)),
		gaugeMetric("BuckHashSys", float64(memStats.BuckHashSys)),
		gaugeMetric("Frees", float64(memStats.Frees)),
		gaugeMetric("GCCPUFraction", memStats.GCCPUFraction),
		gaugeMetric("GCSys", float64(memStats.GCSys)),
		gaugeMetric("HeapAlloc", float64(memStats.HeapAlloc)),
		gaugeMetric("HeapIdle", float64(memStats.HeapIdle)),
		gaugeMetric("HeapInuse", float64(memStats.HeapInuse)),
		gaugeMetric("HeapObjects", float64(memStats.HeapObjects)),
		gaugeMetric("HeapReleased", float64(memStats.HeapReleased)),
		gaugeMetric("HeapSys", float64(memStats.HeapSys)),
		gaugeMetric("LastGC", float64(memStats.LastGC)),
		gaugeMetric("Lookups", float64(memStats.Lookups)),
		gaugeMetric("MCacheInuse", float64(memStats.MCacheInuse)),
		gaugeMetric("MCacheSys", float64(memStats.MCacheSys)),
		gaugeMetric("MSpanInuse", float64(memStats.MSpanInuse)),
		gaugeMetric("MSpanSys", float64(memStats.MSpanSys)),
		gaugeMetric("Mallocs", float64(memStats.Mallocs)),
		gaugeMetric("NextGC", float64(memStats.NextGC)),
		gaugeMetric("NumForcedGC", float64(memStats.NumForcedGC)),
		gaugeMetric("NumGC", float64(memStats.NumGC)),
		gaugeMetric("OtherSys", float64(memStats.OtherSys)),
		gaugeMetric("PauseTotalNs", float64(memStats.PauseTotalNs)),
		gaugeMetric("StackInuse", float64(memStats.StackInuse)),
		gaugeMetric("StackSys", float64(memStats.StackSys)),
		gaugeMetric("Sys", float64(memStats.Sys)),
		gaugeMetric("TotalAlloc", float64(memStats.TotalAlloc)),
		counterMetric("PollCount", m.pollCount),
		gaugeMetric("RandomValue", rand.Float64()),
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

func gaugeMetric(name string, value float64) models.Metrics {
	return models.Metrics{
		ID:    name,
		MType: models.Gauge,
		Value: &value,
	}
}

func counterMetric(name string, value int64) models.Metrics {
	return models.Metrics{
		ID:    name,
		MType: models.Counter,
		Delta: &value,
	}
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
