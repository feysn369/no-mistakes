package types

// RunOverrides are explicit, run-scoped agent choices supplied by the caller.
// They travel with the run and never mutate shared daemon configuration.
type RunOverrides struct {
	Agent           AgentName
	Model           string
	Effort          string
	AdaptiveProfile bool
}

// PurposeProfile selects invocation-scoped model and effort tuning for one
// stable pipeline purpose.
type PurposeProfile struct {
	Model  string `yaml:"model"`
	Effort string `yaml:"effort"`
}
