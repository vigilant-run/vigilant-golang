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

	batcher *batcher
}

// newVigilant creates a new agent from the given config
func newVigilant(config *VigilantConfig) *agent {
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
		noop:        config.Noop,
		batcher:     batcher,
	}
}

// start starts the agent
func (a *agent) start() {
	if a.noop {
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

	if a.noop {
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

// withBaseAttributes adds the service name attribute to the given attributes
func (a *agent) withBaseAttributes(attrs map[string]string) map[string]string {
	updatedAttrs := make(map[string]string)
	if attrs != nil {
		maps.Copy(updatedAttrs, attrs)
	}
	updatedAttrs["service.name"] = a.name
	return updatedAttrs
}
