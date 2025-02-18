package vigilant

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// MetricsHandlerConfig is the configuration for the metrics handler
type MetricsHandlerConfig struct {
	Name     string
	Endpoint string
	Token    string
	Insecure bool
	Noop     bool
}

// MetricsHandlerConfigBuilder is the builder for the metrics handler configuration
type MetricsHandlerConfigBuilder struct {
	Name     string
	Endpoint string
	Token    string
	Insecure bool
	Noop     bool
}

// NewMetricsHandlerConfigBuilder creates a new metrics handler configuration builder
func NewMetricsHandlerConfigBuilder() *MetricsHandlerConfigBuilder {
	return &MetricsHandlerConfigBuilder{}
}

// WithName sets the name of the metrics handler
func (b *MetricsHandlerConfigBuilder) WithName(name string) *MetricsHandlerConfigBuilder {
	b.Name = name
	return b
}

// WithEndpoint sets the endpoint of the metrics handler
func (b *MetricsHandlerConfigBuilder) WithEndpoint(endpoint string) *MetricsHandlerConfigBuilder {
	b.Endpoint = endpoint
	return b
}

// WithToken sets the token of the metrics handler
func (b *MetricsHandlerConfigBuilder) WithToken(token string) *MetricsHandlerConfigBuilder {
	b.Token = token
	return b
}

// WithInsecure sets the insecure of the metrics handler
func (b *MetricsHandlerConfigBuilder) WithInsecure() *MetricsHandlerConfigBuilder {
	b.Insecure = true
	return b
}

// WithNoop sets the noop of the metrics handler
func (b *MetricsHandlerConfigBuilder) WithNoop() *MetricsHandlerConfigBuilder {
	b.Noop = true
	return b
}

// Build builds the metrics handler configuration
func (b *MetricsHandlerConfigBuilder) Build() *MetricsHandlerConfig {
	config := &MetricsHandlerConfig{
		Name:     b.Name,
		Endpoint: b.Endpoint,
		Token:    b.Token,
		Insecure: b.Insecure,
		Noop:     b.Noop,
	}

	if b.Name == "" {
		config.Name = "service-name"
	}

	if b.Endpoint == "" {
		config.Endpoint = "ingress.vigilant.run"
	}

	if b.Token == "" {
		config.Token = "tk_1234567890"
	}

	return config
}

// InitMetricsHandler initializes the metrics handler
func InitMetricsHandler(config *MetricsHandlerConfig) {
	globalMetricsHandler = newMetricsHandler(config.Name, config.Endpoint, config.Token, config.Insecure, config.Noop)
}

// ShutdownMetricsHandler shuts down the metrics handler
func ShutdownMetricsHandler() error {
	return globalMetricsHandler.shutdown()
}

// EmitMetric emits a metric
func EmitMetric(name string, value float64, attrs ...Attribute) {
	if globalMetricsHandler == nil {
		return
	}
	globalMetricsHandler.capture(name, value, attrs...)
}

var globalMetricsHandler *metricsHandler

// metricsHandler is a handler for metrics
type metricsHandler struct {
	name     string
	endpoint string
	token    string
	insecure bool
	noop     bool

	metricsQueue chan *metricMessage
	batchStop    chan struct{}
	wg           sync.WaitGroup
}

// newMetricsHandler creates a new metricsHandler
func newMetricsHandler(
	name string,
	endpoint string,
	token string,
	insecure bool,
	noop bool,
) *metricsHandler {
	var formattedEndpoint string
	if insecure {
		formattedEndpoint = fmt.Sprintf("http://%s/api/message", endpoint)
	} else {
		formattedEndpoint = fmt.Sprintf("https://%s/api/message", endpoint)
	}

	metricsHandler := &metricsHandler{
		name:         name,
		endpoint:     formattedEndpoint,
		token:        token,
		insecure:     insecure,
		noop:         noop,
		metricsQueue: make(chan *metricMessage, 1000),
		batchStop:    make(chan struct{}),
	}

	metricsHandler.startBatcher()
	return metricsHandler
}

// capture captures a metric and sends it to Vigilant
func (m *metricsHandler) capture(name string, value float64, attrs ...Attribute) {
	if m.noop || name == "" {
		return
	}

	attrsMap := make(map[string]string)
	for _, attr := range attrs {
		attrsMap[attr.Key] = attr.Value
	}
	attrsMap["service.name"] = m.name

	select {
	case m.metricsQueue <- &metricMessage{
		Timestamp:  time.Now(),
		Name:       name,
		Value:      value,
		Attributes: attrsMap,
	}:
	default:
	}
}

// Shutdown shuts down the metrics handler
func (m *metricsHandler) shutdown() error {
	m.stopBatcher()

	done := make(chan struct{})
	go func() {
		m.wg.Wait()
		close(done)
	}()

	<-done
	return nil
}

// startBatcher starts the batcher goroutine
func (m *metricsHandler) startBatcher() {
	m.wg.Add(1)
	go m.runBatcher()
}

// runBatcher is the batcher goroutine
func (m *metricsHandler) runBatcher() {
	defer m.wg.Done()

	const maxBatchSize = 100
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	var batch []*metricMessage

	for {
		select {
		case <-m.batchStop:
			if len(batch) > 0 {
				m.sendBatch(batch)
			}
			return

		case msg := <-m.metricsQueue:
			if msg == nil {
				continue
			}

			batch = append(batch, msg)
			if len(batch) >= maxBatchSize {
				m.sendBatch(batch)
				batch = nil
			}

		case <-ticker.C:
			if len(batch) > 0 {
				m.sendBatch(batch)
				batch = nil
			}
		}
	}
}

// stopBatcher closes the batchStop channel
func (m *metricsHandler) stopBatcher() {
	close(m.batchStop)
}

// sendBatch sends a batch of metrics
func (m *metricsHandler) sendBatch(metrics []*metricMessage) {
	if len(metrics) == 0 {
		return
	}

	batch := &messageBatch{
		Token:   m.token,
		Type:    messageTypeMetric,
		Metrics: metrics,
	}

	batchBytes, err := json.Marshal(batch)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", m.endpoint, bytes.NewBuffer(batchBytes))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+m.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
}
