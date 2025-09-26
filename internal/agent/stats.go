package agent

import (
	"fmt"
	"math/rand/v2"
	"runtime"
	"time"

	"github.com/MKhiriev/stunning-adventure/models"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

func (m *MetricsAgent) getSliceOfMetrics(memStats runtime.MemStats) []models.Metrics {
	m.pollCount++
	var metrics []models.Metrics
	virtualMem, _ := mem.VirtualMemory()

	metrics = []models.Metrics{
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
		gaugeMetric("TotalMemory", float64(virtualMem.Total)),
		gaugeMetric("FreeMemory", float64(virtualMem.Free)),
	}

	allCPUs, _ := cpu.Percent(time.Second, true)
	for i, cpuPercentage := range allCPUs {
		metrics = append(metrics, gaugeMetric(fmt.Sprintf("CPUutilization%d", i), cpuPercentage))
	}

	return metrics
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
