/*
 *  Copyright (c) 2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

// Package dhcp manages DHCP configuration and lifecycle.
package dhcp

import (
	"errors"
	"fmt"
	"net/netip"
	"strings"
	"time"
)

// ConfigGroup contains DHCP plugin configuration.
type ConfigGroup struct {
	Config Config `yaml:"dhcp"`
}

// Config contains DHCP settings.
type Config struct {
	Bootstrap BootstrapConfig `yaml:"bootstrap"`
}

// BootstrapConfig defines the initial DHCP subnet.
type BootstrapConfig struct {
	Interface    string `yaml:"interface"`
	CIDR         string `yaml:"cidr"`
	LeaseSeconds int64  `yaml:"lease_seconds"`
}

// Default applies DHCP configuration defaults.
func (c *Config) Default() {
	if c.Bootstrap.LeaseSeconds == 0 {
		c.Bootstrap.LeaseSeconds = int64(time.Hour.Seconds())
	}
}

// Validate verifies DHCP configuration.
func (c *Config) Validate() error {
	if strings.TrimSpace(c.Bootstrap.Interface) == "" {
		return errors.New("dhcp: bootstrap.interface is required")
	}
	if _, err := netip.ParsePrefix(c.Bootstrap.CIDR); err != nil {
		return fmt.Errorf("dhcp: invalid bootstrap.cidr: %w", err)
	}
	if c.Bootstrap.LeaseSeconds <= 0 {
		return errors.New("dhcp: bootstrap.lease_seconds must be positive")
	}
	return nil
}
