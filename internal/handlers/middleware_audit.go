package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/MKhiriev/stunning-adventure/models"
	"github.com/go-chi/chi/v5"
)

func (h *Handler) Audit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		aw := &AuditResponseWriter{
			ResponseWriter: w,
			responseData:   responseData{},
		}

		// get request body - request.Clone() doesn't clone body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			h.logger.Err(err).Str("func", "*Handler.Audit").Msg("failed to read request body")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(body))

		h.logger.Debug().Str("func", "*Handler.Audit").
			Any("body", body).
			Any("url", r.URL).
			Bool("is read body nil?", body == nil).
			Bool("is read body empty byte slice?", len(body) == 0).
			Bool("is r.Body nil?", r.Body == nil).Send()

		next.ServeHTTP(aw, r)

		if aw.responseData.status != http.StatusOK {
			h.logger.Debug().Str("func", "*Handler.Audit").
				Int("status", aw.responseData.status).
				Msg("not sending audit report")
			return
		}

		var auditEvent models.AuditEvent
		ts := time.Now()
		ipAddress := r.RemoteAddr

		// get metric names
		if len(body) == 0 {
			metricName := chi.URLParam(r, "metricName")
			auditEvent, err = models.NewAuditEvent(ipAddress, ts, metricName)
			if err != nil {
				h.logger.Err(err).Str("func", "*Handler.Audit").
					Str("source", "url param").
					Str("metric name", metricName).
					Msg("cannot create audit event")
				return
			}
		} else {
			metricsFromBody, err := getMetricsFromBody(body)
			if err != nil {
				h.logger.Err(err).Str("func", "*Handler.Audit").Msg("error extracting metrics from body")
				return
			}
			metricNames := getListOfMetricsNames(metricsFromBody)
			auditEvent, err = models.NewAuditEvent(ipAddress, ts, metricNames...)
			if err != nil {
				h.logger.Err(err).Str("func", "*Handler.Audit").
					Str("source", "request body").
					Strs("metric name", metricNames).
					Msg("cannot create audit event")
				return
			}
		}

		// send audit event
		err = h.auditService.SendAudit(auditEvent)
		if err != nil {
			h.logger.Err(err).Any("audit event", auditEvent).Msg("error sending audit event")
			return
		}

		h.logger.Debug().Str("func", "*Handler.Audit").Any("audit event", auditEvent).Msg("audit is sent")
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
	var metricsNames []string

	for _, metric := range metrics {
		metricsNames = append(metricsNames, metric.ID)
	}

	return metricsNames
}

type AuditResponseWriter struct {
	http.ResponseWriter
	responseData responseData
}

func (w *AuditResponseWriter) WriteHeader(statusCode int) {
	w.responseData.status = statusCode
	w.ResponseWriter.WriteHeader(w.responseData.status)
}

func (w *AuditResponseWriter) Write(data []byte) (int, error) {
	w.responseData.body = data
	size, err := w.ResponseWriter.Write(data)
	w.responseData.size += size
	return size, err
}
