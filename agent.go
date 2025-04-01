package vigilant

import (
	"net/http"
	"time"
)

// globalAgent is the global agent instance
var globalAgent *agent

// Init initializes the agent, it should be called once when the program is starting
// Before calling this, all other Vigilant functions will be noops
func Init(config *AgentConfig) {
	globalAgent = newAgent(config)
	globalAgent.start()
}

// Shutdown shuts down the agent, it should be called once when the program is shutting down
func Shutdown() error {
	return globalAgent.shutdown()
}

// agent is the internal representation of the agent
// it handles the sending of logs, errors, and metrics to Vigilant
type agent struct {
	name        string
	level       LogLevel
	token       string
	passthrough bool
	noopLogs    bool
	noopErrors  bool
	noopMetrics bool

	batcher *batcher
}

// newAgent creates a new agent from the given config
func newAgent(config *AgentConfig) *agent {
	batcher := newBatcher(
		config.Token,
		getEndpoint(config),
		&http.Client{},
	)
	return &agent{
		name:        config.Name,
		level:       config.Level,
		token:       config.Token,
		passthrough: config.Passthrough,
		noopLogs:    config.NoopLogs,
		noopErrors:  config.NoopErrors,
		noopMetrics: config.NoopMetrics,
		batcher:     batcher,
	}
}

// start starts the agent
func (a *agent) start() {
	if a.noopLogs && a.noopErrors && a.noopMetrics {
		return
	}
	a.batcher.start()
}

// shutdown shuts down the agent
func (a *agent) shutdown() error {
	a.batcher.stop()
	return nil
}

// sendLog sends a log message to the agent
func (a *agent) sendLog(
	level LogLevel,
	message string,
	attrs map[string]string,
) {
	if !isLevelEnabled(level, a.level) {
		return
	}

	a.updateAttributes(attrs)

	if a.passthrough {
		writeLogPassthrough(level, message, attrs)
	}
	if a.noopLogs {
		return
	}

	logMessage := &logMessage{
		Timestamp:  time.Now(),
		Level:      level,
		Body:       message,
		Attributes: attrs,
	}

	a.batcher.addLog(logMessage)
}

// sendError sends an error to the agent
func (a *agent) sendError(
	err error,
	location errorLocation,
	details errorDetails,
	attrs map[string]string,
) {
	a.updateAttributes(attrs)

	if a.passthrough {
		writeErrorPassthrough(err, attrs)
	}
	if a.noopErrors {
		return
	}

	errorMessage := &errorMessage{
		Timestamp:  time.Now(),
		Details:    details,
		Location:   location,
		Attributes: attrs,
	}

	a.batcher.addError(errorMessage)
}

// sendMetric sends a metric to the agent
func (a *agent) sendMetric(
	name string,
	value float64,
	attrs map[string]string,
) {
	a.updateAttributes(attrs)

	if a.passthrough {
		writeMetricPassthrough(name, value, attrs)
	}
	if a.noopMetrics {
		return
	}

	metricMessage := &metricMessage{
		Timestamp:  time.Now(),
		Name:       name,
		Value:      value,
		Attributes: attrs,
	}

	a.batcher.addMetric(metricMessage)
}

// updateAttributes adds the service name attribute to the given attributes
func (a *agent) updateAttributes(attrs map[string]string) {
	if a == nil {
		return
	}
	attrs["service.name"] = a.name
}
