package vigilant

// AgentConfig is the configuration for the Vigilant agent
type AgentConfig struct {
	Name string

	Level LogLevel

	Token       string
	Endpoint    string
	Passthrough bool
	Insecure    bool

	NoopLogs    bool
	NoopErrors  bool
	NoopMetrics bool
}

// AgentConfigBuilder is the builder for the Vigilant agent configuration
type AgentConfigBuilder struct {
	Name        *string
	Level       *LogLevel
	Token       *string
	Endpoint    *string
	Passthrough *bool
	Insecure    *bool
	NoopLogs    *bool
	NoopErrors  *bool
	NoopMetrics *bool
}

// NewAgentConfigBuilder creates a new agent configuration builder
func NewAgentConfigBuilder() *AgentConfigBuilder {
	return &AgentConfigBuilder{}
}

// WithName sets the name of the agent
func (b *AgentConfigBuilder) WithName(name string) *AgentConfigBuilder {
	b.Name = &name
	return b
}

// WithLevel sets the level of the agent
func (b *AgentConfigBuilder) WithLevel(level LogLevel) *AgentConfigBuilder {
	b.Level = &level
	return b
}

// WithToken sets the token of the agent
func (b *AgentConfigBuilder) WithToken(token string) *AgentConfigBuilder {
	b.Token = &token
	return b
}

// WithEndpoint sets the endpoint of the agent
func (b *AgentConfigBuilder) WithEndpoint(endpoint string) *AgentConfigBuilder {
	b.Endpoint = &endpoint
	return b
}

// WithPassthrough sets the passthrough of the agent
func (b *AgentConfigBuilder) WithPassthrough(passthrough bool) *AgentConfigBuilder {
	b.Passthrough = &passthrough
	return b
}

// WithInsecure sets the insecure of the agent
func (b *AgentConfigBuilder) WithInsecure(insecure bool) *AgentConfigBuilder {
	b.Insecure = &insecure
	return b
}

// WithNoopLogs sets the agent to not send logs
func (b *AgentConfigBuilder) WithNoopLogs(noop bool) *AgentConfigBuilder {
	b.NoopLogs = &noop
	return b
}

// WithNoopErrors sets the agent to not send errors
func (b *AgentConfigBuilder) WithNoopErrors(noop bool) *AgentConfigBuilder {
	b.NoopErrors = &noop
	return b
}

// WithNoopMetrics sets the agent to not send metrics
func (b *AgentConfigBuilder) WithNoopMetrics(noop bool) *AgentConfigBuilder {
	b.NoopMetrics = &noop
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
		NoopErrors:  false,
		NoopMetrics: false,
	}

	if b.Name != nil {
		config.Name = *b.Name
	}

	if b.Level != nil {
		config.Level = *b.Level
	}

	if b.Token != nil {
		config.Token = *b.Token
	}

	if b.Endpoint != nil {
		config.Endpoint = *b.Endpoint
	}

	if b.Passthrough != nil {
		config.Passthrough = *b.Passthrough
	}

	if b.Insecure != nil {
		config.Insecure = *b.Insecure
	}

	if b.NoopLogs != nil {
		config.NoopLogs = *b.NoopLogs
	}

	if b.NoopErrors != nil {
		config.NoopErrors = *b.NoopErrors
	}

	if b.NoopMetrics != nil {
		config.NoopMetrics = *b.NoopMetrics
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
		NoopErrors:  true,
		NoopMetrics: true,
	}
}
