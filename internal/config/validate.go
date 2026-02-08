package config

import (
	"fmt"
)

func (c *Config) Validate() error {
	if c.Archive.Driver != "sqlite" {
		return fmt.Errorf("archive.driver must be sqlite (got %q)", c.Archive.Driver)
	}
	if c.Archive.Path == "" {
		return fmt.Errorf("archive.path is required")
	}
	if c.Notifier.WeCom.WebhookURL == "" {
		return fmt.Errorf("notifier.wecom.webhook_url is required (env ${WECOM_WEBHOOK_URL}?)")
	}

	if c.Pipeline.Mode != "once" && c.Pipeline.Mode != "daemon" {
		return fmt.Errorf("pipeline.mode must be once|daemon (got %q)", c.Pipeline.Mode)
	}

	seen := map[string]struct{}{}
	enabledCount := 0
	for _, s := range c.Sources {
		if s.SourceID == "" {
			return fmt.Errorf("sources[].source_id is required")
		}
		if _, ok := seen[s.SourceID]; ok {
			return fmt.Errorf("duplicate source_id: %s", s.SourceID)
		}
		seen[s.SourceID] = struct{}{}

		if s.Type != "rss" {
			return fmt.Errorf("source %s: unsupported type %q (only rss for MVP)", s.SourceID, s.Type)
		}
		if s.URL == "" {
			return fmt.Errorf("source %s: url is required", s.SourceID)
		}
		if s.Enabled {
			enabledCount++
		}
	}

	if enabledCount == 0 {
		return fmt.Errorf("no enabled sources configured")
	}
	return nil
}
