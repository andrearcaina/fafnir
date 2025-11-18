package ui

import (
	"encoding/json"
	"fmt"
	"os"

	"fafnir/tools/logctl/internal/types"
)

func PrintLogs(logs []types.LogEntry, format string) {
	if len(logs) == 0 {
		fmt.Println("No logs found.")
		return
	}

	if format == "json" {
		printLogsJSON(logs)
		return
	}

	for _, log := range logs {
		printLog(log)
	}

	fmt.Printf("\nTotal: %d logs\n", len(logs))
}

func printLog(log types.LogEntry) {
	timestamp := TimestampStyle().Render(log.Timestamp.Format("2006-01-02 15:04:05"))
	serviceName := log.Kubernetes.Container.Name
	if serviceName == "" {
		serviceName = "unknown"
	}
	service := ServiceStyle().Render(fmt.Sprintf("[%s]", serviceName))

	fmt.Printf("%s %s %s\n", timestamp, service, log.Message)

	if log.RequestID != "" {
		fmt.Printf("  %s %s\n", FieldStyle().Render("RequestID:"), log.RequestID)
	}
	if log.TraceID != "" {
		fmt.Printf("  %s %s\n", FieldStyle().Render("TraceID:"), log.TraceID)
	}
	if log.UserID != "" {
		fmt.Printf("  %s %s\n", FieldStyle().Render("UserID:"), log.UserID)
	}
	if log.Error != "" {
		fmt.Printf("  %s %s\n", ErrorStyle().Render("Error:"), log.Error)
	}
	fmt.Println()
}

func printLogsJSON(logs []types.LogEntry) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	encoder.Encode(logs)
}

func ExportLogs(logs []types.LogEntry, format, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	if format == "json" {
		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(logs); err != nil {
			return fmt.Errorf("failed to write JSON: %w", err)
		}
	} else {
		for _, log := range logs {
			line := fmt.Sprintf("[%s] [%s] %s\n",
				log.Timestamp.Format("2006-01-02 15:04:05"),
				log.Kubernetes.Container.Name,
				log.Message,
			)
			if _, err := file.WriteString(line); err != nil {
				return fmt.Errorf("failed to write log: %w", err)
			}
		}
	}

	fmt.Printf("Logs exported to %s\n", outputPath)
	return nil
}
