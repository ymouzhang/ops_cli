package checker

import (
	"ops_cli/internal/config"
)

type Manager struct {
	checkers map[string]Checker
	config   *config.Config
}

func NewManager(cfg *config.Config) *Manager {
	m := &Manager{
		checkers: make(map[string]Checker),
		config:   cfg,
	}

	m.registerCheckers()
	return m
}

func (m *Manager) registerCheckers() {
	m.checkers["ssh"] = NewSSHChecker(m.config.IPs)
	m.checkers["prometheus"] = NewPrometheusChecker(m.config)
	m.checkers["system"] = NewSystemChecker(m.config)
}

func (m *Manager) Check(component string) []CheckResult {
	if component == "all" {
		return m.checkAll()
	}

	if checker, ok := m.checkers[component]; ok {
		return checker.Check()
	}

	return []CheckResult{{
		Component: component,
		Status:    "Failed",
		Message:   "Unknown component, please exec './ops_cli check --help' for more information",
	}}
}

func (m *Manager) checkAll() []CheckResult {
	var results []CheckResult
	for _, checker := range m.checkers {
		results = append(results, checker.Check()...)
	}
	return results
}
