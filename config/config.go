package config

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type NotifierConfig struct {
	WebhookURL string `yaml:"webhook_url"`
	Username   string `yaml:"username"`
	AvatarURL  string `yaml:"avatar_url"`
}

type ProductID struct {
	ID    int64  `yaml:"id"`
	Label string `yaml:"label"`
}

type MonitorConfig struct {
	Name        string      `yaml:"name"`
	Type        string      `yaml:"type"`
	EndpointURL string      `yaml:"endpoint_url"`
	ProductIDs  []ProductID `yaml:"product_ids"`
	Enabled     bool        `yaml:"enabled"`
}

type SiteConfig struct {
	Name     string          `yaml:"name"`
	URL      string          `yaml:"site_url"`
	Monitors []MonitorConfig `yaml:"monitors"`
	Notifier NotifierConfig  `yaml:"notifier"`
	Enabled  bool            `yaml:"enabled"`
}

type Config struct {
	GlobalPollInterval  time.Duration `yaml:"global_poll_interval"`
	GlobalRetryInterval time.Duration `yaml:"global_retry_interval"`
	GlobalMaxRetries    int           `yaml:"global_max_retries"`
	Sites               []SiteConfig  `yaml:"sites"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true) // catch typos in keys

	var cfg Config
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("YAML PARSE ERROR: %w", err)
	}

	validateConfig(&cfg)

	return &cfg, nil
}

func validateConfig(cfg *Config) {
	if cfg.GlobalPollInterval <= 0 {
		panic("CONFIG ERROR: GLOBAL_POLL_INTERVAL MUST BE > 0")
	}

	if cfg.GlobalRetryInterval <= 0 {
		panic("CONFIG ERROR: GLOBAL_RETRY_INTERVAL MUST BE > 0")
	}

	if cfg.GlobalMaxRetries <= 0 {
		panic("CONFIG ERROR: GLOBAL_MAX_RETRIES MUST BE > 0")
	}

	if len(cfg.Sites) == 0 {
		panic("CONFIG ERROR: NO SITES DEFINED")
	}

	for i, site := range cfg.Sites {
		if !site.Enabled {
			continue // skip disabled sites
		}

		if strings.TrimSpace(site.Name) == "" {
			panic(fmt.Sprintf("CONFIG ERROR: SITE #%d HAS NO NAME", i))
		}

		if strings.TrimSpace(site.URL) == "" || !strings.HasPrefix(site.URL, "https://") {
			panic(fmt.Sprintf("CONFIG ERROR: SITE %q HAS INVALID OR MISSING URL", site.Name))
		}

		if site.Notifier.WebhookURL == "" {
			panic(fmt.Sprintf("CONFIG ERROR: SITE %q HAS NO WEBHOOK_URL", site.Name))
		}

		if len(site.Monitors) == 0 {
			panic(fmt.Sprintf("CONFIG ERROR: SITE %q HAS NO MONITORS", site.Name))
		}

		for _, mon := range site.Monitors {
			if !mon.Enabled {
				continue
			}

			if strings.TrimSpace(mon.Name) == "" {
				panic(fmt.Sprintf("CONFIG ERROR: A MONITOR IN SITE %q HAS NO NAME", site.Name))
			}

			if strings.TrimSpace(mon.EndpointURL) == "" {
				panic(fmt.Sprintf("CONFIG ERROR: MONITOR %q IN SITE %q HAS NO ENDPOINT_URL", mon.Name, site.Name))
			}

			if len(mon.ProductIDs) == 0 {
				panic(fmt.Sprintf("CONFIG ERROR: MONITOR %q IN SITE %q HAS NO PRODUCT_IDS", mon.Name, site.Name))
			}
		}
	}
}
