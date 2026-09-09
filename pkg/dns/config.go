//nolint:revive // Plugin configuration follows goppy's required method names.
package dns

import (
	"errors"
	"fmt"
	"net"
	"time"
)

type ConfigGroup struct {
	Config Config `yaml:"dns"`
}
type Config struct {
	Listen       string        `yaml:"listen"`
	CacheEntries int           `yaml:"cache_entries"`
	CacheMaxTTL  time.Duration `yaml:"cache_max_ttl"`
}

const defaultCacheMaxTTL = 24 * time.Hour

func (c *Config) Default() {
	if c.Listen == "" {
		c.Listen = ":53"
	}
	if c.CacheEntries == 0 {
		c.CacheEntries = 10000
	}
	if c.CacheMaxTTL == 0 {
		c.CacheMaxTTL = defaultCacheMaxTTL
	}
}

func (c *Config) Validate() error {
	if _, _, err := net.SplitHostPort(c.Listen); err != nil {
		return fmt.Errorf("dns: invalid listen address: %w", err)
	}
	if c.CacheEntries <= 0 || c.CacheMaxTTL <= 0 {
		return errors.New("dns: cache_entries and cache_max_ttl must be positive")
	}
	return nil
}
