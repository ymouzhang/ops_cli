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
	if checker, ok := m.checkers[queryType]; ok {
		return checker.Check()
	}

	return []checker.CheckResult{{
		Component: queryType,
		Status:    "Failed",
		Message:   "Invalid query type. Use 'query' or 'query_range'",
	}}
}
