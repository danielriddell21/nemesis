package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/danielriddell21/nemesis/internal/gui"
)

func Execute(version string) error {
	if err := newRootCmd(version).Execute(); err != nil {
		return fmt.Errorf("execute: %w", err)
	}
	return nil
}

func newRootCmd(version string) *cobra.Command {
	var cfg gui.Config
	var child int

	cmd := &cobra.Command{
		Use:           "nemesis",
		Short:         "A procedurally generated first-person stealth game hunted by an Alien-Isolation-style AI.",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			if cfg.Rec.Recording() {
				return runRecord(cfg)
			}
			if child > 0 {
				return runChild(cfg)
			}
			return runLeader(cfg)
		},
	}

	cmd.Flags().Int64Var(&cfg.Seed, "seed", 0, "world seed (0 picks a random seed)")
	cmd.Flags().IntVar(&cfg.Width, "width", 48, "level width in tiles")
	cmd.Flags().IntVar(&cfg.Height, "height", 32, "level height in tiles")
	cmd.Flags().IntVar(&cfg.Consoles, "consoles", 3, "objective consoles required to unlock the airlock")
	cmd.Flags().BoolVar(&cfg.Visualiser, "visualiser", false, "open the hunter AI visualiser window alongside the game")
	cmd.Flags().StringVar(&cfg.Rec.Path, "record", "", "record a bot-driven AI visualiser demo to this GIF path, then exit")
	cmd.Flags().IntVar(&cfg.Rec.Frames, "record-frames", 0, "frames to capture when recording (0 uses a default)")
	cmd.Flags().IntVar(&child, "child", 0, "internal: run as a coordinated child window")
	_ = cmd.Flags().MarkHidden("child")

	cmd.AddCommand(completionCmd())
	return cmd
}
