package query

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"ops_cli/internal/checker"
	"ops_cli/internal/config"
	"ops_cli/pkg/log"
	"time"
)

type QueryRangeChecker struct {
	config         *config.Config
	client         *http.Client
	generalPort    int
	opsPort        int
	generalQueries []PrometheusQuery
	opsQueries     []PrometheusQuery
	start          string
	end            string
}

func NewQueryRangeChecker(cfg *config.Config) *QueryRangeChecker {
	generalPort, generalQueries, start, end := loadQueries("query.yaml", "query_range", "general")
	opsPort, opsQueries, _, _ := loadQueries("query.yaml", "query_range", "ops")

	return &QueryRangeChecker{
		config:         cfg,
		client:         &http.Client{Timeout: 10 * time.Second},
		generalPort:    generalPort,
		opsPort:        opsPort,
		generalQueries: generalQueries,
		opsQueries:     opsQueries,
		start:          start,
		end:            end,
	}
}

func (qr *QueryRangeChecker) Name() string {
	return "query_range"
}

func (qr *QueryRangeChecker) Check() []checker.CheckResult {
	var results []checker.CheckResult

	for _, ip := range qr.config.IPs {
		log.Info("Checking Prometheus query range for %s", ip.IP)
		if ip.Role == "ops" {
			for _, query := range qr.opsQueries {
				results = append(results, qr.checkQueryRange(ip, query, qr.opsPort))
			}
		} else {
			for _, query := range qr.generalQueries {
				results = append(results, qr.checkQueryRange(ip, query, qr.generalPort))
			}
		}
	}

	return results
}

func (qr *QueryRangeChecker) checkQueryRange(ip config.IPConfig, query PrometheusQuery, port int) checker.CheckResult {
	log.Info("Checking Prometheus query range for %s", ip.IP)

	// Parse the start and end times into time.Time objects
	parsedStart, err := time.ParseInLocation("2006-01-02 15:04:05", qr.start, time.FixedZone("CST", 8*3600))
	if err != nil {
		return qr.createFailedResult(query.Name, ip, "Failed to parse start time", err)
	}
	parsedEnd, err := time.ParseInLocation("2006-01-02 15:04:05", qr.end, time.FixedZone("CST", 8*3600))
	if err != nil {
		return qr.createFailedResult(query.Name, ip, "Failed to parse end time", err)
	}

	// Convert the times to Unix timestamps
	unixStart := parsedStart.UnixNano() / int64(time.Second) // Convert to seconds
	unixEnd := parsedEnd.UnixNano() / int64(time.Second)     // Convert to seconds

	encodedQuery := url.QueryEscape(query.Query)
	url := fmt.Sprintf("http://%s:%d/api/v1/query_range?query=%s&start=%d&end=%d&step=60s", ip.IP, port, encodedQuery, unixStart, unixEnd)
	log.Info("Making HTTP request to %s with timeout %v", url, qr.client.Timeout)

	resp, err := qr.client.Get(url)
	result := qr.createBaseResult(query.Name, ip)

	if err != nil {
		return qr.createFailedResult(query.Name, ip, "HTTP request failed", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return qr.createFailedResult(query.Name, ip, fmt.Sprintf("API returned status code %d", resp.StatusCode), nil)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return qr.createFailedResult(query.Name, ip, "Failed to read response body", err)
	}

	var jsonResponse struct {
		Status string `json:"status"`
		Data   struct {
			ResultType string `json:"resultType"`
			Result     []struct {
				Metric map[string]string `json:"metric"`
				Values [][]interface{}   `json:"values"`
			} `json:"result"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &jsonResponse); err != nil {
		return qr.createFailedResult(query.Name, ip, "Failed to parse JSON response", err)
	}

	if len(jsonResponse.Data.Result) > 0 && len(jsonResponse.Data.Result[0].Values) > 0 {
		result.Status = "Passed"
		// Get the latest value from the time series
		lastValue := jsonResponse.Data.Result[0].Values[len(jsonResponse.Data.Result[0].Values)-1]
		if len(lastValue) > 1 {
			result.Message = fmt.Sprintf("%v", lastValue[1])
		} else {
			result.Status = "Failed"
			result.Message = "Invalid value format in response"
		}
	} else {
		result.Status = "Failed"
		result.Message = "No data returned"
	}

	log.Info("Prometheus query range check passed for %s", ip.IP)

	return result
}

func (qr *QueryRangeChecker) createBaseResult(item string, ip config.IPConfig) checker.CheckResult {
	return checker.CheckResult{
		Component: qr.Name(),
		Item:      item,
		Role:      ip.Role,
		IP:        ip.IP,
	}
}

func (qr *QueryRangeChecker) createFailedResult(item string, ip config.IPConfig, message string, err error) checker.CheckResult {
	result := qr.createBaseResult(item, ip)
	result.Status = "Failed"
	if err != nil {
		result.Message = fmt.Sprintf("%s: %v", message, err)
	} else {
		result.Message = message
	}
	log.Error("%s check failed for %s: %v", item, ip.IP, err)
	return result
}
