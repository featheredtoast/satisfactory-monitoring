package main

import (
	"context"
	"encoding/json"
	"github.com/benbjohnson/clock"
	"github.com/elastic/go-elasticsearch/v8"
	"net/http"

	"fmt"
	"time"
)

var Clock = clock.New()

type CacheWorker struct {
	ctx        context.Context
	cancel     context.CancelFunc
	frmBaseUrl string
	es         *elasticsearch.Client
}

func NewCacheWorker(frmBaseUrl string) *CacheWorker {
	ctx, cancel := context.WithCancel(context.Background())

	es, err := elasticsearch.NewDefaultClient()
	if err != nil {
		fmt.Println("Error creating the es client: %s", err)
	}
	return &CacheWorker{
		frmBaseUrl: frmBaseUrl,
		ctx:        ctx,
		cancel:     cancel,
		es:         es,
	}
}

func retrieveData(frmAddress string, details any) error {
	resp, err := http.Get(frmAddress)

	if err != nil {
		fmt.Printf("error fetching statistics from FRM: %s\n", err)
		return err
	}

	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)

	err = decoder.Decode(&details)
	return err
}

func (c *CacheWorker) pullMetrics() {
	buildings := []BuildingDetail{}
	retrieveData(c.frmBaseUrl+"/getFactory", buildings)
}

func (c *CacheWorker) Start() {
	for {
		select {
		case <-c.ctx.Done():
			return
		case <-Clock.After(5 * time.Second):
			c.pullMetrics()
		}
	}
}

func (c *CacheWorker) Stop() {
	c.cancel()
}
