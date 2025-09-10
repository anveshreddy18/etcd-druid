// Package output provides a shared, consistent interface for CLI and TUI output formatting
// using charmbracelet tools for beautiful, colored output.
package output

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

var (
	// Logger is the shared logger instance
	Logger *log.Logger

	// Styles for consistent formatting
	styles = struct {
		Success lipgloss.Style
		Error   lipgloss.Style
		Info    lipgloss.Style
		Warning lipgloss.Style
		Header  lipgloss.Style
		Key     lipgloss.Style
		Value   lipgloss.Style
	}{
		Success: lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true), // Green
		Error:   lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true),  // Red
		Info:    lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true), // Blue
		Warning: lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true), // Yellow
		Header:  lipgloss.NewStyle().Foreground(lipgloss.Color("13")).Bold(true), // Magenta
		Key:     lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true),  // Cyan
		Value:   lipgloss.NewStyle().Foreground(lipgloss.Color("7")),             // White
	}
)

func init() {
	// Initialize the logger with custom styling
	Logger = log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller:    false,
		ReportTimestamp: false,
		Prefix:          styles.Header.Render("druid") + " ",
	})
}

// Success prints a success message with green checkmark
func Success(message string) {
	Logger.Info(styles.Success.Render("✓ " + message))
}

// Error prints an error message with red X mark
func Error(message string) {
	Logger.Error(styles.Error.Render("✗ " + message))
}

// Info prints an info message with blue info icon
func Info(message string) {
	Logger.Info(styles.Info.Render("ℹ " + message))
}

// Warning prints a warning message with yellow warning icon
func Warning(message string) {
	Logger.Warn(styles.Warning.Render("⚠ " + message))
}

// Header prints a header message in magenta
func Header(message string) {
	Logger.Info(styles.Header.Render("▶ " + message))
}

// KeyValue prints a key-value pair with consistent styling
func KeyValue(key, value string) {
	fmt.Fprintf(os.Stdout, "%s: %s\n",
		styles.Key.Render(key),
		styles.Value.Render(value))
}

// EtcdOperation prints a formatted message for etcd operations
func EtcdOperation(operation, etcdName, namespace string, allNamespaces bool) {
	if !allNamespaces {
		Info(fmt.Sprintf("%s etcd '%s/%s'", operation, namespace, etcdName))
	} else {
		Info(fmt.Sprintf("%s etcds across all namespaces", operation))
	}
}

// EtcdOperationSuccess prints a success message for etcd operations
func EtcdOperationSuccess(operation string) {
	Success(fmt.Sprintf("%s completed successfully", operation))
}

// EtcdOperationError prints an error message for etcd operations
func EtcdOperationError(operation string, err error) {
	Error(fmt.Sprintf("%s operation could not be completed fully: %v", operation, err))
}

// ProgressMessage prints a progress message for ongoing operations
func ProgressMessage(message string) {
	Logger.Info(styles.Info.Render("⏳ " + message))
}

func StartedProgressMessage(message string) {
	Logger.Info(styles.Info.Render("🚀 " + message))
}

// SetVerbose enables or disables verbose output
func SetVerbose(verbose bool) {
	if verbose {
		Logger.SetLevel(log.DebugLevel)
	} else {
		Logger.SetLevel(log.InfoLevel)
	}
}
