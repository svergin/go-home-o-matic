package tado

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gonzolino/gotado/v2"
	"github.com/halimath/kvlog"
	"github.com/svergin/go-home-o-matic/internal/config"
)

type Handler struct {
	mux http.ServeMux
	cfg config.Config
}

type interalError struct {
	msg    string
	status int
}

const (
	heating_mode_normal = iota
	heating_mode_vacancy
)

const (
	tadohome_name          = "Home"
	tadozone_wohnzimmer    = "Wohnzimmer"
	tadozone_kinderzimmer1 = "Kinderzimmer 1"
	tadozone_kinderzimmer2 = "Kinderzimmer 2"
	tadozone_badezimmer    = "Badezimmer"
)

/*
TODO: Logger einbauen
TODO: Fehlerbehandlung
TODO: Konfigurationen in der DB speichern
TODO: Architektur-Muster anwenden
TODO: REST-Controller ggf. generieren
TODO: UI einbauen

*/

func (ie interalError) Error() string {
	return ie.msg
}

func Provide(cfg *config.Config) *Handler {
	h := &Handler{
		mux: *http.NewServeMux(),
		cfg: *cfg,
	}

	h.mux.HandleFunc("/info", h.handleUserInfo)
	h.mux.HandleFunc("/schedule", h.handleHeatingSchedule)
	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *Handler) handleUserInfo(w http.ResponseWriter, r *http.Request) {

	user, err := getUser(&h.cfg)
	if err != nil {
		kvlog.L.Logs("could not get tado user", kvlog.WithErr(err))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) handleHeatingSchedule(w http.ResponseWriter, r *http.Request) {
	paramMode := r.URL.Query().Get("mode")
	paramZone := r.URL.Query().Get("zone")
	if paramMode == "" || paramZone == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("parameters 'mode' and 'zone' must be supplied"))
		return
	}
	if !(paramMode == "n" || paramMode == "v") {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("parameter 'mode' has unsupported value"))
		return
	}
	if !(paramZone == "w" || paramZone == "b" || paramZone == "k1" || paramZone == "k2") {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("parameter 'zone' has unsupported value"))
		return
	}

	user, err := getUser(&h.cfg)
	if err != nil {
		kvlog.L.Logs("could not get tado user", kvlog.WithErr(err))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	ctx := context.Background()

	err = applyHeatingSchedule(ctx, user, toZone(paramZone), toMode(paramMode))

	if err != nil {
		kvlog.L.Logs("could not apply heating schedule", kvlog.WithErr(err))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.Write([]byte(fmt.Sprintf("heating schedule mode '%v' for zone '%v' applied successfully", paramMode, paramZone)))
	w.WriteHeader(http.StatusOK)
}

func toMode(mode string) int {
	switch mode {
	case "n":
		return heating_mode_normal
	case "v":
		return heating_mode_vacancy
	default:
		panic(fmt.Sprintf("unknown heating mode: %v", mode))
	}
}

func toZone(zone string) string {
	switch zone {
	case "b":
		return tadozone_badezimmer
	case "w":
		return tadozone_wohnzimmer
	case "k1":
		return tadozone_kinderzimmer1
	case "k2":
		return tadozone_kinderzimmer2
	default:
		panic(fmt.Sprintf("unknown zone: %v", zone))
	}
}

func applyHeatingSchedule(ctx context.Context, user *gotado.User, tadozone string, mode int) error {
	home, err := user.GetHome(ctx, tadohome_name)
	if err != nil {
		return err
	}
	zone, err := home.GetZone(ctx, tadozone)
	if err != nil {
		return err
	}
	kvlog.L.Logf("%v was triggered", tadozone)
	switch tadozone {
	case tadozone_wohnzimmer:
		return applyForWohnzimmer(ctx, mode, zone)
	case tadozone_badezimmer:
		return applyForBadezimmer(ctx, mode, zone)
	case tadozone_kinderzimmer1:
		return applyForKinderzimmer1(ctx, mode, zone)
	case tadozone_kinderzimmer2:
		return applyForKinderzimmer2(ctx, mode, zone)
	default:
		return nil

	}

}

func applyForWohnzimmer(ctx context.Context, mode int, zone *gotado.Zone) error {
	switch mode {
	case heating_mode_vacancy:
		newhs, err := zone.ScheduleMonToSun(ctx)
		if err != nil {
			return err
		}
		newhs.
			NewTimeBlock(ctx, gotado.DayTypeMondayToSunday, "00:00", "08:00", true, gotado.PowerOn, 18.0).
			AddTimeBlock(ctx, gotado.DayTypeMondayToSunday, "08:00", "21:30", true, gotado.PowerOn, 21.0)
		zone.SetHeatingSchedule(ctx, newhs)
	case heating_mode_normal:
		newhs, err := zone.ScheduleAllDays(ctx)
		if err != nil {
			return err
		}
		newhs.
			NewTimeBlock(ctx, gotado.DayTypeMonday, "00:00", "06:15", true, gotado.PowerOn, 18.0).
			AddTimeBlock(ctx, gotado.DayTypeMonday, "06:15", "07:00", true, gotado.PowerOn, 21.0).
			AddTimeBlock(ctx, gotado.DayTypeMonday, "07:00", "16:00", true, gotado.PowerOn, 18.0).
			AddTimeBlock(ctx, gotado.DayTypeMonday, "16:00", "21:15", true, gotado.PowerOn, 21.0).
			AddTimeBlock(ctx, gotado.DayTypeMonday, "21:15", "00:00", true, gotado.PowerOn, 18.0).
			AddTimeBlock(ctx, gotado.DayTypeTuesday, "00:00", "06:15", true, gotado.PowerOn, 18.0).
			AddTimeBlock(ctx, gotado.DayTypeTuesday, "06:15", "07:00", true, gotado.PowerOn, 21.0).
			AddTimeBlock(ctx, gotado.DayTypeTuesday, "07:00", "16:00", true, gotado.PowerOn, 18.0).
			AddTimeBlock(ctx, gotado.DayTypeTuesday, "16:00", "21:15", true, gotado.PowerOn, 21.0).
			AddTimeBlock(ctx, gotado.DayTypeTuesday, "21:15", "00:00", true, gotado.PowerOn, 18.0).
			AddTimeBlock(ctx, gotado.DayTypeWednesday, "00:00", "06:15", true, gotado.PowerOn, 18.0).
			AddTimeBlock(ctx, gotado.DayTypeWednesday, "06:15", "07:00", true, gotado.PowerOn, 21.0).
			AddTimeBlock(ctx, gotado.DayTypeWednesday, "07:00", "16:00", true, gotado.PowerOn, 18.0).
			AddTimeBlock(ctx, gotado.DayTypeWednesday, "16:00", "21:15", true, gotado.PowerOn, 21.0).
			AddTimeBlock(ctx, gotado.DayTypeWednesday, "21:15", "00:00", true, gotado.PowerOn, 18.0).
			AddTimeBlock(ctx, gotado.DayTypeThursday, "00:00", "06:15", true, gotado.PowerOn, 18.0).
			AddTimeBlock(ctx, gotado.DayTypeThursday, "06:15", "07:00", true, gotado.PowerOn, 21.0).
			AddTimeBlock(ctx, gotado.DayTypeThursday, "07:00", "16:00", true, gotado.PowerOn, 18.0).
			AddTimeBlock(ctx, gotado.DayTypeThursday, "16:00", "21:15", true, gotado.PowerOn, 21.0).
			AddTimeBlock(ctx, gotado.DayTypeThursday, "21:15", "00:00", true, gotado.PowerOn, 18.0).
			AddTimeBlock(ctx, gotado.DayTypeFriday, "00:00", "06:15", true, gotado.PowerOn, 18.0).
			AddTimeBlock(ctx, gotado.DayTypeFriday, "06:15", "07:00", true, gotado.PowerOn, 21.0).
			AddTimeBlock(ctx, gotado.DayTypeFriday, "07:00", "16:00", true, gotado.PowerOn, 20.0).
			AddTimeBlock(ctx, gotado.DayTypeFriday, "16:00", "21:15", true, gotado.PowerOn, 21.0).
			AddTimeBlock(ctx, gotado.DayTypeFriday, "21:15", "00:00", true, gotado.PowerOn, 18.0).
			AddTimeBlock(ctx, gotado.DayTypeSaturday, "00:00", "08:00", true, gotado.PowerOn, 18.0).
			AddTimeBlock(ctx, gotado.DayTypeSaturday, "08:00", "21:15", true, gotado.PowerOn, 21.0).
			AddTimeBlock(ctx, gotado.DayTypeSaturday, "21:15", "00:00", true, gotado.PowerOn, 18.0).
			AddTimeBlock(ctx, gotado.DayTypeSunday, "00:00", "08:00", true, gotado.PowerOn, 18.0).
			AddTimeBlock(ctx, gotado.DayTypeSunday, "08:00", "21:15", true, gotado.PowerOn, 21.0).
			AddTimeBlock(ctx, gotado.DayTypeSunday, "21:15", "00:00", true, gotado.PowerOn, 18.0)

		zone.SetHeatingSchedule(ctx, newhs)
	}
	return nil
}

func applyForBadezimmer(ctx context.Context, mode int, zone *gotado.Zone) error {
	switch mode {
	case heating_mode_vacancy:
		newhs, err := zone.ScheduleMonToSun(ctx)
		if err != nil {
			return err
		}
		newhs.
			NewTimeBlock(ctx, gotado.DayTypeMondayToSunday, "00:00", "07:45", true, gotado.PowerOn, 18.0).
			AddTimeBlock(ctx, gotado.DayTypeMondayToSunday, "07:45", "09:00", true, gotado.PowerOn, 21.0).
			AddTimeBlock(ctx, gotado.DayTypeMondayToSunday, "09:00", "18:00", true, gotado.PowerOn, 18.0).
			AddTimeBlock(ctx, gotado.DayTypeMondayToSunday, "18:00", "21:00", true, gotado.PowerOn, 21.0).
			AddTimeBlock(ctx, gotado.DayTypeMondayToSunday, "21:00", "00:00", true, gotado.PowerOn, 18.0)
		zone.SetHeatingSchedule(ctx, newhs)
	case heating_mode_normal:
		newhs, err := zone.ScheduleMonToFriSatSun(ctx)
		if err != nil {
			return err
		}
		newhs.
			NewTimeBlock(ctx, gotado.DayTypeMondayToFriday, "00:00", "06:00", true, gotado.PowerOn, 18.0).
			AddTimeBlock(ctx, gotado.DayTypeMondayToFriday, "06:00", "07:00", true, gotado.PowerOn, 21.0).
			AddTimeBlock(ctx, gotado.DayTypeMondayToFriday, "07:00", "17:30", true, gotado.PowerOn, 18.0).
			AddTimeBlock(ctx, gotado.DayTypeMondayToFriday, "17:30", "20:00", true, gotado.PowerOn, 21.0).
			AddTimeBlock(ctx, gotado.DayTypeMondayToFriday, "20:00", "00:00", true, gotado.PowerOn, 18.0).
			AddTimeBlock(ctx, gotado.DayTypeSaturday, "00:00", "07:45", true, gotado.PowerOn, 18.0).
			AddTimeBlock(ctx, gotado.DayTypeSaturday, "07:45", "09:00", true, gotado.PowerOn, 21.0).
			AddTimeBlock(ctx, gotado.DayTypeSaturday, "09:00", "18:00", true, gotado.PowerOn, 18.0).
			AddTimeBlock(ctx, gotado.DayTypeSaturday, "18:00", "21:00", true, gotado.PowerOn, 21.0).
			AddTimeBlock(ctx, gotado.DayTypeSaturday, "21:00", "00:00", true, gotado.PowerOn, 18.0).
			AddTimeBlock(ctx, gotado.DayTypeSunday, "00:00", "07:45", true, gotado.PowerOn, 18.0).
			AddTimeBlock(ctx, gotado.DayTypeSunday, "07:45", "09:00", true, gotado.PowerOn, 21.0).
			AddTimeBlock(ctx, gotado.DayTypeSunday, "09:00", "18:00", true, gotado.PowerOn, 18.0).
			AddTimeBlock(ctx, gotado.DayTypeSunday, "18:00", "21:00", true, gotado.PowerOn, 21.0).
			AddTimeBlock(ctx, gotado.DayTypeSunday, "21:00", "00:00", true, gotado.PowerOn, 18.0)
		zone.SetHeatingSchedule(ctx, newhs)
	}
	return nil
}

func applyForKinderzimmer1(ctx context.Context, mode int, zone *gotado.Zone) error {
	switch mode {
	case heating_mode_vacancy:
		newhs, err := zone.ScheduleMonToSun(ctx)
		if err != nil {
			return err
		}
		newhs.
			NewTimeBlock(ctx, gotado.DayTypeMondayToSunday, "00:00", "08:30", true, gotado.PowerOff, 0.0).
			AddTimeBlock(ctx, gotado.DayTypeMondayToSunday, "08:30", "18:00", true, gotado.PowerOn, 20.0).
			AddTimeBlock(ctx, gotado.DayTypeMondayToSunday, "18:00", "00:00", true, gotado.PowerOff, 0.0)
		zone.SetHeatingSchedule(ctx, newhs)
	case heating_mode_normal:
		newhs, err := zone.ScheduleAllDays(ctx)
		if err != nil {
			return err
		}
		newhs.
			NewTimeBlock(ctx, gotado.DayTypeMonday, "00:00", "07:30", true, gotado.PowerOff, 0.0).
			AddTimeBlock(ctx, gotado.DayTypeMonday, "07:30", "08:00", true, gotado.PowerOn, 21.0).
			AddTimeBlock(ctx, gotado.DayTypeMonday, "08:00", "15:00", true, gotado.PowerOn, 18.0).
			AddTimeBlock(ctx, gotado.DayTypeMonday, "15:00", "18:00", true, gotado.PowerOn, 21.0).
			AddTimeBlock(ctx, gotado.DayTypeMonday, "18:00", "00:00", true, gotado.PowerOff, 0.0).
			AddTimeBlock(ctx, gotado.DayTypeTuesday, "00:00", "06:30", true, gotado.PowerOff, 0.0).
			AddTimeBlock(ctx, gotado.DayTypeTuesday, "06:30", "07:15", true, gotado.PowerOn, 21.0).
			AddTimeBlock(ctx, gotado.DayTypeTuesday, "07:15", "15:00", true, gotado.PowerOn, 18.0).
			AddTimeBlock(ctx, gotado.DayTypeTuesday, "15:00", "18:00", true, gotado.PowerOn, 21.0).
			AddTimeBlock(ctx, gotado.DayTypeTuesday, "18:00", "00:00", true, gotado.PowerOff, 0.0).
			AddTimeBlock(ctx, gotado.DayTypeWednesday, "00:00", "06:30", true, gotado.PowerOff, 0.0).
			AddTimeBlock(ctx, gotado.DayTypeWednesday, "06:30", "07:15", true, gotado.PowerOn, 21.0).
			AddTimeBlock(ctx, gotado.DayTypeWednesday, "07:15", "15:00", true, gotado.PowerOn, 18.0).
			AddTimeBlock(ctx, gotado.DayTypeWednesday, "15:00", "18:00", true, gotado.PowerOn, 21.0).
			AddTimeBlock(ctx, gotado.DayTypeWednesday, "18:00", "00:00", true, gotado.PowerOff, 0.0).
			AddTimeBlock(ctx, gotado.DayTypeThursday, "00:00", "06:30", true, gotado.PowerOff, 0.0).
			AddTimeBlock(ctx, gotado.DayTypeThursday, "06:30", "07:15", true, gotado.PowerOn, 21.0).
			AddTimeBlock(ctx, gotado.DayTypeThursday, "07:15", "15:00", true, gotado.PowerOn, 18.0).
			AddTimeBlock(ctx, gotado.DayTypeThursday, "15:00", "18:00", true, gotado.PowerOn, 21.0).
			AddTimeBlock(ctx, gotado.DayTypeThursday, "18:00", "00:00", true, gotado.PowerOff, 0.0).
			AddTimeBlock(ctx, gotado.DayTypeFriday, "00:00", "06:30", true, gotado.PowerOff, 0.0).
			AddTimeBlock(ctx, gotado.DayTypeFriday, "06:30", "07:15", true, gotado.PowerOn, 21.0).
			AddTimeBlock(ctx, gotado.DayTypeFriday, "07:15", "15:00", true, gotado.PowerOn, 18.0).
			AddTimeBlock(ctx, gotado.DayTypeFriday, "15:00", "18:00", true, gotado.PowerOn, 21.0).
			AddTimeBlock(ctx, gotado.DayTypeFriday, "18:00", "00:00", true, gotado.PowerOff, 0.0).
			AddTimeBlock(ctx, gotado.DayTypeSaturday, "00:00", "08:30", true, gotado.PowerOff, 0.0).
			AddTimeBlock(ctx, gotado.DayTypeSaturday, "08:30", "19:00", true, gotado.PowerOn, 20.0).
			AddTimeBlock(ctx, gotado.DayTypeSaturday, "19:00", "00:00", true, gotado.PowerOff, 0.0).
			AddTimeBlock(ctx, gotado.DayTypeSunday, "00:00", "08:30", true, gotado.PowerOff, 0.0).
			AddTimeBlock(ctx, gotado.DayTypeSunday, "08:30", "18:00", true, gotado.PowerOn, 20.0).
			AddTimeBlock(ctx, gotado.DayTypeSunday, "18:00", "00:00", true, gotado.PowerOff, 0.0)
		zone.SetHeatingSchedule(ctx, newhs)
	}
	return nil
}

func applyForKinderzimmer2(ctx context.Context, mode int, zone *gotado.Zone) error {
	switch mode {
	case heating_mode_vacancy:
		newhs, err := zone.ScheduleMonToSun(ctx)
		if err != nil {
			return err
		}
		newhs.
			NewTimeBlock(ctx, gotado.DayTypeMondayToSunday, "00:00", "08:30", true, gotado.PowerOff, 0.0).
			AddTimeBlock(ctx, gotado.DayTypeMondayToSunday, "08:30", "18:00", true, gotado.PowerOn, 20.0).
			AddTimeBlock(ctx, gotado.DayTypeMondayToSunday, "18:00", "00:00", true, gotado.PowerOff, 0.0)
		zone.SetHeatingSchedule(ctx, newhs)
	case heating_mode_normal:
		newhs, err := zone.ScheduleAllDays(ctx)
		if err != nil {
			return err
		}
		newhs.
			NewTimeBlock(ctx, gotado.DayTypeMonday, "00:00", "06:30", true, gotado.PowerOff, 0.0).
			AddTimeBlock(ctx, gotado.DayTypeMonday, "06:30", "07:15", true, gotado.PowerOn, 21.0).
			AddTimeBlock(ctx, gotado.DayTypeMonday, "07:15", "15:00", true, gotado.PowerOn, 18.0).
			AddTimeBlock(ctx, gotado.DayTypeMonday, "15:00", "18:00", true, gotado.PowerOn, 21.0).
			AddTimeBlock(ctx, gotado.DayTypeMonday, "18:00", "00:00", true, gotado.PowerOff, 0.0).
			AddTimeBlock(ctx, gotado.DayTypeTuesday, "00:00", "06:30", true, gotado.PowerOff, 0.0).
			AddTimeBlock(ctx, gotado.DayTypeTuesday, "06:30", "07:15", true, gotado.PowerOn, 21.0).
			AddTimeBlock(ctx, gotado.DayTypeTuesday, "07:15", "15:00", true, gotado.PowerOn, 18.0).
			AddTimeBlock(ctx, gotado.DayTypeTuesday, "15:00", "18:00", true, gotado.PowerOn, 21.0).
			AddTimeBlock(ctx, gotado.DayTypeTuesday, "18:00", "00:00", true, gotado.PowerOff, 0.0).
			AddTimeBlock(ctx, gotado.DayTypeWednesday, "00:00", "06:30", true, gotado.PowerOff, 0.0).
			AddTimeBlock(ctx, gotado.DayTypeWednesday, "06:30", "07:15", true, gotado.PowerOn, 21.0).
			AddTimeBlock(ctx, gotado.DayTypeWednesday, "07:15", "15:00", true, gotado.PowerOn, 18.0).
			AddTimeBlock(ctx, gotado.DayTypeWednesday, "15:00", "18:00", true, gotado.PowerOn, 21.0).
			AddTimeBlock(ctx, gotado.DayTypeWednesday, "18:00", "00:00", true, gotado.PowerOff, 0.0).
			AddTimeBlock(ctx, gotado.DayTypeThursday, "00:00", "06:30", true, gotado.PowerOff, 0.0).
			AddTimeBlock(ctx, gotado.DayTypeThursday, "06:30", "07:15", true, gotado.PowerOn, 21.0).
			AddTimeBlock(ctx, gotado.DayTypeThursday, "07:15", "15:00", true, gotado.PowerOn, 18.0).
			AddTimeBlock(ctx, gotado.DayTypeThursday, "15:00", "18:00", true, gotado.PowerOn, 21.0).
			AddTimeBlock(ctx, gotado.DayTypeThursday, "18:00", "00:00", true, gotado.PowerOff, 0.0).
			AddTimeBlock(ctx, gotado.DayTypeFriday, "00:00", "07:30", true, gotado.PowerOff, 0.0).
			AddTimeBlock(ctx, gotado.DayTypeFriday, "07:30", "08:00", true, gotado.PowerOn, 21.0).
			AddTimeBlock(ctx, gotado.DayTypeFriday, "08:00", "15:00", true, gotado.PowerOn, 18.0).
			AddTimeBlock(ctx, gotado.DayTypeFriday, "15:00", "18:00", true, gotado.PowerOn, 21.0).
			AddTimeBlock(ctx, gotado.DayTypeFriday, "18:00", "00:00", true, gotado.PowerOff, 0.0).
			AddTimeBlock(ctx, gotado.DayTypeSaturday, "00:00", "08:30", true, gotado.PowerOff, 0.0).
			AddTimeBlock(ctx, gotado.DayTypeSaturday, "08:30", "18:00", true, gotado.PowerOn, 20.0).
			AddTimeBlock(ctx, gotado.DayTypeSaturday, "18:00", "00:00", true, gotado.PowerOff, 0.0).
			AddTimeBlock(ctx, gotado.DayTypeSunday, "00:00", "08:30", true, gotado.PowerOff, 0.0).
			AddTimeBlock(ctx, gotado.DayTypeSunday, "08:30", "18:00", true, gotado.PowerOn, 20.0).
			AddTimeBlock(ctx, gotado.DayTypeSunday, "18:00", "00:00", true, gotado.PowerOff, 0.0)
		zone.SetHeatingSchedule(ctx, newhs)
	}
	return nil
}

func getUser(cfg *config.Config) (*gotado.User, error) {
	ctx := context.Background()
	tc, err := prepareTadoClient(ctx, cfg)
	if err != nil {
		return nil, err
	}

	user, err := tc.Me(ctx, cfg.Tado.Username, cfg.Tado.Password)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func prepareTadoClient(ctx context.Context, cfg *config.Config) (*gotado.Tado, error) {
	tc := gotado.New(cfg.Tado.ClientID, cfg.Tado.ClientSecret)
	if cfg.Tado.Username == "" {
		return nil, interalError{
			msg:    "username is missing",
			status: http.StatusUnauthorized,
		}
	}
	if cfg.Tado.Password == "" {
		return nil, interalError{
			msg:    "password is missing",
			status: http.StatusUnauthorized,
		}
	}
	return tc, nil
}

/*
client_id=tado-web-app
grant_type=password
scope=home.user
username="vergin@gmx.net"
password="XXX"
client_secret=wZaRN7rpjn3FoNyF5IFuxg9uMzYJcvOoQ8QWiIqS3hfk6gLhVlG57j5YNoZL2Rtc
*/
