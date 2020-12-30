package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"github.com/zacscoding/gochecker"
	"testing"
)

func TestHealth(t *testing.T) {
	cases := []struct {
		Name       string
		SetupMock  func(m sqlmock.Sqlmock)
		Supplier   func(db *sql.DB) *Indicator
		AssertFunc func(t *testing.T, status gochecker.ComponentStatus, mock sqlmock.Sqlmock)
	}{
		{
			Name: "MySQL-Indicator Up",
			SetupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(MySQLValidationQuery).WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow("1"))
				m.ExpectQuery(MySQLVersionQuery).WillReturnRows(sqlmock.NewRows([]string{"version()"}).AddRow("5.6.1"))
			},
			Supplier: func(db *sql.DB) *Indicator {
				return NewMySQLIndicator(db)
			},
			AssertFunc: func(t *testing.T, status gochecker.ComponentStatus, mock sqlmock.Sqlmock) {
				// assert status
				assert.True(t, status.IsUp())
				// assert details
				bytes, err := json.Marshal(&status)
				assert.NoError(t, err)
				assert.Equal(t, "mysql", gjson.GetBytes(bytes, "details.database").String())
				assert.Equal(t, MySQLValidationQuery, gjson.GetBytes(bytes, "details.validationQuery").String())
				assert.Equal(t, "5.6.1", gjson.GetBytes(bytes, "details.version").String())
			},
		}, {
			Name: "MySQL-Indicator Up with version query fail",
			SetupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(MySQLValidationQuery).WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow("1"))
				m.ExpectQuery(MySQLVersionQuery).WillReturnError(errors.New("force error"))
			},
			Supplier: func(db *sql.DB) *Indicator {
				return NewMySQLIndicator(db)
			},
			AssertFunc: func(t *testing.T, status gochecker.ComponentStatus, mock sqlmock.Sqlmock) {
				// assert status
				assert.True(t, status.IsUp())
				// assert details
				bytes, err := json.Marshal(&status)
				assert.NoError(t, err)
				assert.Equal(t, "mysql", gjson.GetBytes(bytes, "details.database").String())
				assert.Equal(t, MySQLValidationQuery, gjson.GetBytes(bytes, "details.validationQuery").String())
				assert.Contains(t, gjson.GetBytes(bytes, "details.version").String(), "force error")
			},
		}, {
			Name: "MySQL-Indicator Down",
			SetupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(MySQLValidationQuery).WillReturnError(errors.New("force error"))
			},
			Supplier: func(db *sql.DB) *Indicator {
				return NewMySQLIndicator(db)
			},
			AssertFunc: func(t *testing.T, status gochecker.ComponentStatus, mock sqlmock.Sqlmock) {
				// assert status
				assert.True(t, status.IsDown())
				// assert details
				bytes, err := json.Marshal(&status)
				assert.NoError(t, err)
				assert.Equal(t, "mysql", gjson.GetBytes(bytes, "details.database").String())
				assert.Contains(t, gjson.GetBytes(bytes, "details.err").String(), "force error")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			tc.SetupMock(mock)
			indicator := tc.Supplier(db)
			// when
			status := indicator.Health(context.TODO())
			// then
			tc.AssertFunc(t, status, mock)
		})
	}
}
