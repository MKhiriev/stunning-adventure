package models

import (
	"errors"
	"time"
)

// AuditEvent is log of audit event
type AuditEvent struct {
	TimeStamp int64    `json:"ts"`         // Unix timestamp of event
	Metrics   []string `json:"metrics"`    // Metric names that were sent by agent
	IpAddress string   `json:"ip_address"` // IP address from metrics agent
}

func NewAuditEvent(ipAddress string, ts time.Time, metricNames ...string) (AuditEvent, error) {
	if len(metricNames) == 0 {
		return AuditEvent{}, errors.New("no metrics' names were passed")
	}

	return AuditEvent{
		TimeStamp: ts.Unix(),
		Metrics:   metricNames,
		IpAddress: ipAddress,
	}, nil
}
