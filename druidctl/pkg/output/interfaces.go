package output

import "io"

// Formatter handles styling text for different output types
type Formatter interface {
	FormatSuccess(message string) string
	FormatError(message string, err error) string
	FormatInfo(message string) string
	FormatWarning(message string) string
	FormatHeader(message string) string
	FormatProgress(message string) string
	FormatStart(message string) string
	FormatEtcdOperation(operation, etcdName, namespace string, allNamespaces bool) string
}

// Writer handles output delivery to the underlying logger/output mechanism
type Writer interface {
	LogInfo(message string)
	LogError(message string, keyvals ...interface{})
	LogWarn(message string)
	WriteRaw(message string)
	SetVerbose(verbose bool)
	SetOutput(w io.Writer)
	IsVerbose() bool
}

// Service combines formatting and writing for complete output handling
type Service interface {
	// High-level methods that applications should use
	Success(message string, params ...string)
	Error(message string, err error, params ...string)
	Info(message string, params ...string)
	Warning(message string, params ...string)
	Header(message string, params ...string)
	RawHeader(message string)
	Progress(message string, params ...string)
	Start(message string, params ...string)

	// Configuration
	SetVerbose(verbose bool)
	SetOutput(w io.Writer)
	IsVerbose() bool
}
