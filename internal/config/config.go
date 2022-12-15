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
	// TADO configuration
	Tado TadoConfig
	//the database configuration
	DB Database
}

type TadoConfig struct {
	Username     string `env:"TADO_USERNAME,default="`
	Password     string `env:"TADO_PASSWORD,default="`
	ClientID     string `env:"TADO_CLIENT_ID,default=tado-web-app"`
	ClientSecret string `env:"TADO_CLIENT_ID,default=wZaRN7rpjn3FoNyF5IFuxg9uMzYJcvOoQ8QWiIqS3hfk6gLhVlG57j5YNoZL2Rtc"`
}

type Database struct {
	User     string `default:"user"`
	Password string `default:"password"`
	File     string `default:":memory:"`
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
