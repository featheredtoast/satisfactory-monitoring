package main

import (
	"fmt"
	sql "github.com/DATA-DOG/go-sqlmock"
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/benbjohnson/clock"
)

func setupServer(expectedResponse string) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, expectedResponse)
	}))
	return ts
}

func TestCache(t *testing.T) {
	db, mock, _ := sql.New()
	saveName := "default"
	defer db.Close()
	frm := setupServer(`[{"Name":"Quickwire","Count":500},{"Name":"Wire","Count":500}]`)
	defer frm.Close()
	mock.ExpectBegin()
	mock.ExpectExec(`^delete from cache (.+)`).WithArgs("worldInv", frm.URL, saveName).WillReturnResult(sql.NewResult(1, 1))
	mock.ExpectExec(`^insert into cache \(metric,data,url,save\)(.+)`).WithArgs("worldInv", `{"Name":"Quickwire","Count":500}`, frm.URL, saveName).WillReturnResult(sql.NewResult(1, 1))
	mock.ExpectExec(`^insert into cache \(metric,data,url,save\)(.+)`).WithArgs("worldInv", `{"Name":"Wire","Count":500}`, frm.URL, saveName).WillReturnResult(sql.NewResult(1, 1))
	mock.ExpectCommit()

	c := NewCacheWorker(frm.URL, db)
	c.pullMetrics("worldInv", "/getWorldInv", false)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal("frm didn't update correctly", err)
	}
}

func TestCacheHistory(t *testing.T) {
	db, mock, _ := sql.New()
	saveName := "default"
	defer db.Close()
	frm := setupServer(`[{"ID":"0","VehicleType":"Drone"},{"ID":"1","VehicleType":"Drone"}]`)
	defer frm.Close()
	Clock = clock.NewMock()
	now := Clock.Now()
	mock.ExpectBegin()
	mock.ExpectExec(`^delete from cache (.+)`).WithArgs("drone", frm.URL, saveName).WillReturnResult(sql.NewResult(1, 1))
	mock.ExpectExec(`^insert into cache \(metric,data,url,save\)(.+)`).WithArgs("drone", `{"ID":"0","VehicleType":"Drone"}`, frm.URL, saveName).WillReturnResult(sql.NewResult(1, 1))
	mock.ExpectExec(`^insert into cache \(metric,data,url,save\)(.+)`).WithArgs("drone", `{"ID":"1","VehicleType":"Drone"}`, frm.URL, saveName).WillReturnResult(sql.NewResult(1, 1))
	mock.ExpectCommit()
	mock.ExpectBegin()
	mock.ExpectExec(`^insert into cache_with_history (.+)`).WithArgs("drone", `{"ID":"0","VehicleType":"Drone"}`, now, frm.URL, saveName).WillReturnResult(sql.NewResult(1, 1))
	mock.ExpectExec(`^insert into cache_with_history (.+)`).WithArgs("drone", `{"ID":"1","VehicleType":"Drone"}`, now, frm.URL, saveName).WillReturnResult(sql.NewResult(1, 1))
	mock.ExpectExec(`^delete from cache_with_history (.+)`).WithArgs("drone", frm.URL, saveName, 720*2).WillReturnResult(sql.NewResult(1, 1))
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
