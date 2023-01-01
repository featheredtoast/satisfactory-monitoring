package main

import (
	"encoding/json"
	"fmt"
	"testing"
	sql "github.com/DATA-DOG/go-sqlmock"
	"net/http"
	"net/http/httptest"
)

func setupServer(expectedResponse string) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, expectedResponse)
	}))
	return ts
}

func TestJson(t *testing.T) {
	var content []json.RawMessage
	contentString := []byte(`[{"first":"val1","second":"val2"},{"first":"val3","second":"val4"}]`)
	t.Log("test testing 1")
	err := json.Unmarshal(contentString, &content)
	if err != nil {
		t.Log("nope cannot unmarshal due to ", err)
		t.Fail()
	}
	for _, c := range content {
		fmt.Println("yay got: ", string(c[:]))
	}
}

func TestCache(t *testing.T) {
	db, mock, _ := sql.New()
	defer db.Close()
	frm := setupServer(`[{"Name":"Quickwire","Count":500},{"Name":"Wire","Count":500}]`)
	defer frm.Close()
	mock.ExpectExec(`^insert into cache \(metric,frm_data\)(.+)`).WithArgs("worldInv", `[{"Name":"Quickwire","Count":500},{"Name":"Wire","Count":500}]`)

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
	mock.MatchExpectationsInOrder(true)
	mock.ExpectExec(`^insert into cache \(metric,frm_data\)(.+)`).WithArgs("drone", `[{"ID":"0","VehicleType":"Drone"},{"ID":"1","VehicleType":"Drone"}]`)
	mock.ExpectExec(`^insert into cache_with_history (.+)`).WithArgs("drone", `[{"ID":"0","VehicleType":"Drone"},{"ID":"1","VehicleType":"Drone"}]`).WillReturnResult(sql.NewResult(1,1))
	mock.ExpectExec(`^delete from cache_with_history (.+)`).WithArgs("drone").WillReturnResult(sql.NewResult(1,1))

	c := NewCacheWorker(frm.URL, db)
	if err := c.pullMetrics("drone", "/getDrone", true); err != nil {
t.Fatal("did not expect error", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal("frm didn't update correctly", err)
	}
}
