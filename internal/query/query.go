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

type QueryChecker struct {
	config         *config.Config
	client         *http.Client
	generalQueries []PrometheusQuery
	opsQueries     []PrometheusQuery
	queryTime      string
}

func NewQueryChecker(cfg *config.Config) *QueryChecker {
	generalQueries, queryTime, _ := loadQueries("query", "general")
	opsQueries, _, _ := loadQueries("query", "ops")
	return &QueryChecker{
		config:         cfg,
		client:         &http.Client{Timeout: 60 * time.Second},
		generalQueries: generalQueries,
		opsQueries:     opsQueries,
		queryTime:      queryTime,
	}
}

func (q *QueryChecker) Name() string {
	return "query"
}

func (q *QueryChecker) Check() []checker.CheckResult {
	var results []checker.CheckResult

	for _, ip := range q.config.IPs {
		if ip.Role == "ops" {
			for _, query := range q.opsQueries {
				results = append(results, q.checkQuery(ip, query))
			}
		} else {
			for _, query := range q.generalQueries {
				results = append(results, q.checkQuery(ip, query))
			}
		}
	}

	return results
}

func (q *QueryChecker) checkQuery(ip config.IPConfig, query PrometheusQuery) checker.CheckResult {
	log.Info("Checking Prometheus query for %s", ip.IP)

	// Parse the queryTime into a time.Time object
	parsedTime, err := time.ParseInLocation("2006-01-02 15:04:05", q.queryTime, time.FixedZone("CST", 8*3600))
	if err != nil {
		return q.createFailedResult(query.Name, ip, "Failed to parse query time", err)
	}

	// Convert the time to Unix timestamp
	unixTime := parsedTime.UnixNano() / int64(time.Second)

	encodedQuery := url.QueryEscape(query.Query)
	baseUrl, err := config.GetUrl(ip.IP, ip.Role, config.ComponentPrometheus, config.PathQuery)
	if err != nil {
		return q.createFailedResult(query.Name, ip, "Failed to get base url", err)
	}
	url := fmt.Sprintf("%s?query=%s&time=%d", baseUrl, encodedQuery, unixTime)
	log.Info("Making HTTP request to %s with timeout %v", url, q.client.Timeout)

	resp, err := q.client.Get(url)
	result := q.createBaseResult(query.Name, ip)

	if err != nil {
		return q.createFailedResult(query.Name, ip, "HTTP request failed", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return q.createFailedResult(query.Name, ip, fmt.Sprintf("API returned status code %d", resp.StatusCode), nil)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return q.createFailedResult(query.Name, ip, "Failed to read response body", err)
	}

	var jsonResponse struct {
		Status string `json:"status"`
		Data   struct {
			ResultType string `json:"resultType"`
			Result     []struct {
				Value []interface{} `json:"value"`
			} `json:"result"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &jsonResponse); err != nil {
		return q.createFailedResult(query.Name, ip, "Failed to parse JSON response", err)
	}

	if len(jsonResponse.Data.Result) > 0 && len(jsonResponse.Data.Result[0].Value) > 1 {
		result.Status = "Passed"
		result.Message = fmt.Sprintf("%v", jsonResponse.Data.Result[0].Value[1])
	} else {
		result.Status = "Failed"
		result.Message = "No data returned"
	}

	log.Info("Prometheus query check passed for %s", ip.IP)

	return result
}

func (q *QueryChecker) createBaseResult(item string, ip config.IPConfig) checker.CheckResult {
	return checker.CheckResult{
		Component: q.Name(),
		Item:      item,
		Role:      ip.Role,
		IP:        ip.IP,
	}
}

func (q *QueryChecker) createFailedResult(item string, ip config.IPConfig, message string, err error) checker.CheckResult {
	result := q.createBaseResult(item, ip)
	result.Status = "Failed"
	if err != nil {
		result.Message = fmt.Sprintf("%s: %v", message, err)
	} else {
		result.Message = message
	}
	log.Error("%s check failed for %s: %v", item, ip.IP, err)
	return result
}
