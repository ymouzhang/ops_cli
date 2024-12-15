package checker

import (
	"fmt"
	"ops_cli/internal/config"
	"ops_cli/pkg/log"
	"ops_cli/pkg/ssh"
	"strconv"
	"strings"
)

// 如果时间比ops时间快3分钟，则认为时间不同步
const timeSyncThreshold = 300

type SystemChecker struct {
	config      *config.Config
	timeResults map[string]int64 // 存储每个IP的时间戳
}

func NewSystemChecker(cfg *config.Config) *SystemChecker {
	return &SystemChecker{
		config:      cfg,
		timeResults: make(map[string]int64),
	}
}

func (s *SystemChecker) Name() string {
	return "system"
}

func (s *SystemChecker) Check() []CheckResult {
	var results []CheckResult

	// 首先检查每个节点的系统时间
	for _, ip := range s.config.IPs {
		results = append(results, s.checkSystemTime(ip))
	}

	// 然后检查时间同步状态
	results = append(results, s.checkTimeSync())

	return results
}

func (s *SystemChecker) checkSystemTime(ip config.IPConfig) CheckResult {
	log.Info("Checking system time for %s", ip.IP)

	client := ssh.NewClient(ip.IP, ip.User, ip.Password, ip.Port)
	if err := client.Connect(); err != nil {
		return s.createFailedResult("System Time", ip, "Failed to establish SSH connection", err)
	}
	defer client.Close()

	// 获取系统时间戳
	output, err := client.RunCommand("date +%s")
	if err != nil {
		return s.createFailedResult("System Time", ip, "Failed to get system time", err)
	}

	timestamp := strings.TrimSpace(output)
	timestampInt, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return s.createFailedResult("System Time", ip, "Failed to parse system time", err)
	}

	// 存储时间戳用于后续的同步检查
	s.timeResults[ip.IP] = timestampInt

	result := s.createBaseResult("System Time", ip)
	result.Status = "Passed"
	result.Message = fmt.Sprintf("System time: %s", timestamp)
	log.Info("System time check passed for %s: %s", ip.IP, timestamp)

	return result
}

func (s *SystemChecker) checkTimeSync() CheckResult {
	log.Info("Checking time synchronization between nodes")

	// 查找 OPS 节点的时间戳
	var opsIP string
	var opsTimestamp int64
	var opsConfig config.IPConfig

	for _, ip := range s.config.IPs {
		if ip.Role == "ops" {
			if timestamp, exists := s.timeResults[ip.IP]; exists {
				opsIP = ip.IP
				opsTimestamp = timestamp
				opsConfig = ip
				break
			}
		}
	}

	if opsIP == "" {
		return s.createFailedResult("Time Sync", opsConfig, "No OPS node time reference found", nil)
	}

	// 检查其他节点与 OPS 节点的时间差
	var nonSyncIPs []string
	for _, ip := range s.config.IPs {
		if ip.Role != "ops" {
			if timestamp, exists := s.timeResults[ip.IP]; exists {
				timeDiff := timestamp - opsTimestamp
				if timeDiff > timeSyncThreshold {
					log.Error("Time difference too large for %s: %d seconds", ip.IP, timeDiff)
					nonSyncIPs = append(nonSyncIPs, ip.IP)
				}
			}
		}
	}

	result := s.createBaseResult("Time Sync", opsConfig)
	if len(nonSyncIPs) > 0 {
		result.Status = "Failed"
		result.Message = fmt.Sprintf("Time not synchronized for nodes: %s", strings.Join(nonSyncIPs, ", "))
		log.Error("Time synchronization check failed: %s", result.Message)
	} else {
		result.Status = "Passed"
		result.Message = "All nodes are time synchronized"
		log.Info("Time synchronization check passed for all nodes")
	}

	return result
}

func (s *SystemChecker) createBaseResult(item string, ip config.IPConfig) CheckResult {
	return CheckResult{
		Component: s.Name(),
		Item:      item,
		Role:      ip.Role,
		IP:        ip.IP,
	}
}

func (s *SystemChecker) createFailedResult(item string, ip config.IPConfig, message string, err error) CheckResult {
	result := s.createBaseResult(item, ip)
	result.Status = "Failed"
	result.Message = message
	if err != nil {
		result.Error = err
		result.Message = fmt.Sprintf("%s: %v", message, err)
	}
	log.Error("%s check failed for %s: %s", item, ip.IP, result.Message)
	return result
}
