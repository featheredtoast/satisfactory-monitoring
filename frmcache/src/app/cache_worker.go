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
	insert := `insert into "cache"("metric","frm_data") values($1,$2) ON CONFLICT (metric) DO UPDATE SET FRM_DATA = EXCLUDED.frm_data`
	c.db.Exec(insert, metric, data)
}

func (c *CacheWorker) pullMetrics(metric string, route string) {
	data, err := retrieveData(c.frmBaseUrl + route)
	if err != nil {
		fmt.Println("error when parsing json: ", err)
	}
	c.cacheMetrics(metric, data)
}

func (c *CacheWorker) pullAllMetrics() {
	c.pullMetrics("factory", "/getFactory")
	c.pullMetrics("dropPod", "/getDropPod")
	c.pullMetrics("storageInv", "/getStorageInv")
	c.pullMetrics("worldInv", "/getWorldInv")
	c.pullMetrics("droneStation", "/getDroneStation")
	c.pullMetrics("trainStation", "/getTrainStation")
	c.pullMetrics("truckStation", "/getTruckStation")
}

func (c *CacheWorker) Start() {
	c.pullAllMetrics()
	for {
		select {
		case <-c.ctx.Done():
			return
		case <-time.After(60 * time.Second):
			c.pullAllMetrics()
		}
	}
}

func (c *CacheWorker) Stop() {
	c.cancel()
}
