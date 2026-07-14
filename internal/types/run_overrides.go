package types

// RunOverrides are explicit, run-scoped agent choices supplied by the caller.
// They travel with the run and never mutate shared daemon configuration.
type RunOverrides struct {
	Agent  AgentName
	Model  string
	Effort string
}
