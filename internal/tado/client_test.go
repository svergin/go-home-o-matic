package tado

import (
	"context"
	"testing"

	"github.com/gonzolino/gotado/v2"
	. "github.com/halimath/expect-go"
	"github.com/svergin/go-home-o-matic/internal/config"
)

func TestGetUserinfo(t *testing.T) {
	cfg := config.Provide(context.Background())
	user, err := getUserInfo(&cfg)
	if err != nil {
		t.Fatal("Test failed due to: &v", err)
	}
	ExpectThat(t, user).Is(DeepEqual(gotado.User{}))
}
