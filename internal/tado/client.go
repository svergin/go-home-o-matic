package tado

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gonzolino/gotado/v2"
	"github.com/svergin/go-home-o-matic/internal/config"
)

const home_name = "Home"

type Handler struct {
	mux http.ServeMux
	cfg config.Config
}

type interalError struct {
	msg    string
	status int
}

func (ie interalError) Error() string {
	return ie.msg
}

func Provide(cfg *config.Config) *Handler {
	h := &Handler{
		mux: *http.NewServeMux(),
		cfg: *cfg,
	}

	h.mux.HandleFunc("/tado/info", h.handleUserInfo)
	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *Handler) handleUserInfo(w http.ResponseWriter, r *http.Request) {

	user, err := getUserInfo(&h.cfg)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
	w.WriteHeader(http.StatusOK)
}

func getUserInfo(cfg *config.Config) (*gotado.User, error) {
	tc := gotado.New(cfg.Tado.ClientID, cfg.Tado.ClientSecret)
	ctx := context.Background()
	if cfg.Tado.Username == "" {
		return nil, interalError{
			msg:    "username is missing",
			status: http.StatusBadRequest,
		}
	}
	if cfg.Tado.Password == "" {
		return nil, interalError{
			msg:    "password is missing",
			status: http.StatusBadRequest,
		}
	}

	user, err := tc.Me(ctx, cfg.Tado.Username, cfg.Tado.Password)
	if err != nil {
		return nil, err
	}
	return user, nil
}

/*
client_id=tado-web-app
grant_type=password
scope=home.user
username="vergin@gmx.net"
password="XXX"
client_secret=wZaRN7rpjn3FoNyF5IFuxg9uMzYJcvOoQ8QWiIqS3hfk6gLhVlG57j5YNoZL2Rtc
*/
