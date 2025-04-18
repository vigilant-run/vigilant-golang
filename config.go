package vigilant

// VigilantConfig is the configuration for the Vigilant agent
type VigilantConfig struct {
	// Name is the name of the service being monitored
	Name string

	// Level is the level of log the agent will send to the server
	Level LogLevel

	// Token is the Vigilant API token
	Token string

	// Endpoint is the endpoint of the Vigilant server
	Endpoint string

	// Passthrough is whether to print logs to stdout
	Passthrough bool

	// Insecure is whether to use HTTP instead of HTTPS
	Insecure bool

	// Noop is whether to not send logs to the server
	Noop bool
}

// VigilantConfigBuilder is the builder for the Vigilant agent configuration
type VigilantConfigBuilder struct {
	name        *string
	level       *LogLevel
	token       *string
	endpoint    *string
	passthrough *bool
	insecure    *bool
	noop        *bool
}

// NewAgentConfigBuilder creates a new agent configuration builder
func NewAgentConfigBuilder() *VigilantConfigBuilder {
	return &VigilantConfigBuilder{}
}

// WithName sets the name of the agent
func (b *VigilantConfigBuilder) WithName(name string) *VigilantConfigBuilder {
	b.name = &name
	return b
}

// WithLevel sets the level of the agent
func (b *VigilantConfigBuilder) WithLevel(level LogLevel) *VigilantConfigBuilder {
	b.level = &level
	return b
}

// WithToken sets the token of the agent
func (b *VigilantConfigBuilder) WithToken(token string) *VigilantConfigBuilder {
	b.token = &token
	return b
}

// WithEndpoint sets the endpoint of the agent
func (b *VigilantConfigBuilder) WithEndpoint(endpoint string) *VigilantConfigBuilder {
	b.endpoint = &endpoint
	return b
}

// WithPassthrough sets the passthrough of the agent
func (b *VigilantConfigBuilder) WithPassthrough(passthrough bool) *VigilantConfigBuilder {
	b.passthrough = &passthrough
	return b
}

// WithInsecure sets the insecure of the agent
func (b *VigilantConfigBuilder) WithInsecure(insecure bool) *VigilantConfigBuilder {
	b.insecure = &insecure
	return b
}

// WithNoop sets the agent to not send logs
func (b *VigilantConfigBuilder) WithNoop(noop bool) *VigilantConfigBuilder {
	b.noop = &noop
	return b
}

// Build builds the agent configuration
func (b *VigilantConfigBuilder) Build() *VigilantConfig {
	config := &VigilantConfig{
		Name:        "server-name",
		Level:       LEVEL_TRACE,
		Token:       "tk_1234567890",
		Endpoint:    "ingress.vigilant.run",
		Passthrough: false,
		Insecure:    false,
		Noop:        false,
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

	if b.noop != nil {
		config.Noop = *b.noop
	}

	return config
}

// NewNoopAgentConfig creates a new noop agent config, this is useful for testing
func NewNoopAgentConfig() *VigilantConfig {
	return &VigilantConfig{
		Name:        "server-name",
		Level:       LEVEL_TRACE,
		Token:       "tk_1234567890",
		Endpoint:    "ingress.vigilant.run",
		Insecure:    false,
		Passthrough: true,
		Noop:        true,
	}
}
