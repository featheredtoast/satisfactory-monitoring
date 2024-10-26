package main

import (
	"fmt"
	sql "github.com/DATA-DOG/go-sqlmock"
	"github.com/benbjohnson/clock"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupServer(expectedResponse string) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, expectedResponse)
	}))
	return ts
}

func TestCache(t *testing.T) {
	db, mock, _ := sql.New()
	sessionName := "default"
	defer db.Close()
	frm := setupServer(`[{"Name":"Quickwire","Count":500},{"Name":"Wire","Count":500}]`)
	defer frm.Close()
	mock.ExpectBegin()
	mock.ExpectExec(`^delete from cache (.+)`).WithArgs("worldInv", frm.URL, sessionName).WillReturnResult(sql.NewResult(1, 1))
	mock.ExpectExec(`^insert into cache \(metric,data,url,session_name\)(.+)`).WithArgs("worldInv", `{"Name":"Quickwire","Count":500}`, frm.URL, sessionName).WillReturnResult(sql.NewResult(1, 1))
	mock.ExpectExec(`^insert into cache \(metric,data,url,session_name\)(.+)`).WithArgs("worldInv", `{"Name":"Wire","Count":500}`, frm.URL, sessionName).WillReturnResult(sql.NewResult(1, 1))
	mock.ExpectCommit()

	c := NewCacheWorker(frm.URL, db)
	c.pullMetrics("worldInv", "/getWorldInv", false)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal("frm didn't update correctly", err)
	}
}

func TestCacheHistory(t *testing.T) {
	db, mock, _ := sql.New()
	sessionName := "default"
	defer db.Close()
	frm := setupServer(`[{"ID":"0","VehicleType":"Drone"},{"ID":"1","VehicleType":"Drone"}]`)
	defer frm.Close()
	Clock = clock.NewMock()
	now := Clock.Now()
	mock.ExpectBegin()
	mock.ExpectExec(`^delete from cache (.+)`).WithArgs("drone", frm.URL, sessionName).WillReturnResult(sql.NewResult(1, 1))
	mock.ExpectExec(`^insert into cache \(metric,data,url,session_name\)(.+)`).WithArgs("drone", `{"ID":"0","VehicleType":"Drone"}`, frm.URL, sessionName).WillReturnResult(sql.NewResult(1, 1))
	mock.ExpectExec(`^insert into cache \(metric,data,url,session_name\)(.+)`).WithArgs("drone", `{"ID":"1","VehicleType":"Drone"}`, frm.URL, sessionName).WillReturnResult(sql.NewResult(1, 1))
	mock.ExpectCommit()
	mock.ExpectBegin()
	mock.ExpectExec(`^insert into cache_with_history (.+)`).WithArgs("drone", `{"ID":"0","VehicleType":"Drone"}`, now, frm.URL, sessionName).WillReturnResult(sql.NewResult(1, 1))
	mock.ExpectExec(`^insert into cache_with_history (.+)`).WithArgs("drone", `{"ID":"1","VehicleType":"Drone"}`, now, frm.URL, sessionName).WillReturnResult(sql.NewResult(1, 1))
	mock.ExpectExec(`^delete from cache_with_history (.+)`).WithArgs("drone", frm.URL, sessionName, 720*2).WillReturnResult(sql.NewResult(1, 1))
	mock.ExpectCommit()

	c := NewCacheWorker(frm.URL, db)
	c.now = now
	if err := c.pullMetrics("drone", "/getDrone", true); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal("frm didn't update correctly", err)
	}
}
