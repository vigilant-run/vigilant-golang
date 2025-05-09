package vigilant

import (
	"maps"
	"net/http"
	"time"
)

// globalInstance is the global Vigilant instance
var globalInstance *instance

// Init initializes the Vigilant instance, it should be called once when the program is starting
// Before calling this, all other Vigilant functions will be noops
func Init(config *VigilantConfig) {
	if globalInstance != nil {
		return
	}
	globalInstance = newVigilant(config)
	globalInstance.start()
}

// Shutdown shuts down the Vigilant instance, it should be called once when the program is shutting down
func Shutdown() error {
	if globalInstance == nil {
		return nil
	}
	return globalInstance.shutdown()
}

// instance is the internal representation of the Vigilant instance
// it handles the sending of logs and metrics to the server
type instance struct {
	name        string
	level       LogLevel
	token       string
	passthrough bool
	noop        bool

	logBatcher      *logBatcher
	metricCollector *metricCollector
}

// newVigilant creates a new Vigilant instance from the given config
func newVigilant(config *VigilantConfig) *instance {
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
	return &instance{
		name:            config.Name,
		level:           config.Level,
		token:           config.Token,
		passthrough:     config.Passthrough,
		noop:            config.Noop,
		logBatcher:      logBatcher,
		metricCollector: metricCollector,
	}
}

// start starts the Vigilant instance
func (a *instance) start() {
	if a.noop {
		return
	}
	a.logBatcher.start()
	a.metricCollector.start()
}

// shutdown shuts down the Vigilant instance
func (a *instance) shutdown() error {
	a.logBatcher.stop()
	a.metricCollector.stop()
	return nil
}

// captureLog captures a log message
func (a *instance) captureLog(log *logMessage) {
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
func (a *instance) captureCounter(counter *counterEvent) {
	if a.noop {
		return
	}

	a.metricCollector.addCounter(counter)
}

// captureGauge captures a gauge metric
func (a *instance) captureGauge(gauge *gaugeEvent) {
	if a.noop {
		return
	}

	a.metricCollector.addGauge(gauge)
}

// captureHistogram captures a histogram metric
func (a *instance) captureHistogram(histogram *histogramEvent) {
	if a.noop {
		return
	}

	a.metricCollector.addHistogram(histogram)
}

// withBaseAttributes adds the service name attribute to the given attributes
func (a *instance) withBaseAttributes(attrs map[string]string) map[string]string {
	updatedAttrs := make(map[string]string)
	if attrs != nil {
		maps.Copy(updatedAttrs, attrs)
	}
	updatedAttrs["service"] = a.name
	return updatedAttrs
}
