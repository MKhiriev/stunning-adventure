package agent

import (
	"errors"
	"fmt"
	"github.com/MKhiriev/stunning-adventure/models"
	"math"
	"math/rand/v2"
	"net/http"
	"runtime"
	"time"
)

const (
	PollInterval   = time.Duration(1 * time.Second)
	ReportInterval = time.Duration(5 * time.Second)
)

type MetricsAgent struct {
	ServerAddress string
	Route         string
	Client        *http.Client
	Memory        *AgentStorage
	PollCount     int64
}

func NewMetricsAgent(serverAddress string, route string) *MetricsAgent {
	return &MetricsAgent{
		ServerAddress: "http://" + serverAddress,
		Route:         route,
		Client:        &http.Client{Timeout: 10 * time.Second},
		Memory:        NewStorage(),
		PollCount:     0,
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

		request, err := http.NewRequest(http.MethodPost, route, nil)
		if err != nil {
			return err
		}
		request.Header.Add("Content-Type", "text/plain")
		response, err := m.Client.Do(request)
		if err != nil {
			return err
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			return fmt.Errorf("error during metrics sending: %s", response.Status)
		}
	}
	return nil
}

func (m *MetricsAgent) Run() error {
	timeToSendMetrics := time.Now().Add(ReportInterval)
	for {
		// Read Metrics
		if err := m.ReadMetrics(); err != nil {
			return err
		}
		// wait 2 seconds
		time.Sleep(PollInterval)

		thisMoment := time.Now()
		// check if it's time to send metrics
		if thisMoment.Equal(timeToSendMetrics) || thisMoment.After(timeToSendMetrics) {
			// Send Metrics
			if err := m.SendMetrics(); err != nil {
				return err
			}
			// set time to send metrics to the server
			timeToSendMetrics = time.Now().Add(ReportInterval)
		}
	}
}

func (m *MetricsAgent) getSliceOfMetrics(memStats runtime.MemStats) []models.Metrics {
	return []models.Metrics{
		gaugeMetric("Alloc", math.Float64frombits(memStats.Alloc)),
		gaugeMetric("BuckHashSys", math.Float64frombits(memStats.BuckHashSys)),
		gaugeMetric("Frees", math.Float64frombits(memStats.Frees)),
		gaugeMetric("GCCPUFraction", memStats.GCCPUFraction),
		gaugeMetric("GCSys", math.Float64frombits(memStats.GCSys)),
		gaugeMetric("HeapAlloc", math.Float64frombits(memStats.HeapAlloc)),
		gaugeMetric("HeapIdle", math.Float64frombits(memStats.HeapIdle)),
		gaugeMetric("HeapInuse", math.Float64frombits(memStats.HeapInuse)),
		gaugeMetric("HeapObjects", math.Float64frombits(memStats.HeapObjects)),
		gaugeMetric("HeapReleased", math.Float64frombits(memStats.HeapReleased)),
		gaugeMetric("HeapSys", math.Float64frombits(memStats.HeapSys)),
		gaugeMetric("LastGC", math.Float64frombits(memStats.LastGC)),
		gaugeMetric("Lookups", math.Float64frombits(memStats.Lookups)),
		gaugeMetric("MCacheInuse", math.Float64frombits(memStats.MCacheInuse)),
		gaugeMetric("MCacheSys", math.Float64frombits(memStats.MCacheSys)),
		gaugeMetric("MSpanInuse", math.Float64frombits(memStats.MSpanInuse)),
		gaugeMetric("MSpanSys", math.Float64frombits(memStats.MSpanSys)),
		gaugeMetric("Mallocs", math.Float64frombits(memStats.Mallocs)),
		gaugeMetric("NextGC", math.Float64frombits(memStats.NextGC)),
		gaugeMetric("NumForcedGC", float64(math.Float32frombits(memStats.NumForcedGC))),
		gaugeMetric("NumGC", float64(math.Float32frombits(memStats.NumGC))),
		gaugeMetric("OtherSys", math.Float64frombits(memStats.OtherSys)),
		gaugeMetric("PauseTotalNs", math.Float64frombits(memStats.PauseTotalNs)),
		gaugeMetric("StackInuse", math.Float64frombits(memStats.StackInuse)),
		gaugeMetric("StackSys", math.Float64frombits(memStats.StackSys)),
		gaugeMetric("Sys", math.Float64frombits(memStats.Sys)),
		gaugeMetric("TotalAlloc", math.Float64frombits(memStats.TotalAlloc)),
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

		return fmt.Sprintf("%s/%s/%s/%s/%.0f", m.ServerAddress, m.Route,
			metric.MType, metric.ID, *metric.Value), nil
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
