package checker

import (
	"ops_cli/internal/config"
	"ops_cli/pkg/log"
	"ops_cli/pkg/ssh"
)

type SSHChecker struct {
	config []config.IPConfig
}

func NewSSHChecker(cfg []config.IPConfig) *SSHChecker {
	return &SSHChecker{
		config: cfg,
	}
}

func (s *SSHChecker) Name() string {
	return "ssh"
}

func (s *SSHChecker) Check() []CheckResult {
	var results []CheckResult
	for _, ip := range s.config {
		results = append(results, s.checkSSHConnection(ip))
	}
	return results
}

func (s *SSHChecker) checkSSHConnection(ip config.IPConfig) CheckResult {
	log.Info("Checking SSH connection to %s", ip.IP)
	client := ssh.NewClient(ip.IP, ip.User, ip.Password, ip.Port)
	err := client.Connect()

	result := s.createBaseResult("SSH Connection", ip)

	if err != nil {
		return s.createFailedResult("SSH Connection", ip, "SSH connection failed", err)
	}

	// 尝试执行一个简单的命令来验证连接
	if _, err := client.RunCommand("echo 'SSH connection test'"); err != nil {
		client.Close()
		return s.createFailedResult("SSH Connection", ip, "SSH command execution failed", err)
	}

	client.Close()
	result.Status = "Passed"
	result.Message = "SSH connection successful"
	log.Info("SSH connection successful to %s", ip.IP)

	return result
}

func (s *SSHChecker) createBaseResult(item string, ip config.IPConfig) CheckResult {
	return CheckResult{
		Component: s.Name(),
		Item:      item,
		Role:      ip.Role,
		IP:        ip.IP,
	}
}

func (s *SSHChecker) createFailedResult(item string, ip config.IPConfig, message string, err error) CheckResult {
	result := s.createBaseResult(item, ip)
	result.Status = "Failed"
	result.Message = message
	result.Error = err
	log.Error("%s check failed for %s: %v", item, ip.IP, err)
	return result
}
