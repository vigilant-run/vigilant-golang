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
func Init(config *VigilantConfig) {
	if globalAgent != nil {
		return
	}
	globalAgent = newVigilant(config)
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
// it handles the sending of logs
type agent struct {
	name        string
	level       LogLevel
	token       string
	passthrough bool
	noop        bool

	logBatcher      *logBatcher
	metricCollector *metricCollector
}

// newVigilant creates a new agent from the given config
func newVigilant(config *VigilantConfig) *agent {
	logBatcher := newLogBatcher(
		config.Token,
		getEndpoint(config),
		&http.Client{},
	)
	metricCollector := newMetricCollector(
		time.Minute,
		config.Token,
		getEndpoint(config),
		&http.Client{},
	)
	return &agent{
		name:            config.Name,
		level:           config.Level,
		token:           config.Token,
		passthrough:     config.Passthrough,
		noop:            config.Noop,
		logBatcher:      logBatcher,
		metricCollector: metricCollector,
	}
}

// start starts the agent
func (a *agent) start() {
	if a.noop {
		return
	}
	a.logBatcher.start()
	a.metricCollector.start()
}

// shutdown shuts down the agent
func (a *agent) shutdown() error {
	a.logBatcher.stop()
	a.metricCollector.stop()
	return nil
}

// captureLog captures a log message
func (a *agent) captureLog(
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

	if a.noop {
		return
	}

	logMessage := &logMessage{
		Timestamp:  time.Now(),
		Level:      level,
		Body:       message,
		Attributes: updatedAttrs,
	}

	a.logBatcher.addLog(logMessage)
}

// captureCounter captures a counter metric
func (a *agent) captureCounter(
	name string,
	value float64,
	tags map[string]string,
) {
	if a.noop {
		return
	}

	event := &metricEvent{
		timestamp: time.Now(),
		name:      name,
		value:     value,
		tags:      tags,
	}

	a.metricCollector.addCounter(event)
}

// captureGauge captures a gauge metric
func (a *agent) captureGauge(
	name string,
	value float64,
	tags map[string]string,
) {
	if a.noop {
		return
	}

	event := &metricEvent{
		timestamp: time.Now(),
		name:      name,
		value:     value,
		tags:      tags,
	}

	a.metricCollector.addGauge(event)
}

// captureHistogram captures a histogram metric
func (a *agent) captureHistogram(
	name string,
	value float64,
	tags map[string]string,
) {
	if a.noop {
		return
	}

	event := &metricEvent{
		timestamp: time.Now(),
		name:      name,
		value:     value,
		tags:      tags,
	}

	a.metricCollector.addHistogram(event)
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
