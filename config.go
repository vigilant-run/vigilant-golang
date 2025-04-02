package vigilant

// AgentConfig is the configuration for the Vigilant agent
type AgentConfig struct {
	// Name is the name of the service being monitored
	Name string

	// Level is the level of log the agent will send to the server
	Level LogLevel

	// Token is the Vigilant API token
	Token string

	// Endpoint is the endpoint of the Vigilant server
	Endpoint string

	// Passthrough is whether to print logs, alerts, and metrics to stdout
	Passthrough bool

	// Insecure is whether to use HTTP instead of HTTPS
	Insecure bool

	// NoopLogs is whether to not send logs to the server
	NoopLogs bool

	// NoopAlerts is whether to not send alerts to the server
	NoopAlerts bool

	// NoopMetrics is whether to not send metrics to the server
	NoopMetrics bool
}

// AgentConfigBuilder is the builder for the Vigilant agent configuration
type AgentConfigBuilder struct {
	name        *string
	level       *LogLevel
	token       *string
	endpoint    *string
	passthrough *bool
	insecure    *bool
	noopLogs    *bool
	noopAlerts  *bool
	noopMetrics *bool
}

// NewAgentConfigBuilder creates a new agent configuration builder
func NewAgentConfigBuilder() *AgentConfigBuilder {
	return &AgentConfigBuilder{}
}

// WithName sets the name of the agent
func (b *AgentConfigBuilder) WithName(name string) *AgentConfigBuilder {
	b.name = &name
	return b
}

// WithLevel sets the level of the agent
func (b *AgentConfigBuilder) WithLevel(level LogLevel) *AgentConfigBuilder {
	b.level = &level
	return b
}

// WithToken sets the token of the agent
func (b *AgentConfigBuilder) WithToken(token string) *AgentConfigBuilder {
	b.token = &token
	return b
}

// WithEndpoint sets the endpoint of the agent
func (b *AgentConfigBuilder) WithEndpoint(endpoint string) *AgentConfigBuilder {
	b.endpoint = &endpoint
	return b
}

// WithPassthrough sets the passthrough of the agent
func (b *AgentConfigBuilder) WithPassthrough(passthrough bool) *AgentConfigBuilder {
	b.passthrough = &passthrough
	return b
}

// WithInsecure sets the insecure of the agent
func (b *AgentConfigBuilder) WithInsecure(insecure bool) *AgentConfigBuilder {
	b.insecure = &insecure
	return b
}

// WithNoopLogs sets the agent to not send logs
func (b *AgentConfigBuilder) WithNoopLogs(noop bool) *AgentConfigBuilder {
	b.noopLogs = &noop
	return b
}

// WithNoopAlerts sets the agent to not send alerts
func (b *AgentConfigBuilder) WithNoopAlerts(noop bool) *AgentConfigBuilder {
	b.noopAlerts = &noop
	return b
}

// WithNoopMetrics sets the agent to not send metrics
func (b *AgentConfigBuilder) WithNoopMetrics(noop bool) *AgentConfigBuilder {
	b.noopMetrics = &noop
	return b
}

// Build builds the agent configuration
func (b *AgentConfigBuilder) Build() *AgentConfig {
	config := &AgentConfig{
		Name:        "server-name",
		Level:       LEVEL_TRACE,
		Token:       "tk_1234567890",
		Endpoint:    "ingress.vigilant.run",
		Passthrough: false,
		Insecure:    false,
		NoopLogs:    false,
		NoopAlerts:  false,
		NoopMetrics: false,
	}

	if b.name != nil {
		config.Name = *b.name
	}

	if b.level != nil {
		config.Level = *b.level
	}

	if b.token != nil {
		config.Token = *b.token
	}

	if b.endpoint != nil {
		config.Endpoint = *b.endpoint
	}

	if b.passthrough != nil {
		config.Passthrough = *b.passthrough
	}

	if b.insecure != nil {
		config.Insecure = *b.insecure
	}

	if b.noopLogs != nil {
		config.NoopLogs = *b.noopLogs
	}

	if b.noopAlerts != nil {
		config.NoopAlerts = *b.noopAlerts
	}

	if b.noopMetrics != nil {
		config.NoopMetrics = *b.noopMetrics
	}

	return config
}

// NewNoopAgentConfig creates a new noop agent config, this is useful for testing
func NewNoopAgentConfig() *AgentConfig {
	return &AgentConfig{
		Name:        "server-name",
		Level:       LEVEL_TRACE,
		Token:       "tk_1234567890",
		Endpoint:    "ingress.vigilant.run",
		Insecure:    false,
		Passthrough: true,
		NoopLogs:    true,
		NoopMetrics: true,
		NoopAlerts:  true,
	}
}
