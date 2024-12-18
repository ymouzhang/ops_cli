package checker

type CheckResult struct {
	Component string
	Item      string
	Status    string
	Message   string
	Error     error
	Role      string
	IP        string
}

type Checker interface {
	Name() string
	Check() []CheckResult
}
