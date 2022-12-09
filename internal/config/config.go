// Package config provides types and functions to handle application config
// set via env variables.
package config

import (
	"context"

	"github.com/sethvargo/go-envconfig"
)

// Config defines the main configuration values.
type Config struct {
	// The TCP port to listen on for HTTP connections
	HTTPPort int `env:"HTTP_PORT,default=8080"`
}

// Provide provides the application's Config by applying default and env values.
// This function panics in case an error occurs during config processing.
func Provide(ctx context.Context) Config {
	var c Config
	if err := envconfig.Process(ctx, &c); err != nil {
		panic(err)
	}
	return c
}
