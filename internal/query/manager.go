package query

import (
	"ops_cli/internal/checker"
	"ops_cli/internal/config"
)

type Manager struct {
	checkers map[string]checker.Checker
	config   *config.Config
}

func NewManager(cfg *config.Config) *Manager {
	m := &Manager{
		checkers: make(map[string]checker.Checker),
		config:   cfg,
	}

	m.registerCheckers()
	return m
}

func (m *Manager) registerCheckers() {
	m.checkers["query"] = NewQueryChecker(m.config)
	m.checkers["query_range"] = NewQueryRangeChecker(m.config)
}

func (m *Manager) Check(queryType string) []checker.CheckResult {
	if queryType == "all" {
		return m.checkAll()
	}

	if checker, ok := m.checkers[queryType]; ok {
		return checker.Check()
	}

	return []checker.CheckResult{{
		Component: queryType,
		Status:    "Failed",
		Message:   "Unknown query type, please exec './ops_cli query --help' for more information",
	}}
}

func (m *Manager) checkAll() []checker.CheckResult {
	var results []checker.CheckResult
	for _, checker := range m.checkers {
		results = append(results, checker.Check()...)
	}
	return results
}
