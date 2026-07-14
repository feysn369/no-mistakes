package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kunchenguid/no-mistakes/internal/types"
)

func TestLoadGlobalPurposeProfilesAndMerge(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	data := `agent: codex
purpose_profiles:
  codex:
    review: {model: gpt-5.5, effort: medium}
    mechanical: {model: gpt-5.5-codex, effort: low}
  claude:
    review: {model: opus, effort: max}
`
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	global, err := LoadGlobal(path)
	if err != nil {
		t.Fatal(err)
	}
	resolved := Merge(global, &RepoConfig{})
	if got := resolved.PurposeProfiles[types.AgentCodex]["mechanical"]; got.Model != "gpt-5.5-codex" || got.Effort != "low" {
		t.Fatalf("codex mechanical profile = %+v", got)
	}
	if got := resolved.PurposeProfiles[types.AgentClaude]["review"]; got.Model != "opus" || got.Effort != "max" {
		t.Fatalf("claude review profile = %+v", got)
	}
}

func TestLoadGlobalRejectsInvalidPurposeProfiles(t *testing.T) {
	tests := []struct {
		name string
		yaml string
		want string
	}{
		{"unsupported provider", "purpose_profiles:\n  opencode:\n    review: {model: gpt-5}\n", "only claude and codex"},
		{"codex max", "purpose_profiles:\n  codex:\n    review: {effort: max}\n", "unsupported effort"},
		{"unsafe model", "purpose_profiles:\n  claude:\n    review: {model: 'sonnet;rm'}\n", "invalid model"},
		{"empty", "purpose_profiles:\n  codex:\n    review: {}\n", "must set model or effort"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.yaml")
			if err := os.WriteFile(path, []byte(tt.yaml), 0o600); err != nil {
				t.Fatal(err)
			}
			_, err := LoadGlobal(path)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("LoadGlobal error = %v, want containing %q", err, tt.want)
			}
		})
	}
}
