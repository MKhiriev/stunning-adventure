package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/MKhiriev/stunning-adventure/models"
	"github.com/go-chi/chi/v5"
)

func (h *Handler) Audit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		aw := &AuditResponseWriter{
			ResponseWriter: w,
		}

		var body []byte
		if r.Body != nil {
			// get request body - request.Clone() doesn't clone body
			var err error
			body, err = io.ReadAll(r.Body)
			if err != nil {
				h.logger.Err(err).Str("func", "*Handler.SendEvent").Msg("failed to read request body")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(body))
		}

		next.ServeHTTP(aw, r)

		if aw.status != http.StatusOK {
			h.logger.Debug().Str("func", "*Handler.SendEvent").
				Int("status", aw.status).
				Msg("not sending audit report")
			return
		}

		var auditEvent models.AuditEvent
		ts := time.Now()
		ipAddress, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			h.logger.Err(err).Str("func", "*Handler.SendEvent").
				Str("reason", "cannot get source ip address").
				Msg("error creating audit event")
			return
		}

		// if metrics were in http-body
		if len(body) != 0 {
			metricsFromBody, err := getMetricsFromBody(body)
			if err != nil {
				h.logger.Err(err).Str("func", "*Handler.SendEvent").Msg("error extracting metrics from body")
				return
			}
			metricNames := getListOfMetricsNames(metricsFromBody)
			auditEvent, err = models.NewAuditEvent(ipAddress, ts, metricNames...)
			if err != nil {
				h.logger.Err(err).Str("func", "*Handler.SendEvent").
					Str("source", "request body").
					Strs("metric name", metricNames).
					Msg("cannot create audit event")
				return
			}
		} else { // if metric were from url
			metricName := chi.URLParam(r, "metricName")
			auditEvent, err = models.NewAuditEvent(ipAddress, ts, metricName)
			if err != nil {
				h.logger.Err(err).Str("func", "*Handler.SendEvent").
					Str("source", "url param").
					Str("metric name", metricName).
					Msg("cannot create audit event")
				return
			}
		}

		// send audit event
		go func() {
			err = h.auditService.NotifyAll(context.Background(), auditEvent)
			if err != nil {
				h.logger.Err(err).Any("audit event", auditEvent).Msg("error sending audit event")
				return
			}

			h.logger.Debug().Str("func", "*Handler.SendEvent").Any("audit event", auditEvent).Msg("event is sent to all recipients")
		}()
	})
}

func getMetricsFromBody(body []byte) ([]models.Metrics, error) {
	var metricsFromBody []models.Metrics

	if err := json.Unmarshal(body, &metricsFromBody); err != nil {
		return nil, fmt.Errorf("error unmarshalling metrics from body: %w", err)
	}

	return metricsFromBody, nil
}

func getListOfMetricsNames(metrics []models.Metrics) []string {
	metricsNames := make([]string, 0, len(metrics))

	for _, metric := range metrics {
		metricsNames = append(metricsNames, metric.ID)
	}

	return metricsNames
}

type AuditResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *AuditResponseWriter) WriteHeader(statusCode int) {
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *AuditResponseWriter) Write(data []byte) (int, error) {
	return w.ResponseWriter.Write(data)
}
