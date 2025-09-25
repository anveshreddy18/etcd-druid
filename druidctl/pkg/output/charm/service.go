// Package charm provides output formatting and writing implementations using charmbracelet libraries
package charm

import (
	"fmt"
	"io"
)

// CharmService implements output.Service using CharmFormatter and CharmWriter
type CharmService struct {
	formatter *CharmFormatter
	writer    *CharmWriter
}

// NewCharmService creates a new CharmService with default settings
func NewCharmService() *CharmService {
	return &CharmService{
		formatter: NewCharmFormatter(),
		writer:    NewCharmWriter(),
	}
}

func constructPrefixFromParams(params ...string) string {
	prefix := ""
	// Check if both etcdName and namespace are provided
	if len(params) >= 2 {
		etcdName := params[0]
		namespace := params[1]
		prefix = fmt.Sprintf("[%s/%s] ", namespace, etcdName)
	}
	return prefix
}

// Success displays a success message
func (s *CharmService) Success(message string, params ...string) {
	s.writer.LogInfo(s.formatter.FormatHeader(constructPrefixFromParams(params...)) + s.formatter.FormatSuccess(message))
}

// Error displays an error message
func (s *CharmService) Error(message string, err error, params ...string) {
	s.writer.LogError(s.formatter.FormatHeader(constructPrefixFromParams(params...))+s.formatter.FormatError(message, err), "error", err)
}

// Info displays an informational message
func (s *CharmService) Info(message string, params ...string) {
	s.writer.LogInfo(s.formatter.FormatHeader(constructPrefixFromParams(params...)) + s.formatter.FormatInfo(message))
}

// Warning displays a warning message
func (s *CharmService) Warning(message string, params ...string) {
	s.writer.LogWarn(s.formatter.FormatHeader(constructPrefixFromParams(params...)) + s.formatter.FormatWarning(message))
}

// Header displays a header message
func (s *CharmService) Header(message string, params ...string) {
	s.writer.LogInfo(s.formatter.FormatHeader(constructPrefixFromParams(params...)) + s.formatter.FormatHeader(message))
}

// Progress displays a progress message
func (s *CharmService) Progress(message string, params ...string) {
	s.writer.LogInfo(s.formatter.FormatHeader(constructPrefixFromParams(params...)) + s.formatter.FormatProgress(message))
}

// Start displays a start message
func (s *CharmService) Start(message string, params ...string) {
	s.writer.LogInfo(s.formatter.FormatHeader(constructPrefixFromParams(params...)) + s.formatter.FormatStart(message))
}

// SetVerbose enables or disables verbose mode
func (s *CharmService) SetVerbose(verbose bool) {
	s.writer.SetVerbose(verbose)
}

// SetOutput sets the output writer
func (s *CharmService) SetOutput(w io.Writer) {
	s.writer.SetOutput(w)
}

// IsVerbose returns whether verbose mode is enabled
func (s *CharmService) IsVerbose() bool {
	return s.writer.IsVerbose()
}
