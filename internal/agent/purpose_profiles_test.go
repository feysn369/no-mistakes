package agent

import (
	"context"
	"testing"

	"github.com/kunchenguid/no-mistakes/internal/types"
)

func TestPurposeProfilesExactAndMechanicalFallback(t *testing.T) {
	profiles := map[string]types.PurposeProfile{
		"review":     {Model: "gpt-5.5", Effort: "high"},
		"mechanical": {Model: "gpt-5.5-codex", Effort: "low"},
	}
	for _, tc := range []struct {
		purpose string
		model   string
		effort  string
	}{
		{"review", "gpt-5.5", "high"},
		{"pr-draft", "gpt-5.5-codex", "low"},
		{"test-evidence", "", ""},
	} {
		inner := &recordingAgent{name: "codex", resumable: true}
		wrapped := WithPurposeProfiles(inner, profiles, false, false)
		session := &SessionRef{ID: "thread-1", Agent: "codex"}
		if _, err := wrapped.Run(context.Background(), RunOpts{Prompt: "do it", Purpose: tc.purpose, Session: session}); err != nil {
			t.Fatal(err)
		}
		if inner.gotOpts.Model != tc.model || inner.gotOpts.Effort != tc.effort {
			t.Fatalf("purpose %s tuning = %q/%q, want %q/%q", tc.purpose, inner.gotOpts.Model, inner.gotOpts.Effort, tc.model, tc.effort)
		}
		if inner.gotOpts.Session != session || !SupportsSessionResume(wrapped) {
			t.Fatalf("purpose %s lost resume state/capability", tc.purpose)
		}
	}
}

func TestPurposeProfilesExplicitRunLocksWinPerField(t *testing.T) {
	profiles := map[string]types.PurposeProfile{"review": {Model: "profile-model", Effort: "low"}}
	inner := &recordingAgent{name: "claude"}
	wrapped := WithPurposeProfiles(inner, profiles, true, false)
	if _, err := wrapped.Run(context.Background(), RunOpts{Purpose: "review"}); err != nil {
		t.Fatal(err)
	}
	if inner.gotOpts.Model != "" || inner.gotOpts.Effort != "low" {
		t.Fatalf("per-field locks produced %q/%q, want empty/low", inner.gotOpts.Model, inner.gotOpts.Effort)
	}
}
