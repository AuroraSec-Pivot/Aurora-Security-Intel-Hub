package config

import "time"

type Config struct {
	Archive  ArchiveConfig  `yaml:"archive"`
	Pipeline PipelineConfig `yaml:"pipeline"`
	HTTP     HTTPConfig     `yaml:"http"`
	Notifier NotifierConfig `yaml:"notifier"`
	Sources  []SourceConfig `yaml:"sources"`
}

type ArchiveConfig struct {
	Driver         string `yaml:"driver"`
	Path           string `yaml:"path"`
	WAL            bool   `yaml:"wal"`
	BusyTimeoutMS  int    `yaml:"busy_timeout_ms"`
	MigrateOnStart bool   `yaml:"migrate_on_start"`
}

type PipelineConfig struct {
	Mode                      string        `yaml:"mode"`
	Interval                  time.Duration `yaml:"interval"`
	Jitter                    time.Duration `yaml:"jitter"`
	Concurrency               int           `yaml:"concurrency"`
	MaxPushPerRun             int           `yaml:"max_push_per_run"`
	DefaultPushPolicy         string        `yaml:"default_push_policy"`
	DropIfPublishedBeforeDays int           `yaml:"drop_if_published_before_days"`
	DryRun                    bool          `yaml:"dry_run"`
}

type HTTPConfig struct {
	Timeout   time.Duration `yaml:"timeout"`
	UserAgent string        `yaml:"user_agent"`
	Proxy     string        `yaml:"proxy"`
	Retry     RetryConfig   `yaml:"retry"`
	RateLimit RateLimit     `yaml:"rate_limit"`
}

type RetryConfig struct {
	Max     int           `yaml:"max"`
	Backoff time.Duration `yaml:"backoff"`
}

type RateLimit struct {
	DefaultRPS float64            `yaml:"default_rps"`
	PerHost    map[string]float64 `yaml:"per_host"`
}

type NotifierConfig struct {
	WeCom WeComConfig `yaml:"wecom"`
}

type WeComConfig struct {
	WebhookURL string        `yaml:"webhook_url"`
	Template   string        `yaml:"template"`
	Timeout    time.Duration `yaml:"timeout"`
	Retry      RetryConfig   `yaml:"retry"`
}

type SourceConfig struct {
	SourceID          string         `yaml:"source_id"`
	Type              string         `yaml:"type"`
	URL               string         `yaml:"url"`
	Interval          time.Duration  `yaml:"interval"`
	Priority          string         `yaml:"priority"`
	FingerprintFields []string       `yaml:"fingerprint_fields"`
	PushPolicy        string         `yaml:"push_policy"`
	Enabled           bool           `yaml:"enabled"`
	Tags              []string       `yaml:"tags"`
	Notes             map[string]any `yaml:"notes"`
}
