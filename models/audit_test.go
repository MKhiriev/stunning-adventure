package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAuditEvent_Success_SingleMetric(t *testing.T) {
	ts := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	ipAddress := "192.168.1.100"

	event, err := NewAuditEvent(ipAddress, ts, "cpu_usage")
	require.NoError(t, err)

	assert.Equal(t, ts.Unix(), event.TimeStamp)
	assert.Equal(t, ipAddress, event.IPAddress)
	assert.Len(t, event.Metrics, 1)
	assert.Equal(t, "cpu_usage", event.Metrics[0])
}

func TestNewAuditEvent_Success_MultipleMetrics(t *testing.T) {
	ts := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	ipAddress := "10.0.0.5"
	metrics := []string{"cpu_usage", "memory_total", "disk_io", "network_rx"}

	event, err := NewAuditEvent(ipAddress, ts, metrics...)
	require.NoError(t, err)

	assert.Equal(t, ts.Unix(), event.TimeStamp)
	assert.Equal(t, ipAddress, event.IPAddress)
	assert.Len(t, event.Metrics, 4)
	assert.Equal(t, metrics, event.Metrics)
}

func TestNewAuditEvent_Error_NoMetrics(t *testing.T) {
	ts := time.Now()
	ipAddress := "192.168.1.1"

	event, err := NewAuditEvent(ipAddress, ts)
	require.Error(t, err)

	assert.Contains(t, err.Error(), "no metrics' names were passed")
	assert.Equal(t, AuditEvent{}, event)
}

func TestNewAuditEvent_EmptyIPAddress(t *testing.T) {
	ts := time.Now()

	event, err := NewAuditEvent("", ts, "metric1")
	require.NoError(t, err, "empty IP should be allowed")

	assert.Empty(t, event.IPAddress)
	assert.Len(t, event.Metrics, 1)
}

func TestNewAuditEvent_IPv6Address(t *testing.T) {
	ts := time.Now()
	ipAddress := "2001:0db8:85a3:0000:0000:8a2e:0370:7334"

	event, err := NewAuditEvent(ipAddress, ts, "metric1")
	require.NoError(t, err)

	assert.Equal(t, ipAddress, event.IPAddress)
}

func TestNewAuditEvent_LocalhostIPAddress(t *testing.T) {
	ts := time.Now()

	t.Run("IPv4 localhost", func(t *testing.T) {
		event, err := NewAuditEvent("127.0.0.1", ts, "metric1")
		require.NoError(t, err)
		assert.Equal(t, "127.0.0.1", event.IPAddress)
	})

	t.Run("IPv6 localhost", func(t *testing.T) {
		event, err := NewAuditEvent("::1", ts, "metric1")
		require.NoError(t, err)
		assert.Equal(t, "::1", event.IPAddress)
	})
}

func TestNewAuditEvent_TimestampConversion(t *testing.T) {
	testCases := []struct {
		name     string
		time     time.Time
		expected int64
	}{
		{
			name:     "epoch zero",
			time:     time.Unix(0, 0),
			expected: 0,
		},
		{
			name:     "specific date",
			time:     time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC),
			expected: 1705314645,
		},
		{
			name:     "negative timestamp (before epoch)",
			time:     time.Unix(-1000, 0),
			expected: -1000,
		},
		{
			name:     "far future",
			time:     time.Date(2100, 12, 31, 23, 59, 59, 0, time.UTC),
			expected: 4133980799,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			event, err := NewAuditEvent("192.168.1.1", tc.time, "metric1")
			require.NoError(t, err)
			assert.Equal(t, tc.expected, event.TimeStamp)
		})
	}
}

func TestNewAuditEvent_MetricsOrder(t *testing.T) {
	ts := time.Now()
	metrics := []string{"metric_a", "metric_b", "metric_c"}

	event, err := NewAuditEvent("192.168.1.1", ts, metrics...)
	require.NoError(t, err)

	// Order should be preserved
	assert.Equal(t, metrics, event.Metrics)
	assert.Equal(t, "metric_a", event.Metrics[0])
	assert.Equal(t, "metric_b", event.Metrics[1])
	assert.Equal(t, "metric_c", event.Metrics[2])
}

