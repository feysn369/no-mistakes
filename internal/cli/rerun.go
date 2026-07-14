package cli

import (
	"context"
	"fmt"

	"github.com/kunchenguid/no-mistakes/internal/daemon"
	"github.com/kunchenguid/no-mistakes/internal/git"
	"github.com/kunchenguid/no-mistakes/internal/ipc"
	"github.com/spf13/cobra"
)

func newRerunCmd() *cobra.Command {
	var agentValue string
	var modelValue string
	var effortValue string
	var adaptiveProfile bool
	cmd := &cobra.Command{
		Use:   "rerun",
		Short: "Rerun the pipeline for the current branch",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return trackCommand("rerun", func() error {
				agentName, err := parseRunAgent(agentValue)
				if err != nil {
					return err
				}
				overrides, err := parseRunTuning(agentName, modelValue, effortValue, adaptiveProfile)
				if err != nil {
					return err
				}
				p, d, err := openResources()
				if err != nil {
					return err
				}
				defer d.Close()

				repo, err := findRepo(d)
				if err != nil {
					return err
				}

				branch, err := git.CurrentBranch(context.Background(), ".")
				if err != nil {
					return fmt.Errorf("get current branch: %w", err)
				}
				if branch == "HEAD" {
					return fmt.Errorf("not on a branch")
				}

				if err := daemon.EnsureDaemon(p); err != nil {
					return fmt.Errorf("start daemon: %w", err)
				}

				client, err := ipc.Dial(p.Socket())
				if err != nil {
					return fmt.Errorf("connect to daemon: %w", err)
				}
				defer client.Close()

				var result ipc.RerunResult
				if err := client.Call(ipc.MethodRerun, &ipc.RerunParams{RepoID: repo.ID, Branch: branch, Agent: overrides.Agent, Model: overrides.Model, Effort: overrides.Effort, AdaptiveProfile: overrides.AdaptiveProfile}, &result); err != nil {
					return fmt.Errorf("rerun pipeline: %w", err)
				}

				fmt.Fprintf(cmd.OutOrStdout(), "  %s Rerun started for %s %s\n", sGreen.Render("✓"), branch, sDim.Render(result.RunID))
				return nil
			})
		},
	}
	cmd.Flags().StringVar(&agentValue, "agent", "", "pipeline agent for this rerun only (claude or codex); defaults to the previous run's explicit selection")
	cmd.Flags().StringVar(&modelValue, "model", "", "model for this rerun only (requires --agent); defaults to the previous run's explicit selection")
	cmd.Flags().StringVar(&effortValue, "effort", "", "reasoning effort for this rerun only (requires --agent); defaults to the previous run's explicit selection")
	cmd.Flags().BoolVar(&adaptiveProfile, "adaptive-profile", false, "treat model/effort as a baseline and allow configured purpose profiles to override them")
	return cmd
}
