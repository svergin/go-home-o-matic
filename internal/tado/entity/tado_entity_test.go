package entity

import (
	"context"
	"testing"

	. "github.com/halimath/expect-go"
	"github.com/svergin/go-home-o-matic/internal/config"
	"github.com/svergin/go-home-o-matic/internal/database"
)

func TestStoreAndLoadHeatingScheduleConfig_Success(t *testing.T) {
	ctx := context.Background()
	cfg := config.Provide(ctx)
	db := database.Provide(&cfg)
	repo := Provide(db)

	err := repo.SaveHeatingSchedule("invalid json", "Wohnzimmer", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	jsonstring, err := repo.GetHeatingSchedule("Wohnzimmer", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ExpectThat(t, jsonstring).Is(DeepEqual("invalid json"))
}

func TestLoadHeatingScheduleConfig_Failed(t *testing.T) {
	ctx := context.Background()
	cfg := config.Provide(ctx)
	db := database.Provide(&cfg)
	repo := Provide(db)
	_, err := repo.GetHeatingSchedule("Wohnzimmer", 0)
	ExpectThat(t, err.Error()).Is(StringContaining("sql: no rows in result set"))

}
