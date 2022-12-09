// Package health provides some types and functions to implement liveness
// and readyness checks based on HTTP probes.
package health

import (
	"context"
	"net/http"
	"sync"

	"github.com/halimath/kvlog"
	"golang.org/x/sync/errgroup"
)

// Check defines the interface for custom readyness checks.
type Check interface {
	// Check is called to execute the check. Any non-nil return value
	// is considered a check failure incl. context deadlines.
	Check(context.Context) error
}

// CheckFunc is a convenience type to implement Check using a bare function.
type CheckFunc func(context.Context) error

func (f CheckFunc) Check(ctx context.Context) error { return f(ctx) }

// Handler implements liveness and readyness checking.
type Handler struct {
	checks []Check
	lock   sync.RWMutex
	mux    http.ServeMux
}

// AddCheck registers c as another readyness check.
func (h *Handler) AddCheck(c Check) {
	h.lock.Lock()
	defer h.lock.Unlock()

	h.checks = append(h.checks)
}

// ServeHTTP dispatches and executes health checks.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *Handler) handleLive(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) handleReady(w http.ResponseWriter, r *http.Request) {
	eg, ctx := errgroup.WithContext(r.Context())

	h.lock.RLock()

	for _, c := range h.checks {
		eg.Go(func() error { return c.Check(ctx) })
	}

	if err := eg.Wait(); err != nil {
		kvlog.L.Logs("read check failed", kvlog.WithErr(err))
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Provide creates a new Handler ready to use. The Handler must be
// mounted on some HTTP path (i.e. on a http.ServeMux) to receive
// requests.
func Provide() *Handler {
	h := &Handler{
		mux: *http.NewServeMux(),
	}

	h.mux.HandleFunc("/livez", h.handleLive)
	h.mux.HandleFunc("/readyz", h.handleReady)

	return h
}
