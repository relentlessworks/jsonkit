package config

import (
	"flag"
	"fmt"
	"os"
)

// Config holds all configuration for the jsonkit service.
type Config struct {
	Addr   string
	Port   int
	Secret string
}

// Default returns the default configuration.
func Default() *Config {
	return &Config{
		Addr:   "0.0.0.0",
		Port:   8484,
		Secret: "",
	}
}

// Load parses flags and environment variables, layering them over defaults.
// Order: defaults < env vars < flags.
func Load() *Config {
	c := Default()

	// Env vars
	if v := os.Getenv("JSONKIT_ADDR"); v != "" {
		c.Addr = v
	}
	if v := os.Getenv("JSONKIT_PORT"); v != "" {
		fmt.Sscanf(v, "%d", &c.Port)
	}
	if v := os.Getenv("JSONKIT_SECRET"); v != "" {
		c.Secret = v
	}

	// Flags (override env)
	flag.StringVar(&c.Addr, "addr", c.Addr, "listen address")
	flag.IntVar(&c.Port, "port", c.Port, "listen port")
	flag.StringVar(&c.Secret, "secret", c.Secret, "token signing secret (auto-generated if empty)")
	flag.Parse()

	return c
}

// ListenAddr returns the full listen address.
func (c *Config) ListenAddr() string {
	return fmt.Sprintf("%s:%d", c.Addr, c.Port)
}
