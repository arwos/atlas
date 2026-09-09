/*
 *  Copyright (c) 2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package dhcp

import (
	"fmt"
	"net/netip"
	"strings"
	"time"
)

type ConfigGroup struct {
	Config Config `yaml:"dhcp"`
}

type Config struct {
	Bootstrap BootstrapConfig `yaml:"bootstrap"`
}

type BootstrapConfig struct {
	Interface    string `yaml:"interface"`
	CIDR         string `yaml:"cidr"`
	LeaseSeconds int64  `yaml:"lease_seconds"`
}

func (c *Config) Default() {
	if c.Bootstrap.LeaseSeconds == 0 {
		c.Bootstrap.LeaseSeconds = int64(time.Hour.Seconds())
	}
}

func (c *Config) Validate() error {
	if strings.TrimSpace(c.Bootstrap.Interface) == "" {
		return fmt.Errorf("dhcp: bootstrap.interface is required")
	}
	if _, err := netip.ParsePrefix(c.Bootstrap.CIDR); err != nil {
		return fmt.Errorf("dhcp: invalid bootstrap.cidr: %w", err)
	}
	if c.Bootstrap.LeaseSeconds <= 0 {
		return fmt.Errorf("dhcp: bootstrap.lease_seconds must be positive")
	}
	return nil
}
