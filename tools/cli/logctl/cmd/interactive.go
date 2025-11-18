package cmd

import (
	"fafnir/tools/logctl/internal/tui"
	"fmt"

	"github.com/spf13/cobra"
)

// interactiveCmd represents the interactive command
var interactiveCmd = &cobra.Command{
	Use:     "interactive",
	Aliases: []string{"i", "tui"},
	Short:   "Start interactive mode for log querying",
	Long: `Start an interactive TUI session for querying logs from Elasticsearch.
In interactive mode, you can build queries with a beautiful terminal UI powered by Bubble Tea.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runInteractiveMode()
	},
}

func runInteractiveMode() error {
	if err := tui.Start(); err != nil {
		return fmt.Errorf("failed to start interactive mode: %w", err)
	}
	return nil
}
