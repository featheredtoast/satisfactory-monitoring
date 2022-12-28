package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/benbjohnson/clock"
	"net/http"
	"bytes"

	"fmt"
	"time"
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

func retrieveData(frmAddress string, details *any) error {
	resp, err := http.Get(frmAddress)

	if err != nil {
		fmt.Printf("error fetching statistics from FRM: %s\n", err)
		return err
	}

	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)

	err = decoder.Decode(details)
	return err
}

func (c *CacheWorker) cacheMetrics(metric string, details any) {
	byteArray, err := json.Marshal(details)
	CheckError(err)
	data := bytes.NewBuffer(byteArray).String()
	insert := `insert into "cache"("metric","frm_data") values($1,$2) ON CONFLICT (metric) DO UPDATE SET FRM_DATA = EXCLUDED.frm_data`
	c.db.Exec(insert, metric, data)
}

func (c *CacheWorker) pullMetrics(metric string, route string, details any) {
	err := retrieveData(c.frmBaseUrl+route, &details)
	if err != nil {
		fmt.Println("error when parsing json: ", err)
	}
	c.cacheMetrics(metric, details)
}

func (c *CacheWorker) pullAllMetrics() {
	c.pullMetrics("factory", "/getFactory", []BuildingDetail{})
}

func (c *CacheWorker) Start() {
	c.pullAllMetrics()
	for {
		select {
		case <-c.ctx.Done():
			return
		case <-Clock.After(5 * time.Second):
			c.pullAllMetrics()
		}
	}
}

func (c *CacheWorker) Stop() {
	c.cancel()
}
