package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

var envPattern = regexp.MustCompile(`\$\{[A-Za-z_][A-Za-z0-9_]*\}`)

func Load(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	expanded, err := expandEnv(string(b))
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}

	applyDefaults(&cfg)
	return &cfg, nil
}

func expandEnv(s string) (string, error) {
	out := envPattern.ReplaceAllStringFunc(s, func(m string) string {
		// m like ${ENV_NAME}
		key := strings.TrimSuffix(strings.TrimPrefix(m, "${"), "}")
		return os.Getenv(key)
	})
	// 不强制要求 env 必须存在；校验阶段会检查必填项
	return out, nil
}

func applyDefaults(cfg *Config) {
	if cfg.Archive.Driver == "" {
		cfg.Archive.Driver = "sqlite"
	}
	if cfg.Pipeline.Mode == "" {
		cfg.Pipeline.Mode = "once"
	}
	if cfg.Pipeline.Concurrency <= 0 {
		cfg.Pipeline.Concurrency = 1
	}
	if cfg.Pipeline.MaxPushPerRun <= 0 {
		cfg.Pipeline.MaxPushPerRun = 30
	}
	if cfg.Pipeline.DefaultPushPolicy == "" {
		cfg.Pipeline.DefaultPushPolicy = "push"
	}
	if cfg.Pipeline.DropIfPublishedBeforeDays <= 0 {
		cfg.Pipeline.DropIfPublishedBeforeDays = 30
	}
	// sources 默认 enabled=true
	for i := range cfg.Sources {
		if cfg.Sources[i].Type == "" {
			cfg.Sources[i].Type = "rss"
		}
		// 如果 YAML 未显式写 enabled，会是 false，这里统一设为 true
		// 允许用户明确写 enabled:false
		if _, ok := cfg.Sources[i].Notes["_enabled_set"]; ok {
			// no-op（保留给未来更精确的解析方式）
		}
		if cfg.Sources[i].Enabled == false {
			// 这里无法区分“未写”和“显式 false”，所以 MVP 先约定：示例配置都写 enabled
			// 生产实现时建议用 *bool 来区分。
		}
		if cfg.Sources[i].FingerprintFields == nil || len(cfg.Sources[i].FingerprintFields) == 0 {
			cfg.Sources[i].FingerprintFields = []string{"url"}
		}
	}
}
