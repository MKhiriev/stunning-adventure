// Package config provides flag-parsing utilities for server and agent binaries.
// It defines configuration defaults, a NetAddress type implementing flag.Value,
// and parsing helpers returning fully resolved configuration parameters.
package config

import "github.com/caarlos0/env/v11"

func parseEnv(cfg any) error {
	err := env.Parse(cfg)
	if err != nil {
		return err
	}

	return nil
}
