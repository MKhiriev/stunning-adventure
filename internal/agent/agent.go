package agent

import (
	"errors"
	"fmt"
	"github.com/MKhiriev/stunning-adventure/models"
	"github.com/go-resty/resty/v2"
	"math/rand/v2"
	"net/http"
	"runtime"
	"strconv"
	"time"
)

const DefaultPollInterval = time.Duration(2 * time.Second)
const DefaultReportInterval = time.Duration(10 * time.Second)

type MetricsAgent struct {
	ServerAddress  string
	Route          string
	Client         *resty.Client
	Memory         *AgentStorage
	PollCount      int64
	ReportInterval time.Duration
	PollInterval   time.Duration
}

func NewMetricsAgent(serverAddress string, route string, reportInterval, pollInterval time.Duration) *MetricsAgent {
	return &MetricsAgent{
		ServerAddress:  "http://" + serverAddress,
		Route:          route,
		Client:         resty.New(),
		Memory:         NewStorage(),
		PollCount:      0,
		ReportInterval: reportInterval,
		PollInterval:   pollInterval,
	}
}

func (m *MetricsAgent) ReadMetrics() error {
	memStats := runtime.MemStats{}
	runtime.ReadMemStats(&memStats)

	allMetrics := m.getSliceOfMetrics(memStats)
	if len(allMetrics) == 0 {
		return errors.New("error occurred during getting MemStats metrics: no metrics in MemStats")
	}

	m.Memory.RefreshAllMetrics(allMetrics...)
	return nil
}

func (m *MetricsAgent) SendMetrics() error {
	// get all metrics from memory
	allMetrics := m.Memory.GetAllMetrics()
	if len(allMetrics) == 0 {
		return errors.New("no metrics are passed")
	}

	// send every metric retrieved from memory
	for _, metric := range allMetrics {
		// construct route based on metric's type
		route, err := m.getRoute(metric)
		if err != nil {
			return err
		}

		response, err := m.Client.R().
			SetHeader("Content-Type", "text/plain").
			Post(route)
		if err != nil {
			return err
		}

		if response.StatusCode() != http.StatusOK {
			return fmt.Errorf("error during metrics sending: %s", response.Status())
		}
	}
	return nil
}

func (m *MetricsAgent) Run() error {
	timeToSendMetrics := time.Now().Add(m.ReportInterval)
	for {
		// Read Metrics
		if err := m.ReadMetrics(); err != nil {
			return err
		}
		// wait 2 seconds
		time.Sleep(m.PollInterval)

		thisMoment := time.Now()
		// check if it's time to send metrics
		if thisMoment.Equal(timeToSendMetrics) || thisMoment.After(timeToSendMetrics) {
			// Send Metrics
			if err := m.SendMetrics(); err != nil {
				return err
			}
			// set time to send metrics to the server
			timeToSendMetrics = time.Now().Add(m.ReportInterval)
		}
	}
}

func (m *MetricsAgent) getSliceOfMetrics(memStats runtime.MemStats) []models.Metrics {
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
		counterMetric("PollCount", m.PollCount+1),
		gaugeMetric("RandomValue", rand.Float64()),
	}
}

func (m *MetricsAgent) getRoute(metric models.Metrics) (string, error) {
	if metric.MType == models.Counter {
		// check if Counter's Delta is not nil
		if metric.Delta == nil {
			return "", errors.New("no metric's data has been passed: field Delta is nil")
		}

		return fmt.Sprintf("%s/%s/%s/%s/%d", m.ServerAddress, m.Route,
			metric.MType, metric.ID, *metric.Delta), nil
	}

	if metric.MType == models.Gauge {
		// check if Gauge's Value is not nil
		if metric.Value == nil {
			return "", errors.New("no metric's data has been passed: field Value in nil")
		}

		return fmt.Sprintf("%s/%s/%s/%s/%s", m.ServerAddress, m.Route,
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
