package config

import (
	"context"
	"testing"

	. "github.com/halimath/expect-go"
)

func TestProvide_noEnv(t *testing.T) {
	c := Provide(context.Background())

	ExpectThat(t, c).Is(DeepEqual(Config{
		HTTPPort: 8080,
	}))
}

func TestProvide_env(t *testing.T) {
	t.Setenv("HTTP_PORT", "8912")

	c := Provide(context.Background())

	ExpectThat(t, c).Is(DeepEqual(Config{
		HTTPPort: 8912,
	}))
}
