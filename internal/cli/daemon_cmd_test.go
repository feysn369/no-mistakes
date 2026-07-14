package cli

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/kunchenguid/no-mistakes/internal/types"
)

func TestParseSkipPushOptions(t *testing.T) {
	got, err := parseSkipPushOptions([]string{
		"ci.skip",
		"no-mistakes.skip=test,lint",
	})
	if err != nil {
		t.Fatalf("parseSkipPushOptions() error = %v", err)
	}
	want := []types.StepName{types.StepTest, types.StepLint}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parseSkipPushOptions() = %v, want %v", got, want)
	}
}

func TestParseSkipPushOptionsRejectsUnknownStep(t *testing.T) {
	_, err := parseSkipPushOptions([]string{"no-mistakes.skip=test,deploy"})
	if err == nil {
		t.Fatal("expected unknown step to fail")
	}
}

func TestNormalizeNotifyGatePathResolvesLegacyDotGate(t *testing.T) {
	bare := filepath.Join(t.TempDir(), "repo123.git")
	if err := os.MkdirAll(bare, 0o755); err != nil {
		t.Fatal(err)
	}
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(bare); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	}()
	t.Setenv("PWD", ".")

	got, err := normalizeNotifyGatePath(".")
	if err != nil {
		t.Fatalf("normalizeNotifyGatePath: %v", err)
	}
	if got == "." || !filepath.IsAbs(got) {
		t.Fatalf("normalizeNotifyGatePath(.) = %q, want absolute path", got)
	}
	want, err := filepath.EvalSymlinks(bare)
	if err != nil {
		want = bare
	}
	gotResolved, err := filepath.EvalSymlinks(got)
	if err != nil {
		gotResolved = got
	}
	if gotResolved != want {
		t.Fatalf("normalizeNotifyGatePath(.) = %q (resolved %q), want %q", got, gotResolved, want)
	}
}

func TestFormatSkipPushOptions(t *testing.T) {
	got := formatSkipPushOptions([]types.StepName{types.StepTest, types.StepLint})
	want := []string{"no-mistakes.skip=test,lint"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("formatSkipPushOptions() = %v, want %v", got, want)
	}
}

func TestAgentPushOptionRoundTrip(t *testing.T) {
	for _, want := range []types.AgentName{types.AgentClaude, types.AgentCodex} {
		got, err := parseAgentPushOptions([]string{"ci.skip", formatAgentPushOption(want)})
		if err != nil {
			t.Fatalf("parseAgentPushOptions(%q): %v", want, err)
		}
		if got != want {
			t.Fatalf("parseAgentPushOptions(%q) = %q", want, got)
		}
	}
}

func TestParseAgentPushOptionsRejectsConflict(t *testing.T) {
	_, err := parseAgentPushOptions([]string{
		"no-mistakes.agent=claude",
		"no-mistakes.agent=codex",
	})
	if err == nil {
		t.Fatal("expected conflicting run agents to fail")
	}
}

func TestParseRunAgentRejectsUnsupportedAgent(t *testing.T) {
	if _, err := parseRunAgent("auto"); err == nil {
		t.Fatal("expected auto to be rejected for a run-scoped override")
	}
}

func TestRunTuningRequiresExplicitProvider(t *testing.T) {
	if _, err := parseRunTuning("", "gpt-5.5", "medium"); err == nil {
		t.Fatal("provider-specific tuning without --agent should fail")
	}
	got, err := parseRunTuning(types.AgentCodex, "gpt-5.5", "xhigh")
	if err != nil {
		t.Fatal(err)
	}
	if got.Model != "gpt-5.5" || got.Effort != "xhigh" {
		t.Fatalf("unexpected tuning: %#v", got)
	}
	if _, err := parseRunTuning(types.AgentCodex, "gpt-5.5", "max"); err == nil {
		t.Fatal("codex max effort should fail")
	}
	if _, err := parseRunTuning(types.AgentClaude, "opus", "max"); err != nil {
		t.Fatalf("claude max effort should pass: %v", err)
	}
}

func TestModelAndEffortPushOptionsRoundTrip(t *testing.T) {
	options := []string{
		formatStringPushOption(modelPushOptionPrefix, "sonnet"),
		formatStringPushOption(effortPushOptionPrefix, "high"),
	}
	model, err := parseStringPushOption(options, modelPushOptionPrefix, "model")
	if err != nil || model != "sonnet" {
		t.Fatalf("model = %q, err=%v", model, err)
	}
	effort, err := parseStringPushOption(options, effortPushOptionPrefix, "effort")
	if err != nil || effort != "high" {
		t.Fatalf("effort = %q, err=%v", effort, err)
	}
}

func TestIntentPushOptionRoundTrip(t *testing.T) {
	// Multi-line, comma- and colon-bearing intent must survive the
	// line-oriented push-option transport intact.
	intent := "add retry to the uploader\n\nwhy: flaky network, commas, colons: ok"
	opt := formatIntentPushOption(intent)
	if opt == "" {
		t.Fatal("formatIntentPushOption returned empty for a non-empty intent")
	}
	got, err := parseIntentPushOptions([]string{"no-mistakes.skip=test", opt})
	if err != nil {
		t.Fatalf("parseIntentPushOptions() error = %v", err)
	}
	if got != intent {
		t.Fatalf("round-trip mismatch:\n got %q\nwant %q", got, intent)
	}
}

func TestFormatIntentPushOptionEmpty(t *testing.T) {
	if got := formatIntentPushOption("   "); got != "" {
		t.Fatalf("formatIntentPushOption(blank) = %q, want empty", got)
	}
}

func TestParseIntentPushOptionsNone(t *testing.T) {
	got, err := parseIntentPushOptions([]string{"no-mistakes.skip=test", "ci.skip"})
	if err != nil {
		t.Fatalf("parseIntentPushOptions() error = %v", err)
	}
	if got != "" {
		t.Fatalf("parseIntentPushOptions(no intent) = %q, want empty", got)
	}
}
