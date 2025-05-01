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
func (a *agent) captureLog(log *logMessage) {
	if !isLevelEnabled(log.Level, a.level) {
		return
	}

	log.Attributes = a.withBaseAttributes(log.Attributes)

	if a.passthrough {
		writeLogPassthrough(log.Level, log.Body, log.Attributes)
	}

	if a.noop {
		return
	}

	a.logBatcher.addLog(log)
}

// captureCounter captures a counter metric
func (a *agent) captureCounter(counter *counterEvent) {
	if a.noop {
		return
	}

	a.metricCollector.addCounter(counter)
}

// captureGauge captures a gauge metric
func (a *agent) captureGauge(gauge *gaugeEvent) {
	if a.noop {
		return
	}

	a.metricCollector.addGauge(gauge)
}

// captureHistogram captures a histogram metric
func (a *agent) captureHistogram(histogram *histogramEvent) {
	if a.noop {
		return
	}

	a.metricCollector.addHistogram(histogram)
}

// withBaseAttributes adds the service name attribute to the given attributes
func (a *agent) withBaseAttributes(attrs map[string]string) map[string]string {
	updatedAttrs := make(map[string]string)
	if attrs != nil {
		maps.Copy(updatedAttrs, attrs)
	}
	updatedAttrs["service"] = a.name
	return updatedAttrs
}
