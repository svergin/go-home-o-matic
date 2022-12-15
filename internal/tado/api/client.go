package api

import (
	"context"
	"errors"

	"github.com/gonzolino/gotado/v2"
	"github.com/svergin/go-home-o-matic/internal/config"
)

type TadoAPIClient struct {
	tado *gotado.Tado
	cfg  config.Config
}

func Provide(cfg config.Config) *TadoAPIClient {
	return &TadoAPIClient{
		tado: gotado.New(cfg.Tado.ClientID, cfg.Tado.ClientSecret),
		cfg:  cfg,
	}

}

func (api *TadoAPIClient) GetUser(ctx context.Context) (*gotado.User, error) {
	if api.cfg.Tado.Username == "" {
		return nil, errors.New("username is missing")

	}
	if api.cfg.Tado.Password == "" {
		return nil, errors.New("password is missing")
	}

	user, err := api.tado.Me(ctx, api.cfg.Tado.Username, api.cfg.Tado.Password)
	if err != nil {
		return nil, err
	}
	return user, nil
}
