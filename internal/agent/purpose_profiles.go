package agent

import (
	"context"

	"github.com/kunchenguid/no-mistakes/internal/types"
)

var mechanicalPurposes = map[string]bool{
	"pr-draft": true,
}

type purposeProfileAgent struct {
	Agent
	profiles     map[string]types.PurposeProfile
	modelLocked  bool
	effortLocked bool
}

func (p purposeProfileAgent) Run(ctx context.Context, opts RunOpts) (*Result, error) {
	profile, ok := p.profiles[opts.Purpose]
	if !ok && mechanicalPurposes[opts.Purpose] {
		profile, ok = p.profiles["mechanical"]
	}
	if ok {
		if !p.modelLocked && profile.Model != "" {
			opts.Model = profile.Model
		}
		if !p.effortLocked && profile.Effort != "" {
			opts.Effort = profile.Effort
		}
	}
	return p.Agent.Run(ctx, opts)
}

func (p purposeProfileAgent) SupportsSessionResume() bool {
	return SupportsSessionResume(p.Agent)
}

func (p purposeProfileAgent) SupportsSessionProvider(provider string) bool {
	return SupportsSessionProvider(p.Agent, provider)
}

func (p purposeProfileAgent) ReportsAgentAttempts() bool {
	return ReportsAgentAttempts(p.Agent)
}

func (p purposeProfileAgent) NeutralizesGateInstructions() bool {
	return NeutralizesGateInstructions(p.Agent)
}

// WithPurposeProfiles applies exact-purpose profiles and the optional
// mechanical fallback. Explicit run-level model/effort locks win per field.
func WithPurposeProfiles(a Agent, profiles map[string]types.PurposeProfile, modelLocked, effortLocked bool) Agent {
	if a == nil || len(profiles) == 0 || (modelLocked && effortLocked) {
		return a
	}
	return purposeProfileAgent{Agent: a, profiles: profiles, modelLocked: modelLocked, effortLocked: effortLocked}
}
