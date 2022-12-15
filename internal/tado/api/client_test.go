package api

import (
	"context"
	"testing"

	"github.com/gonzolino/gotado/v2"
	. "github.com/halimath/expect-go"
	"github.com/svergin/go-home-o-matic/internal/config"
)

func TestGetUserinfo(t *testing.T) {
	ctx := context.Background()
	cfg := config.Provide(ctx)
	apiclient := Provide(cfg)
	user, err := apiclient.GetUser(ctx)
	if err != nil {
		t.Fatal("Test failed due to: &v", err)
	}
	ExpectThat(t, user).Is(DeepEqual(gotado.User{}))
}
