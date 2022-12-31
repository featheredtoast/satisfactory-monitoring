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

func (c *CacheWorker) cacheMetrics(metric string, data string) error {
	insert := `insert into "cache" ("metric","frm_data") values($1,$2) ON CONFLICT (metric) DO UPDATE SET FRM_DATA = EXCLUDED.frm_data`
	_, err := c.db.Exec(insert, metric, data)
	return err
}

func (c *CacheWorker) cacheMetricsWithHistory(metric string, data string) error {
	insert := `insert into "cache_with_history" ("metric","frm_data", "time") values($1,$2, now())`
	_, err := c.db.Exec(insert, metric, data)
	return err
}

// flush the metric history cache
func (c *CacheWorker) flushMetricHistory() error {
	delete := `truncate cache_with_history;`
	_, err := c.db.Exec(delete)
	if err != nil {
		fmt.Println("flush metrics history db error: ", err)
	}
	return err
}

// Keep at most 1 hour of records
func (c *CacheWorker) rotateMetricHistory(metric string) error {
	delete := `delete from "cache_with_history" where
metric = $1 and
id NOT IN (
select id from "cache_with_history" where metric = $1
order by id desc
limit 720
);`
	_, err := c.db.Exec(delete, metric)
	if err != nil {
		fmt.Println("rotate metrics history db error: ", err)
	}
	return err
}

func (c *CacheWorker) pullMetrics(metric string, route string, keepHistory bool) {
	data, err := retrieveData(c.frmBaseUrl + route)
	if err != nil {
		fmt.Println("error when parsing json: ", err)
		return
	}
	c.cacheMetrics(metric, data)
	if err != nil {
		fmt.Println("error when caching metrics", err)
	}
	if keepHistory {
		err = c.cacheMetricsWithHistory(metric, data)
		if err != nil {
			fmt.Println("error when caching metrics history", err)
			return
		}
		err = c.rotateMetricHistory(metric)
		if err != nil {
			fmt.Println("error when rotating metrics", err)
		}
	}
}

func (c *CacheWorker) pullLowCadenceMetrics() {
	c.pullMetrics("factory", "/getFactory", false)
	c.pullMetrics("extractor", "/getExtractor", false)
	c.pullMetrics("dropPod", "/getDropPod", false)
	c.pullMetrics("storageInv", "/getStorageInv", false)
	c.pullMetrics("worldInv", "/getWorldInv", false)
	c.pullMetrics("droneStation", "/getDroneStation", false)
	c.pullMetrics("trainStation", "/getTrainStation", false)
	c.pullMetrics("truckStation", "/getTruckStation", false)
}

func (c *CacheWorker) pullRealtimeMetrics() {
	c.pullMetrics("drone", "/getDrone", true)
	c.pullMetrics("train", "/getTrains", true)
	c.pullMetrics("truck", "/getVehicles", true)
}

func (c *CacheWorker) Start() {
	c.flushMetricHistory()
	c.pullLowCadenceMetrics()
	c.pullRealtimeMetrics()
	counter := 0
	for {
		select {
		case <-c.ctx.Done():
			return
		case <-time.After(5 * time.Second):
			counter = counter + 1
			c.pullRealtimeMetrics()
			if counter > 11 {
				c.pullLowCadenceMetrics()
				counter = 0
			}
		}
	}
}

func (c *CacheWorker) Stop() {
	c.cancel()
}
