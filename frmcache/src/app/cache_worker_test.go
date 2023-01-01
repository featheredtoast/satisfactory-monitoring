package main

import (
	"fmt"
	sql "github.com/DATA-DOG/go-sqlmock"
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
	defer db.Close()
	frm := setupServer(`[{"Name":"Quickwire","Count":500},{"Name":"Wire","Count":500}]`)
	defer frm.Close()
	mock.ExpectBegin()
	mock.ExpectExec(`^delete from cache (.+)`).WithArgs("worldInv").WillReturnResult(sql.NewResult(1, 1))
	mock.ExpectExec(`^insert into cache \(metric,data\)(.+)`).WithArgs("worldInv", `{"Name":"Quickwire","Count":500}`).WillReturnResult(sql.NewResult(1, 1))
	mock.ExpectExec(`^insert into cache \(metric,data\)(.+)`).WithArgs("worldInv", `{"Name":"Wire","Count":500}`).WillReturnResult(sql.NewResult(1, 1))
	mock.ExpectCommit()

	c := NewCacheWorker(frm.URL, db)
	c.pullMetrics("worldInv", "/getWorldInv", false)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal("frm didn't update correctly", err)
	}
}

func TestCacheHistory(t *testing.T) {
	db, mock, _ := sql.New()
	defer db.Close()
	frm := setupServer(`[{"ID":"0","VehicleType":"Drone"},{"ID":"1","VehicleType":"Drone"}]`)
	defer frm.Close()
	mock.ExpectBegin()
	mock.ExpectExec(`^delete from cache (.+)`).WithArgs("drone").WillReturnResult(sql.NewResult(1, 1))
	mock.ExpectExec(`^insert into cache \(metric,data\)(.+)`).WithArgs("drone", `{"ID":"0","VehicleType":"Drone"}`).WillReturnResult(sql.NewResult(1, 1))
	mock.ExpectExec(`^insert into cache \(metric,data\)(.+)`).WithArgs("drone", `{"ID":"1","VehicleType":"Drone"}`).WillReturnResult(sql.NewResult(1, 1))
	mock.ExpectCommit()
	mock.ExpectBegin()
	mock.ExpectExec(`^insert into cache_with_history (.+)`).WithArgs("drone", `{"ID":"0","VehicleType":"Drone"}`).WillReturnResult(sql.NewResult(1, 1))
	mock.ExpectExec(`^insert into cache_with_history (.+)`).WithArgs("drone", `{"ID":"1","VehicleType":"Drone"}`).WillReturnResult(sql.NewResult(1, 1))
	mock.ExpectExec(`^delete from cache_with_history (.+)`).WithArgs("drone", 720*2).WillReturnResult(sql.NewResult(1, 1))
	mock.ExpectCommit()

	c := NewCacheWorker(frm.URL, db)
	if err := c.pullMetrics("drone", "/getDrone", true); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal("frm didn't update correctly", err)
	}
}
