package freeze

// SnapshotOption is a function that configures a SnapshotConfig.
type SnapshotOption func(*SnapshotConfig)

// SnapshotConfig holds configuration for snapshot scrubbing and filtering.
type SnapshotConfig struct {
	Scrubbers []Scrubber
	Ignore    []IgnorePattern
}

// newSnapshotConfig creates a new SnapshotConfig with the given options applied.
func newSnapshotConfig(opts []SnapshotOption) *SnapshotConfig {
	config := &SnapshotConfig{
		Scrubbers: []Scrubber{},
		Ignore:    []IgnorePattern{},
	}
	for _, opt := range opts {
		opt(config)
	}
	return config
}

// WithScrubber adds a custom scrubber to the configuration.
func WithScrubber(scrubber Scrubber) SnapshotOption {
	return func(c *SnapshotConfig) {
		c.Scrubbers = append(c.Scrubbers, scrubber)
	}
}

// WithIgnorePattern adds a custom ignore pattern to the configuration.
func WithIgnorePattern(pattern IgnorePattern) SnapshotOption {
	return func(c *SnapshotConfig) {
		c.Ignore = append(c.Ignore, pattern)
	}
}