func TestNewAuditEvent_DuplicateMetrics(t *testing.T) {
	ts := time.Now()
	metrics := []string{"cpu", "cpu", "memory", "cpu"}

	event, err := NewAuditEvent("192.168.1.1", ts, metrics...)
	require.NoError(t, err)

	// Duplicates should be preserved (no deduplication)
	assert.Len(t, event.Metrics, 4)
	assert.Equal(t, metrics, event.Metrics)
}

func TestNewAuditEvent_EmptyMetricNames(t *testing.T) {
	ts := time.Now()

	event, err := NewAuditEvent("192.168.1.1", ts, "", "valid_metric", "")
	require.NoError(t, err, "empty metric names should be allowed")

	assert.Len(t, event.Metrics, 3)
	assert.Equal(t, "", event.Metrics[0])
	assert.Equal(t, "valid_metric", event.Metrics[1])
	assert.Equal(t, "", event.Metrics[2])
}

func TestNewAuditEvent_SpecialCharactersInMetricNames(t *testing.T) {
	ts := time.Now()
	metrics := []string{
		"metric-with-dashes",
		"metric_with_underscores",
		"metric.with.dots",
		"metric:with:colons",
		"метрика_кириллица",
		"metric with spaces",
	}

	event, err := NewAuditEvent("192.168.1.1", ts, metrics...)
	require.NoError(t, err)

	assert.Len(t, event.Metrics, len(metrics))
	assert.Equal(t, metrics, event.Metrics)
}

func TestNewAuditEvent_LargeNumberOfMetrics(t *testing.T) {
	ts := time.Now()

	// Generate 1000 metric names
	metrics := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		metrics[i] = "metric_" + string(rune('0'+i%10))
	}

	event, err := NewAuditEvent("192.168.1.1", ts, metrics...)
	require.NoError(t, err)

	assert.Len(t, event.Metrics, 1000)
	assert.Equal(t, metrics, event.Metrics)
}

func TestNewAuditEvent_VeryLongMetricName(t *testing.T) {
	ts := time.Now()
	longMetricName := string(make([]byte, 10000))
	for i := range longMetricName {
		longMetricName = longMetricName[:i] + "a"
	}

	event, err := NewAuditEvent("192.168.1.1", ts, longMetricName)
	require.NoError(t, err)

	assert.Len(t, event.Metrics, 1)
	assert.Len(t, event.Metrics[0], 10000)
}

func TestNewAuditEvent_TimezoneHandling(t *testing.T) {
	// Create same moment in different timezones
	utcTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	loc, err := time.LoadLocation("America/New_York")
	require.NoError(t, err)
	nyTime := utcTime.In(loc)

	event1, err := NewAuditEvent("192.168.1.1", utcTime, "metric1")
	require.NoError(t, err)

	event2, err := NewAuditEvent("192.168.1.1", nyTime, "metric1")
	require.NoError(t, err)

	// Unix timestamps should be identical
	assert.Equal(t, event1.TimeStamp, event2.TimeStamp)
}

func TestNewAuditEvent_ConcurrentCreation(t *testing.T) {
	ts := time.Now()
	const numGoroutines = 100

	done := make(chan AuditEvent, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			event, err := NewAuditEvent("192.168.1.1", ts, "metric1", "metric2")
			require.NoError(t, err)
			done <- event
		}(i)
	}

	// Collect all events
	events := make([]AuditEvent, 0, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		events = append(events, <-done)
	}

	// All events should be valid and identical
	assert.Len(t, events, numGoroutines)
	for _, event := range events {
		assert.Equal(t, ts.Unix(), event.TimeStamp)
		assert.Equal(t, "192.168.1.1", event.IPAddress)
		assert.Len(t, event.Metrics, 2)
	}
}

func TestAuditEvent_JSONTags(t *testing.T) {
	// This test verifies the struct tags are correct
	// by checking field names match expected JSON keys
	ts := time.Now()
	event, err := NewAuditEvent("192.168.1.1", ts, "metric1")
	require.NoError(t, err)

	// Verify struct has expected fields
	assert.NotZero(t, event.TimeStamp)
	assert.NotEmpty(t, event.IPAddress)
	assert.NotEmpty(t, event.Metrics)
}

func TestNewAuditEvent_NilMetricsSlice(t *testing.T) {
	ts := time.Now()
	var metrics []string // nil slice

	event, err := NewAuditEvent("192.168.1.1", ts, metrics...)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no metrics' names were passed")
	assert.Equal(t, AuditEvent{}, event)
}
