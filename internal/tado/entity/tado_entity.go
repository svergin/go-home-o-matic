package entity

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

const (
	HeatingModeDefault = iota
	HeatingModeVacancy
)

type HeatingScheduleConfiguration struct {
	ID          string
	Zone        string
	HeatingMode int
	JSONPayload string
}

type HeatingScheduleConfigRepo interface {
	GetHeatingSchedule(string, int) (string, error)
	SaveHeatingSchedule(string, string, int) error
}

type heatingScheduleConfigRepoImpl struct {
	db *sql.DB
}

var _ HeatingScheduleConfigRepo = &heatingScheduleConfigRepoImpl{}

func (repo *heatingScheduleConfigRepoImpl) GetHeatingSchedule(zone string, mode int) (string, error) {
	var result HeatingScheduleConfiguration

	row := repo.db.QueryRow("SELECT * FROM heatingscheduleconfiguration WHERE zone = ? AND heatingmode = ?", zone, mode)
	if err := row.Scan(&result.ID, &result.Zone, &result.HeatingMode, &result.JSONPayload); err != nil {
		return "", fmt.Errorf("error GetHeatingSchedule (%q,%q): %v", zone, mode, err)
	}

	return result.JSONPayload, nil
}

func (repo *heatingScheduleConfigRepoImpl) SaveHeatingSchedule(jsonpayload string, zone string, mode int) error {
	_, err := repo.db.Exec("INSERT INTO heatingscheduleconfiguration (id, zone, heatingmode,jsonpayload) VALUES (?,?,?,?)", getId(), zone, mode, jsonpayload)
	if err != nil {
		return err
	}

	return nil
}

func getId() string {
	return uuid.New().String()
}

func Provide(db *sql.DB) heatingScheduleConfigRepoImpl {
	setupTable(db)
	return heatingScheduleConfigRepoImpl{
		db: db,
	}
}

func setupTable(db *sql.DB) error {
	_, err := db.Exec(`
	CREATE TABLE heatingscheduleconfiguration (
		id VARCHAR(128) NOT NULL, 
		zone VARCHAR(128) NOT NULL, 
		heatingmode INT NOT NULL,
		jsonpayload VARCHAR(512) NOT NULL,
		PRIMARY KEY ('id'))`)
	if err != nil {
		return err
	}
	return nil
}
