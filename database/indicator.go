package database

import (
	"context"
	"database/sql"
	"github.com/zacscoding/gochecker"
)

const (
	MySQLValidationQuery   = "SELECT 1"
	MySQLVersionQuery      = "SELECT VERSION()"
	Sqlite3ValidationQuery = "SELECT 1"
	Sqlite3VersionQuery    = "SELECT sqlite_version()"
)

// Indicator is conforms of gochecker.Indicator to check database health status.
type Indicator struct {
	db             *sql.DB
	driverName     string
	validatorQuery string
	versionQuery   string
}

func (i *Indicator) Health(ctx context.Context) gochecker.ComponentStatus {
	status := gochecker.NewComponentStatus()
	var (
		ok      string
		version string
		err     error
	)
	if i.db == nil {
		return *status
	}
	status.WithDetail("database", i.driverName)

	if i.validatorQuery != "" {
		status.WithDetail("validationQuery", i.validatorQuery)
		err = i.db.QueryRowContext(ctx, i.validatorQuery).Scan(&ok)
	} else {
		status.WithDetail("validationQuery", "ping()")
		err = i.db.PingContext(ctx)
	}
	if err != nil {
		return *status.WithDown().WithDetail("err", err.Error())
	}

	if i.versionQuery != "" {
		err = i.db.QueryRowContext(ctx, i.versionQuery).Scan(&version)
		if err != nil {
			status.WithDetail("version", err.Error())
		} else {
			status.WithDetail("version", version)
		}
	}
	return *status.WithUp()
}

// NewMySQLIndicator creates a new mysql database health check indicator
func NewMySQLIndicator(db *sql.DB) *Indicator {
	return NewIndicator(db, "mysql", MySQLValidationQuery, MySQLVersionQuery)
}

// NewSqlite3Indicator creates a new sqlite3 health check indicator
func NewSqlite3Indicator(db *sql.DB) *Indicator {
	return NewIndicator(db, "sqlite3", Sqlite3ValidationQuery, Sqlite3VersionQuery)
}

// NewIndicator creates a new database health check indicator with given validation, version query
func NewIndicator(db *sql.DB, driverName, validationQuery, versionQuery string) *Indicator {
	return &Indicator{db: db, driverName: driverName, validatorQuery: validationQuery, versionQuery: versionQuery}
}
