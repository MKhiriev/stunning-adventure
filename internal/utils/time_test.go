package utils

import (
	"encoding/json"
	"testing"
	"time"
)

func TestDuration_UnmarshalJSON_Valid(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		expected time.Duration
	}{
		{
			name:     "seconds",
			jsonData: `"10s"`,
			expected: 10 * time.Second,
		},
		{
			name:     "milliseconds",
			jsonData: `"250ms"`,
			expected: 250 * time.Millisecond,
		},
		{
			name:     "minutes",
			jsonData: `"5m"`,
			expected: 5 * time.Minute,
		},
		{
			name:     "combined duration",
			jsonData: `"1h30m"`,
			expected: time.Hour + 30*time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d Duration
			if err := json.Unmarshal([]byte(tt.jsonData), &d); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if d.Duration != tt.expected {
				t.Fatalf(
					"unexpected duration: want %v, got %v",
					tt.expected,
					d.Duration,
				)
			}
		})
	}
}

func TestDuration_UnmarshalJSON_InvalidJSONType(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
	}{
		{
			name:     "number",
			jsonData: `10`,
		},
		{
			name:     "object",
			jsonData: `{}`,
		},
		{
			name:     "array",
			jsonData: `[]`,
		},
		{
			name:     "null",
			jsonData: `null`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d Duration
			if err := json.Unmarshal([]byte(tt.jsonData), &d); err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestDuration_UnmarshalJSON_InvalidDurationFormat(t *testing.T) {
	var d Duration
	err := json.Unmarshal([]byte(`"not-a-duration"`), &d)

	if err == nil {
		t.Fatal("expected duration parse error, got nil")
	}
}

func TestDuration_UnmarshalJSON_InStruct(t *testing.T) {
	type config struct {
		Interval Duration `json:"interval"`
	}

	var cfg config
	err := json.Unmarshal([]byte(`{"interval":"2s"}`), &cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Interval.Duration != 2*time.Second {
		t.Fatalf(
			"unexpected duration: want %v, got %v",
			2*time.Second,
			cfg.Interval.Duration,
		)
	}
}
