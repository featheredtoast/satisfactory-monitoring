package main

import (
	"context"
	"database/sql"
	"io"
	"net/http"
	"strings"

	"fmt"
	"time"
	"github.com/benbjohnson/clock"
)

var Clock = clock.New()

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
		return "", fmt.Errorf("error when parsing json: %s", err)
	}

	buf := new(strings.Builder)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		return "", fmt.Errorf("error when parsing json: %s", err)
	}
	defer resp.Body.Close()
	return buf.String(), nil
}

func (c *CacheWorker) cacheMetrics(metric string, data string) error {
	insert := `insert into cache (metric,frm_data) values($1,$2) ON CONFLICT (metric) DO UPDATE SET FRM_DATA = EXCLUDED.frm_data`
	_, err := c.db.Exec(insert, metric, data)
	return err
}

func (c *CacheWorker) cacheMetricsWithHistory(metric string, data string) error {
	insert := `insert into cache_with_history (metric,frm_data, time) values($1,$2, now())`
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
	delete := `delete from cache_with_history where
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
		err = c.rotateMetricHistory(metric)
		if err != nil {
			return fmt.Errorf("error when rotating metrics %s", err)
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
	c.pullMetricsLog("factory", "/getFactory", false)
	c.pullMetricsLog("extractor", "/getExtractor", false)
	c.pullMetricsLog("dropPod", "/getDropPod", false)
	c.pullMetricsLog("storageInv", "/getStorageInv", false)
	c.pullMetricsLog("worldInv", "/getWorldInv", false)
	c.pullMetricsLog("droneStation", "/getDroneStation", false)
	c.pullMetricsLog("trainStation", "/getTrainStation", false)
	c.pullMetricsLog("truckStation", "/getTruckStation", false)
}

func (c *CacheWorker) pullRealtimeMetrics() {
	c.pullMetricsLog("drone", "/getDrone", true)
	c.pullMetricsLog("train", "/getTrains", true)
	c.pullMetricsLog("truck", "/getVehicles", true)
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
		case <-Clock.After(5 * time.Second):
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
