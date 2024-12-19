package checker

import (
	"fmt"
	"io"
	"net/http"
	"ops_cli/internal/config"
	"ops_cli/pkg/log"
	"time"
)

type PrometheusChecker struct {
	config *config.Config
	client *http.Client
}

func NewPrometheusChecker(cfg *config.Config) *PrometheusChecker {
	return &PrometheusChecker{
		config: cfg,
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

func (p *PrometheusChecker) Name() string {
	return "prometheus"
}

func (p *PrometheusChecker) Check() []CheckResult {
	var results []CheckResult

	for _, ip := range p.config.IPs {
		results = append(results, p.checkHealth(ip), p.checkTargets(ip), p.checkFederation(ip))
	}

	return results
}

func (p *PrometheusChecker) checkHealth(ip config.IPConfig) CheckResult {
	log.Info("Checking Prometheus health for %s", ip.IP)

	baseUrl, err := config.GetUrl(ip.IP, ip.Role, config.ComponentPrometheus, config.PathHealth)
	if err != nil {
		return p.createFailedResult("API Health", ip, "Failed to get base url", err)
	}
	log.Debug("Making HTTP request to %s with timeout %v", baseUrl, p.client.Timeout)

	resp, err := p.client.Get(baseUrl)
	result := p.createBaseResult("API Health", ip)

	if err != nil {
		return p.createFailedResult("API Health", ip, "API health check failed", err)
	}
	defer resp.Body.Close()

	log.Debug("Response from %s - Status: %d, Headers: %v", baseUrl, resp.StatusCode, resp.Header)

	if resp.StatusCode != http.StatusOK {
		return p.createFailedResult("API Health", ip, fmt.Sprintf("API returned status code %d", resp.StatusCode), nil)
	}

	result.Status = "Passed"
	result.Message = "API is healthy"
	log.Info("Prometheus health check passed for %s", ip.IP)
	log.Debug("Health check successful - Response time: %v", resp.Header.Get("Date"))

	return result
}

func (p *PrometheusChecker) checkTargets(ip config.IPConfig) CheckResult {
	log.Info("Checking Prometheus targets for %s", ip.IP)

	baseUrl, err := config.GetUrl(ip.IP, ip.Role, config.ComponentPrometheus, config.PathTargets)
	if err != nil {
		return p.createFailedResult("Targets Status", ip, "Failed to get base url", err)
	}
	log.Debug("Fetching targets from %s", baseUrl)

	resp, err := p.client.Get(baseUrl)
	result := p.createBaseResult("Targets Status", ip)

	if err != nil {
		return p.createFailedResult("Targets Status", ip, "Failed to get targets status", err)
	}
	defer resp.Body.Close()

	if body, err := io.ReadAll(resp.Body); err == nil {
		log.Debug("Targets response body: %s", string(body))
	}

	if resp.StatusCode != http.StatusOK {
		return p.createFailedResult("Targets Status", ip, fmt.Sprintf("API returned status code %d", resp.StatusCode), nil)
	}

	result.Status = "Passed"
	result.Message = fmt.Sprintf("Target %s are up", ip.IP)
	log.Info("Prometheus targets check passed for %s", ip.IP)

	return result
}

func (p *PrometheusChecker) checkFederation(ip config.IPConfig) CheckResult {
	log.Info("Checking Prometheus federation for %s", ip.IP)

	baseUrl, err := config.GetUrl(ip.IP, ip.Role, config.ComponentPrometheus, config.PathFederate)
	if err != nil {
		return p.createFailedResult("Federation Status", ip, "Failed to get base url", err)
	}
	url := fmt.Sprintf("%s?match[]=up", baseUrl)
	log.Debug("Fetching federation data from %s", url)

	resp, err := p.client.Get(url)
	result := p.createBaseResult("Federation Status", ip)

	if err != nil {
		return p.createFailedResult("Federation Status", ip, "Failed to get federation status", err)
	}
	defer resp.Body.Close()

	log.Debug("Received response with status code: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		return p.createFailedResult("Federation Status", ip, fmt.Sprintf("API returned status code %d", resp.StatusCode), nil)
	}

	result.Status = "Passed"
	result.Message = "Federation is healthy and accessible"
	log.Info("Prometheus federation check passed for %s", ip.IP)

	return result
}

func (p *PrometheusChecker) createBaseResult(item string, ip config.IPConfig) CheckResult {
	return CheckResult{
		Component: p.Name(),
		Item:      item,
		Role:      ip.Role,
		IP:        ip.IP,
	}
}

func (p *PrometheusChecker) createFailedResult(item string, ip config.IPConfig, message string, err error) CheckResult {
	result := p.createBaseResult(item, ip)
	result.Status = "Failed"
	result.Message = message
	result.Error = err
	log.Error("%s check failed for %s: %v", item, ip.IP, err)
	return result
}
