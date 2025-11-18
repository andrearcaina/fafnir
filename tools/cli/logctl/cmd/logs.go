package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"errors"
	"fafnir/tools/logctl/internal/elastic"
	"fafnir/tools/logctl/internal/types"
	"fafnir/tools/logctl/internal/ui"
	"fafnir/tools/logctl/internal/utils"
)

var (
	serviceFlag   string
	sinceFlag     string
	untilFlag     string
	followFlag    bool
	searchFlag    string
	limitFlag     int
	formatFlag    string
	outputFlag    string
	intervalFlag  int
	requestIDFlag string
)

// logsCmd represents the logs command
var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Query and display logs from Elasticsearch",
	Long: `Query and display logs from Elasticsearch with various filtering options.

Examples:
  logctl logs --svc auth
  logctl logs --svc user --since 1h
  logctl logs --svc api-gateway --follow
  logctl logs --svc auth --search "failed login" --limit 50
  logctl logs --svc user --since 2024-01-01 --until 2024-01-02
  logctl logs --request-id abc123-def456`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if serviceFlag == "" && requestIDFlag == "" {
			return errors.New("either --svc or --request-id flag is required")
		}

		// validate service
		if serviceFlag != "" {
			serviceFlag = utils.NormalizeService(serviceFlag)
			if err := utils.ValidateService(serviceFlag); err != nil {
				return err
			}
		}

		// validate limit
		if err := utils.ValidateLimit(limitFlag); err != nil {
			return err
		}

		// validate interval for follow mode
		if followFlag {
			if err := utils.ValidateInterval(intervalFlag); err != nil {
				return err
			}
		}

		// parse time ranges
		var since, until time.Time
		var err error

		if sinceFlag != "" {
			since, err = utils.ParseTimeFlag(sinceFlag)
			if err != nil {
				return fmt.Errorf("invalid --since flag: %w", err)
			}
		} else {
			since = time.Now().Add(-5 * time.Minute) // default to last 5 minutes
		}

		if untilFlag != "" {
			until, err = utils.ParseTimeFlag(untilFlag)
			if err != nil {
				return fmt.Errorf("invalid --until flag: %w", err)
			}
		} else {
			until = time.Now().Add(24 * time.Hour) // handle timezone differences
		}

		// validate format and output flags
		if (formatFlag != "" && outputFlag == "") || (outputFlag != "" && formatFlag == "") {
			return errors.New("both --format and --output flags must be set together")
		}

		if formatFlag != "" {
			formatFlag = strings.ToLower(formatFlag)
			if formatFlag != "json" && formatFlag != "text" {
				return errors.New("format must be either 'json' or 'text'")
			}
		}

		// build query options
		queryOpts := &types.QueryOptions{
			Service:   serviceFlag,
			Since:     since,
			Until:     until,
			Search:    searchFlag,
			Limit:     limitFlag,
			RequestID: requestIDFlag,
		}

		// create Elasticsearch client
		client, err := elastic.NewClient()
		if err != nil {
			return fmt.Errorf("failed to create Elasticsearch client: %w", err)
		}

		if followFlag {
			return tailLogs(client, queryOpts, intervalFlag)
		}

		// query logs once
		logs, err := client.QueryLogs(queryOpts)
		if err != nil {
			return fmt.Errorf("failed to query logs: %w", err)
		}

		// handle output
		if outputFlag != "" {
			return ui.ExportLogs(logs, formatFlag, outputFlag)
		}

		ui.PrintLogs(logs, formatFlag)
		return nil
	},
}

func init() {
	logsCmd.Flags().StringVarP(&serviceFlag, "svc", "s", "", "Service name to filter logs (e.g., auth, user, stock)")
	logsCmd.Flags().StringVar(&sinceFlag, "since", "", "Start time for log query (e.g., 1h, 30m, 2024-01-01, now-1h)")
	logsCmd.Flags().StringVar(&untilFlag, "until", "", "End time for log query (e.g., 2024-01-02, now)")
	logsCmd.Flags().BoolVarP(&followFlag, "follow", "f", false, "Follow log output (tail mode)")
	logsCmd.Flags().StringVar(&searchFlag, "search", "", "Search keyword in log messages")
	logsCmd.Flags().IntVar(&limitFlag, "limit", 100, "Maximum number of log entries to return")
	logsCmd.Flags().StringVar(&formatFlag, "format", "", "Output format (json, text)")
	logsCmd.Flags().StringVarP(&outputFlag, "output", "o", "", "Output file path")
	logsCmd.Flags().IntVarP(&intervalFlag, "interval", "i", 2, "Polling interval in seconds for follow mode")
	logsCmd.Flags().StringVar(&requestIDFlag, "request-id", "", "Filter logs by request ID")
}

func tailLogs(client *elastic.Client, opts *types.QueryOptions, interval int) error {
	lastTimestamp := opts.Since
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	fmt.Printf("Following logs for service '%s' (press Ctrl+C to stop)...\n\n", opts.Service)

	for {
		// update time range to fetch only new logs
		opts.Since = lastTimestamp
		opts.Until = time.Now().Add(24 * time.Hour) // handle timezone differences

		logs, err := client.QueryLogs(opts)
		if err != nil {
			fmt.Printf("Error querying logs: %v\n", err)
			<-ticker.C
			continue
		}

		if len(logs) > 0 {
			ui.PrintLogs(logs, "")
			// update last timestamp to the most recent log
			lastTimestamp = logs[len(logs)-1].Timestamp
		}

		<-ticker.C
	}
}
