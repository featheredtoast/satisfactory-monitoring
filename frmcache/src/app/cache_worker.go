package main

import (
	"context"
	"database/sql"
	"io"
	"net/http"
	"strings"

	"fmt"
	"time"
)

type CacheWorker struct {
	ctx        context.Context
	cancel     context.CancelFunc
	frmBaseUrl string
	db         *sql.DB
}

func NewCacheWorker(frmBaseUrl string, db *sql.DB) *CacheWorker {
	ctx, cancel := context.WithCancel(context.Background())

	return &CacheWorker{
		frmBaseUrl: frmBaseUrl,
		db:         db,
		ctx:        ctx,
		cancel:     cancel,
	}
}

func retrieveData(frmAddress string) (string, error) {
	resp, err := http.Get(frmAddress)

	if err != nil {
		fmt.Printf("error fetching statistics from FRM: %s\n", err)
		return "", err
	}

	buf := new(strings.Builder)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		fmt.Printf("error fetching statistics from FRM: %s\n", err)
		return "", err
	}
	defer resp.Body.Close()
	return buf.String(), nil
}

func (c *CacheWorker) cacheMetrics(metric string, data string) {
	insert := `insert into "cache" ("metric","frm_data") values($1,$2) ON CONFLICT (metric) DO UPDATE SET FRM_DATA = EXCLUDED.frm_data`
	c.db.Exec(insert, metric, data)
}

func (c *CacheWorker) cacheMetricsWithHistory(metric string, data string) {
	insert := `insert into "cache_with_history" ("metric","frm_data", "time") values($1,$2, now())`
	c.db.Exec(insert, metric, data)
}

func (c *CacheWorker) rotateCacheHistory() {
	insert := `delete from "cache_history" where time < 'now'::timestamp - '1 hour'::interval`
	c.db.Exec(insert)
}

func (c *CacheWorker) pullMetrics(metric string, route string, keepHistory bool) {
	data, err := retrieveData(c.frmBaseUrl + route)
	if err != nil {
		fmt.Println("error when parsing json: ", err)
	}
	if keepHistory {
		c.cacheMetricsWithHistory(metric, data)
	} else {
		c.cacheMetrics(metric, data)
	}
}

func (c *CacheWorker) pullLowCadenceMetrics() {
	c.pullMetrics("factory", "/getFactory", false)
	c.pullMetrics("dropPod", "/getDropPod", false)
	c.pullMetrics("storageInv", "/getStorageInv", false)
	c.pullMetrics("worldInv", "/getWorldInv", false)
	c.pullMetrics("droneStation", "/getDroneStation", false)
	c.pullMetrics("trainStation", "/getTrainStation", false)
	c.pullMetrics("truckStation", "/getTruckStation", false)
}

func (c *CacheWorker) pullRealtimeMetrics() {
	c.pullMetrics("droneStation", "/getDrone", true)
	c.pullMetrics("trainStation", "/getTrains", true)
	c.pullMetrics("truckStation", "/getVehicles", true)
}

func (c *CacheWorker) Start() {
	c.pullLowCadenceMetrics()
	for {
		select {
		case <-c.ctx.Done():
			return
		case <-time.After(10 * time.Minute):
			c.rotateCacheHistory()
		case <-time.After(60 * time.Second):
			c.pullLowCadenceMetrics()
		case <-time.After(5 * time.Second):
			c.pullRealtimeMetrics()
		}
	}
}

func (c *CacheWorker) Stop() {
	c.cancel()
}
