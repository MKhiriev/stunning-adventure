package agent

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/MKhiriev/stunning-adventure/internal/config"
	"github.com/MKhiriev/stunning-adventure/internal/grpc"
	"github.com/MKhiriev/stunning-adventure/internal/utils"
	"github.com/MKhiriev/stunning-adventure/models"
	"github.com/rs/zerolog"
)

// MetricsAgent provides a concrete implementation of a local metrics collection agent.
// It reads metrics from the host, caches them in memory, and sends them to a remote server.
type MetricsAgent struct {
	serverAddress  string        // full server address including http://
	route          string        // server route for sending metrics
	client         MetricsSender // HTTP/gRPC client for sending metrics
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
	realIPAddress  string
}

// NewMetricsAgent initializes and returns a new MetricsAgent with configuration values.
func NewMetricsAgent(route string, publicKey *rsa.PublicKey, cfg *config.AgentConfig, logger *zerolog.Logger) (*MetricsAgent, error) {
	retryIntervals := map[int]time.Duration{
		1: 1 * time.Second,
		2: 3 * time.Second,
		3: 5 * time.Second,
	}

	client, err := GetAgentMetricsSender(route, retryIntervals, utils.NewHasher(cfg.HashKey), publicKey, cfg, logger)
	if err != nil {
		return nil, err
	}

	agent := &MetricsAgent{
		serverAddress:  cfg.ServerAddress,
		client:         client,
		memory:         NewStorage(),
		pollCount:      0,
		reportInterval: cfg.ReportInterval,
		pollInterval:   cfg.PollInterval,
		logger:         logger,
		mu:             &sync.Mutex{},
		rateLimit:      cfg.RateLimit,
		publicKey:      publicKey,
	}

	return agent, nil
}

func GetAgentMetricsSender(route string, retryIntervals map[int]time.Duration, hasher *utils.Hasher, publicKey *rsa.PublicKey, cfg *config.AgentConfig, logger *zerolog.Logger) (MetricsSender, error) {
	switch {
	case cfg.GrpcServerAddress != "":
		logger.Debug().Str("func", "GetAgentMetricsSender").Msg("creating gRPC metrics sender...")
		return createGRPCMetricsSender(retryIntervals, cfg, logger)
	case cfg.ServerAddress != "":
		logger.Debug().Str("func", "GetAgentMetricsSender").Msg("creating HTTP metrics sender...")
		return createHTTPMetricsSender(route, retryIntervals, hasher, publicKey, cfg, logger)
	}

	return nil, nil
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
	jobs := m.readMetricsGenerator(ctx, pollTicker, reportTicker)

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

// SendMetricsWorker consumes metric batches from a channel and sends them.
func (m *MetricsAgent) SendMetricsWorker(metricBatches <-chan []models.Metrics) {
	for batch := range metricBatches {
		m.logger.Debug().Any("batch", batch).Msg("worker is called")
		_ = m.sendMetrics(batch...)
		m.pollCount = 0
	}
	m.logger.Debug().Msg("worker stopped working")
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

func (m *MetricsAgent) sendMetrics(metrics ...models.Metrics) error {
	if len(metrics) == 0 {
		m.logger.Error().Str("func", "*MetricsAgent.sendMetrics").Msg("no metric was passed!")
		return errors.New("no metric was passed")
	}

	sendMetricError := m.client.Send(metrics...)
	if sendMetricError != nil {
		m.logger.Err(sendMetricError).Str("func", "*MetricsAgent.sendMetrics").Msg("error occurred during sending metric")
		return fmt.Errorf("error occurred during sending metric: %w", sendMetricError)
	}

	m.logger.Info().Str("func", "*MetricsAgent.sendMetrics").Msg("metrics are sent!")
	return nil
}

// readMetricsGenerator reads metrics and returns a channel that will feed the worker metrics for sending
func (m *MetricsAgent) readMetricsGenerator(ctx context.Context, pollInterval *time.Ticker, reportInterval *time.Ticker) chan []models.Metrics {
	metricsChannel := make(chan []models.Metrics)

	go func() {
		defer close(metricsChannel)
		for {
			select {
			case <-ctx.Done():
				m.logger.Debug().Msg("read metrics job generator stopped")
				return
			case <-pollInterval.C:
				m.logger.Debug().Str("func", "readMetricsGenerator").Msg("time to READ metrics")
				_ = m.ReadMetrics()
			case <-reportInterval.C:
				m.logger.Debug().Str("func", "readMetricsGenerator").Msg("time to SEND metrics")
				allMetrics := m.memory.GetAllMetrics()
				metricsChannel <- allMetrics
			}
		}
	}()

	return metricsChannel
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

func createHTTPMetricsSender(route string, retryIntervals map[int]time.Duration, hasher *utils.Hasher, publicKey *rsa.PublicKey, cfg *config.AgentConfig, logger *zerolog.Logger) (MetricsSender, error) {
	if err := utils.CheckIfValidIPAddress(cfg.ServerAddress); err != nil {
		return nil, fmt.Errorf("invalid server address: %w", err)
	}

	client, err := NewHTTPMetricsSender(cfg.ServerAddress, route, retryIntervals, hasher, publicKey, logger)
	if err != nil {
		return nil, fmt.Errorf("unexpected error creating agent metrics sender: %w", err)
	}

	return client, nil
}

func createGRPCMetricsSender(retryIntervals map[int]time.Duration, cfg *config.AgentConfig, logger *zerolog.Logger) (MetricsSender, error) {
	if err := utils.CheckIfValidIPAddress(cfg.GrpcServerAddress); err != nil {
		return nil, fmt.Errorf("invalid grpc server address: %w", err)
	}

	retryConfig := grpc.RetryConfig{
		MaxAttempts: len(retryIntervals),
		Intervals:   retryIntervals,
	}

	client, err := NewGRPCMetricsSender(cfg.GrpcServerAddress, retryConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("unexpected error creating agent metrics sender: %w", err)
	}

	return client, nil
}
