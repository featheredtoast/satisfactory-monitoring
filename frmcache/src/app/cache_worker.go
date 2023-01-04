package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"

	"fmt"
	"github.com/benbjohnson/clock"
	"time"
)

var Clock = clock.New()

type CacheWorker struct {
	ctx        context.Context
	cancel     context.CancelFunc
	frmBaseUrl string
	db         *sql.DB
	now time.Time
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

func retrieveData(frmAddress string) ([]string, error) {
	resp, err := http.Get(frmAddress)

	if err != nil {
		return nil, fmt.Errorf("error when parsing json: %s", err)
	}

	var content []json.RawMessage
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&content)
	if err != nil {
		return nil, fmt.Errorf("error when parsing json: %s", err)
	}
	defer resp.Body.Close()
	result := []string{}
	for _, c := range content {
		result = append(result, string(c[:]))
	}
	return result, nil
}

func (c *CacheWorker) cacheMetrics(metric string, data []string) (err error) {
	tx, err := c.db.Begin()
	if err != nil {
		return
	}
	defer func() {
		switch err {
		case nil:
			err = tx.Commit()
		default:
			tx.Rollback()
		}
	}()

	delete := `delete from cache where metric = $1;`
	_, err = tx.Exec(delete, metric)
	if (err != nil) {
		return
	}
	for _, s := range data {
		insert := `insert into cache (metric,data) values($1,$2)`
		_, err = tx.Exec(insert, metric, s)
		if err != nil {
			return
		}
	}
	return
}

func (c *CacheWorker) cacheMetricsWithHistory(metric string, data []string) (err error) {
	tx, err := c.db.Begin()
	if err != nil {
		return
	}
	defer func() {
		switch err {
		case nil:
			err = tx.Commit()
		default:
			tx.Rollback()
		}
	}()
	for _, s := range data {
		insert := `insert into cache_with_history (metric,data, time) values($1,$2,$3)`
		_, err = tx.Exec(insert, metric, s, c.now)
		if err != nil {
			return
		}
	}

	//720 = 1 hour, 5 second increments. retain that many rows for every data.
	keep := 720 * len(data)

	delete := `delete from cache_with_history where
metric = $1 and
id NOT IN (
select id from "cache_with_history" where metric = $1
order by id desc
limit $2
);`
	_, err = tx.Exec(delete, metric, keep)
	return
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

func (c *CacheWorker) pullMetrics(metric string, route string, keepHistory bool) error {
	data, err := retrieveData(c.frmBaseUrl + route)
	if err != nil {
		return fmt.Errorf("error when parsing json: %s", err)
	}
	c.cacheMetrics(metric, data)
	if err != nil {
		return fmt.Errorf("error when caching metrics %s", err)
	}
	if keepHistory {
		err = c.cacheMetricsWithHistory(metric, data)
		if err != nil {
			return fmt.Errorf("error when caching metrics history %s", err)
		}
	}
	return nil
}

func (c *CacheWorker) pullMetricsLog(metric string, route string, keepHistory bool) error {
	if err := c.pullMetrics(metric, route, keepHistory); err != nil {
		fmt.Println("Error when pulling metrics ", metric, ": ", err)
		return err
	}
	return nil
}

func (c *CacheWorker) pullLowCadenceMetrics() {
	c.pullMetricsLog("factory", "/getFactory", true)
	c.pullMetricsLog("extractor", "/getExtractor", true)
	c.pullMetricsLog("dropPod", "/getDropPod", false)
	c.pullMetricsLog("storageInv", "/getStorageInv", false)
	c.pullMetricsLog("worldInv", "/getWorldInv", false)
	c.pullMetricsLog("droneStation", "/getDroneStation", false)
}

func (c *CacheWorker) pullRealtimeMetrics() {
	c.pullMetricsLog("drone", "/getDrone", true)
	c.pullMetricsLog("train", "/getTrains", true)
	c.pullMetricsLog("truck", "/getVehicles", true)
	c.pullMetricsLog("trainStation", "/getTrainStation", true)
	c.pullMetricsLog("truckStation", "/getTruckStation", true)
}

func (c *CacheWorker) Start() {
	//TODO: grab current time, update here instead of using postgres time.now so all metrics can sync on the same time and can aggregate properly.
	c.now = Clock.Now()
	c.flushMetricHistory()
	c.pullLowCadenceMetrics()
	c.pullRealtimeMetrics()
	counter := 0
	for {
		select {
		case <-c.ctx.Done():
			return
		case <-Clock.After(5 * time.Second):
			c.now = Clock.Now()
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
