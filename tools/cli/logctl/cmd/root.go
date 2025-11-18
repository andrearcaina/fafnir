package cmd

import (
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "logctl",
	Short: "A CLI tool to view and manage Elasticsearch logs",
	Long: `logctl is a CLI tool designed to help developers view and manage logs from Elasticsearch.
It provides both interactive and command-line modes for querying, filtering, and tailing logs.

Examples:
  logctl logs --svc auth
  logctl logs --svc user --since 1h
  logctl logs --svc stock --follow
  logctl -i`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(logsCmd)
	rootCmd.AddCommand(interactiveCmd)
}
