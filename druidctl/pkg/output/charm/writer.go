// Package charm provides output formatting and writing implementations using charmbracelet libraries
package charm

import (
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/log"
)

// CharmWriter implements output.Writer using charmbracelet/log
type CharmWriter struct {
	logger  *log.Logger
	verbose bool
}

// NewCharmWriter creates a new CharmWriter with default settings
func NewCharmWriter() *CharmWriter {
	w := &CharmWriter{
		logger: log.NewWithOptions(os.Stderr, log.Options{
			ReportCaller:    false,
			ReportTimestamp: false,
		}),
		verbose: false,
	}
	return w
}

// LogInfo logs an informational message
func (w *CharmWriter) LogInfo(message string) {
	w.logger.Info(message)
}

// LogError logs an error message
func (w *CharmWriter) LogError(message string, keyvals ...interface{}) {
	w.logger.Error(message, keyvals...)
}

// LogWarn logs a warning message
func (w *CharmWriter) LogWarn(message string) {
	w.logger.Warn(message)
}

// SetVerbose sets the verbose mode
func (w *CharmWriter) SetVerbose(verbose bool) {
	w.verbose = verbose
	if verbose {
		w.logger.SetLevel(log.DebugLevel)
	} else {
		w.logger.SetLevel(log.InfoLevel)
	}
}

// IsVerbose returns whether verbose mode is enabled
func (w *CharmWriter) IsVerbose() bool {
	return w.verbose
}

// SetOutput sets the output writer
func (w *CharmWriter) SetOutput(output io.Writer) {
	w.logger.SetOutput(output)
}

// WriteRaw writes a message directly to stdout without any logging prefixes
func (w *CharmWriter) WriteRaw(message string) {
	fmt.Println(message)
}
