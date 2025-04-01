package vigilant

import (
	"maps"
	"net/http"
	"time"
)

// globalAgent is the global agent instance
var globalAgent *agent

// Init initializes the agent, it should be called once when the program is starting
// Before calling this, all other Vigilant functions will be noops
func Init(config *AgentConfig) {
	if globalAgent != nil {
		return
	}
	globalAgent = newAgent(config)
	globalAgent.start()
}

// Shutdown shuts down the agent, it should be called once when the program is shutting down
func Shutdown() error {
	if globalAgent == nil {
		return nil
	}
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
	noopAlerts  bool

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
	if a.noopLogs && a.noopErrors && a.noopMetrics && a.noopAlerts {
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

	updatedAttrs := a.withBaseAttributes(attrs)

	if a.passthrough {
		writeLogPassthrough(level, message, updatedAttrs)
	}

	if a.noopLogs {
		return
	}

	logMessage := &logMessage{
		Timestamp:  time.Now(),
		Level:      level,
		Body:       message,
		Attributes: updatedAttrs,
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
	updatedAttrs := a.withBaseAttributes(attrs)

	if a.passthrough {
		writeErrorPassthrough(err, updatedAttrs)
	}

	if a.noopErrors {
		return
	}

	errorMessage := &errorMessage{
		Timestamp:  time.Now(),
		Details:    details,
		Location:   location,
		Attributes: updatedAttrs,
	}

	a.batcher.addError(errorMessage)
}

// sendAlert sends an alert to the agent
func (a *agent) sendAlert(
	title string,
	attrs map[string]string,
) {
	updatedAttrs := a.withBaseAttributes(attrs)

	if a.passthrough {
		writeAlertPassthrough(title, updatedAttrs)
	}

	if a.noopAlerts {
		return
	}

	alertMessage := &alertMessage{
		Timestamp:  time.Now(),
		Title:      title,
		Attributes: updatedAttrs,
	}

	a.batcher.addAlert(alertMessage)
}

// sendMetric sends a metric to the agent
func (a *agent) sendMetric(
	name string,
	value float64,
	attrs map[string]string,
) {
	updatedAttrs := a.withBaseAttributes(attrs)

	if a.passthrough {
		writeMetricPassthrough(name, value, updatedAttrs)
	}

	if a.noopMetrics {
		return
	}

	metricMessage := &metricMessage{
		Timestamp:  time.Now(),
		Name:       name,
		Value:      value,
		Attributes: updatedAttrs,
	}

	a.batcher.addMetric(metricMessage)
}

// withBaseAttributes adds the service name attribute to the given attributes
func (a *agent) withBaseAttributes(attrs map[string]string) map[string]string {
	updatedAttrs := make(map[string]string)
	if attrs != nil {
		maps.Copy(updatedAttrs, attrs)
	}
	updatedAttrs["service.name"] = a.name
	return updatedAttrs
}
