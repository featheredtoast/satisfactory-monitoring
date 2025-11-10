package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/benbjohnson/clock"
)

// Struct for the JSON body of POST requests
type FrmApiRequest struct {
	Function string `json:"function"`
	Endpoint string `json:"endpoint"`
}

// Reusable HTTP client for HTTPS w/ self-signed certs
var tlsClient = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	},
}

var Clock = clock.New()

func sanitizeSessionName(sessionName string) string {
	re := regexp.MustCompile(`[^\w\s]`)
	return re.ReplaceAllString(sessionName, "")
}

type CacheWorker struct {
	ctx         context.Context
	cancel      context.CancelFunc
	frmBaseUrl  string
	sessionName string
	db          *sql.DB
	now         time.Time
}

type SessionInfo struct {
	SessionName string `json:"SessionName"`
}

func NewCacheWorker(frmBaseUrl string, db *sql.DB) *CacheWorker {
	ctx, cancel := context.WithCancel(context.Background())

	return &CacheWorker{
		frmBaseUrl:  frmBaseUrl,
		sessionName: "default",
		db:          db,
		ctx:         ctx,
		cancel:      cancel,
	}
}

func retrieveDataViaGET(frmAddress string) ([]string, error) {
	resp, err := http.Get(frmAddress)

	if err != nil {
		return nil, fmt.Errorf("error when parsing json: %s", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("non-200 returned when retireving data: %d", resp.StatusCode)
	}

	var content []json.RawMessage
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&content)
	if err != nil {
		return nil, fmt.Errorf("error when parsing json: %s", err)
	}
	result := []string{}
	for _, c := range content {
		result = append(result, string(c[:]))
	}
	return result, nil
}

func retrieveSessionInfoViaGET(frmAddress string, data any) error {
	resp, err := http.Get(frmAddress)

	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("non-200 returned when retireving data: %d", resp.StatusCode)
	}

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&data)
	return err
}

func retrieveDataViaPOST(frmApiUrl string, endpointName string, details any) error {
	reqBody := FrmApiRequest{
		Function: "frm",
		Endpoint: endpointName,
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("error marshalling request body: %s", err)
	}

	req, err := http.NewRequest("POST", frmApiUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating POST request: %s", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := tlsClient.Do(req)
	if err != nil {
		return fmt.Errorf("error making POST request: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("non-200 returned: %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(&details)
}

func retrieveData(frmAddress string) ([]string, error) {
	u, err := url.Parse(frmAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid FRM address URL: %s", err)
	}

	// Check if we're using the Dedicated Server API
	if strings.HasPrefix(u.Scheme, "https") && strings.HasSuffix(u.Path, "/api/v1") {
		return nil, fmt.Errorf("dedicated server API detected in retrieveData, but this function expects old URL format")
	} else if strings.HasPrefix(u.Scheme, "https") && strings.Contains(u.Path, "/api/v1/") {
		parts := strings.SplitN(frmAddress, "/api/v1/", 2)
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid dedicated server API URL format: %s", frmAddress)
		}

		baseUrl := parts[0] + "/api/v1"
		endpointName := parts[1] // "getDrone", "getTrains", etc.

		var content []json.RawMessage
		err = retrieveDataViaPOST(baseUrl, endpointName, &content)
		if err != nil {
			return nil, fmt.Errorf("error when parsing json: %s", err)
		}

		// Convert []json.RawMessage to []string
		result := []string{}
		for _, c := range content {
			result = append(result, string(c[:]))
		}
		return result, nil

	} else {
		// Standard Web Server mode.
		return retrieveDataViaGET(frmAddress)
	}
}

func retrieveSessionInfo(frmAddress string, data any) error {
	u, err := url.Parse(frmAddress)
	if err != nil {
		return fmt.Errorf("invalid FRM address URL: %s", err)
	}

	// Check if we're using the Dedicated Server API
	if strings.HasPrefix(u.Scheme, "https") && strings.HasSuffix(u.Path, "/api/v1") {
		// This should not be hit by this function, but we handle it.
		return fmt.Errorf("dedicated server API detected in retrieveSessionInfo, but this function expects old URL format")
	} else if strings.HasPrefix(u.Scheme, "https") && strings.Contains(u.Path, "/api/v1/") {
		// This is the one we see in the logs: "https://.../api/v1/getSessionInfo"
		parts := strings.SplitN(frmAddress, "/api/v1/", 2)
		if len(parts) < 2 {
			return fmt.Errorf("invalid dedicated server API URL format: %s", frmAddress)
		}
		baseUrl := parts[0] + "/api/v1"
		endpointName := parts[1] // This will be "getSessionInfo"

		return retrieveDataViaPOST(baseUrl, endpointName, data)

	} else {
		// Standard Web Server mode.
		return retrieveSessionInfoViaGET(frmAddress, data)
	}
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

	delete := `delete from cache where metric = $1 AND url = $2 AND session_name = $3;`
	_, err = tx.Exec(delete, metric, c.frmBaseUrl, c.sessionName)
	if err != nil {
		return
	}
	for _, s := range data {
		insert := `insert into cache (metric,data,url,session_name) values($1,$2,$3,$4)`
		_, err = tx.Exec(insert, metric, s, c.frmBaseUrl, c.sessionName)
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
		insert := `insert into cache_with_history (metric,data,time,url,session_name) values($1,$2,$3,$4,$5)`
		_, err = tx.Exec(insert, metric, s, c.now, c.frmBaseUrl, c.sessionName)
		if err != nil {
			return
		}
	}

	delete := `delete from cache_with_history where
metric = $1 AND
url = $2 AND session_name = $3 AND
time < $4;`
	_, err = tx.Exec(delete, metric, c.frmBaseUrl, c.sessionName, c.now.Add(time.Duration(-1)*time.Hour))
	return
}

// flush the metric history cache
func (c *CacheWorker) flushMetricHistory() error {
	delete := `DELETE from cache_with_history c WHERE c.url = $1 and c.session_name = $2;`
	_, err := c.db.Exec(delete, c.frmBaseUrl, c.sessionName)
	if err != nil {
		log.Println("flush metrics history db error: ", err)
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
		log.Println("Error when pulling metrics ", metric, ": ", err)
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
	c.pullMetricsLog("cloudInv", "/getCloudInv", false)
	c.pullMetricsLog("droneStation", "/getDroneStation", false)
}

func (c *CacheWorker) pullRealtimeMetrics() {
	c.pullMetricsLog("drone", "/getDrone", true)
	c.pullMetricsLog("train", "/getTrains", true)
	c.pullMetricsLog("truck", "/getVehicles", true)
	c.pullMetricsLog("trainStation", "/getTrainStation", true)
	c.pullMetricsLog("truckStation", "/getTruckStation", true)
	c.pullMetricsLog("player", "/getPlayer", false)
}

func (c *CacheWorker) pullSessionName() {
	sessionInfo := SessionInfo{}
	err := retrieveSessionInfo(c.frmBaseUrl+"/getSessionInfo", &sessionInfo)
	if err != nil {
		log.Printf("error reading session name from FRM: %s\n", err)
	}
	newSessionName := sanitizeSessionName(sessionInfo.SessionName)
	if newSessionName != "" && newSessionName != c.sessionName {
		log.Println(c.frmBaseUrl + " has a new session name: " + newSessionName)
		c.sessionName = newSessionName
		c.flushMetricHistory()
	}
}

func (c *CacheWorker) Start() {
	c.now = Clock.Now()
	c.pullSessionName()
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
			c.pullSessionName()
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
