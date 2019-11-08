package logging

import (
	"fmt"
	"log"
	"os"

	"github.com/elastic/go-elasticsearch/v8"
)

var (
	glogger *Logger
)

func init() {
	log.SetOutput(os.Stdout)
	glogger = NewLogger(INFO, 1)
}

// Message is an interface representing a log message
type Message interface {
	Severity() Severity
	String() string
	Error() string
}

// LogMessage implements the Message interface and is the primary struct representing log messages
type LogMessage struct {
	severity Severity
	text     string
}

// NewLogMsg will instatiate and return a new LogMessage
func NewLogMsg(text string, severity Severity) LogMessage {
	return LogMessage{
		severity: severity,
		text:     text,
	}
}

// Severity will return the severity
func (msg LogMessage) Severity() Severity {
	return msg.severity
}

// String will return the text of the log message
func (msg LogMessage) String() string {
	return msg.text
}

// Error will return the text of the log message
func (msg LogMessage) Error() string {
	return msg.String()
}

// SetLogSeverity will determine which log to output to the console, as well as elasticsearch.
// Anything equal to or higher in severity will be logged. In other words, if the severity
// `ERROR` is passed to this function, the global logging level will be set to `ERROR` and will
// therefore never log `DEBUG`, `INFO` & `WARNING`.
func SetLogSeverity(severity Severity) {
	glogger.severity = severity
	glogger.LogfAs(INFO, "Log level set to: %s", severity.ToString())
}

// SetElasticClient will create an elasticsearch logger client, which the information given
// on function invokation. You cannot instatiate the elastic client more than once, and any
// attempt of setting it more than once, will produce an error. This is to avoid unwanted
// processor instatiation.
func SetElasticClient(processors int, service string, config elasticsearch.Config) error {
	glogger.destroy()
	glogger = NewLogger(glogger.severity, processors)
	return glogger.SetElasticClient(service, config)
}

// Debug will log a debug message
func Debug(msg string) {
	glogger.Debug(msg)
}

// Debugf will log a debug message
func Debugf(format string, a ...interface{}) {
	glogger.Debugf(format, a...)
}

// Info will log an info message
func Info(msg string) {
	glogger.Info(msg)
}

// Infof will log an info message
func Infof(format string, a ...interface{}) {
	glogger.Infof(format, a...)
}

// Warn will log a warn message
func Warn(msg string) {
	glogger.Warn(msg)
}

// Warnf will log a warn message
func Warnf(format string, a ...interface{}) {
	glogger.Warnf(format, a...)
}

// Err will log an error message
func Err(msg string) {
	glogger.Err(msg)
}

// Errf will log an error message
func Errf(format string, a ...interface{}) {
	glogger.Errf(format, a...)
}

// Fatal will log a fatal error message
func Fatal(msg string) {
	glogger.Fatal(msg)
}

// Fatalf will log a formatted fatal error message
func Fatalf(format string, a ...interface{}) {
	glogger.Fatalf(format, a...)
}

// LogAs will log an error message with the given severity level
func LogAs(severity Severity, msg string) {
	glogger.Log(NewLogMsg(msg, severity))
}

// LogfAs will log an error message with the given severity level
func LogfAs(severity Severity, format string, a ...interface{}) {
	glogger.Log(NewLogMsg(fmt.Sprintf(format, a...), severity))
}

// Log will log all given error messages which are equal to or above the
// current global severity level. Messages determined above the global
// severity level, will be output to console as well as being sent to
// elasicsearch (messages with `ERROR` or higher will be sent to SentryIO)
func Log(msgs ...Message) {
	glogger.Log(msgs...)
}

// Severity represents the logging message severity
type Severity int

// ToString will transform a Severity integer to it's
// string representation, conforming to the error level
// of sentry IO.
func (severity Severity) ToString() string {
	switch severity {
	case FATAL:
		return "fatal"
	case ERROR:
		return "error"
	case WARN:
		return "warning"
	case INFO:
		return "info"
	case DEBUG:
		return "debug"
	default:
		return "unknown"
	}
}

var (
	// FATAL is an msgor severe enough to end an application
	FATAL Severity

	// ERROR is an msgor which is unexpected, but not severe enough to exit the application
	ERROR Severity = 1

	// WARN is typically used for warning users that something that they might not expect has occurred
	WARN Severity = 2

	// INFO is returned for common logging
	INFO Severity = 3

	// DEBUG is any information which could be helpful for debugging an application
	// but is not necessary to log for anyone other than a developer / systems admin
	DEBUG Severity = 4

	// EMPTY represents a description of nothing
	EMPTY = ""
)

// SentryIOErrorTag returns a value used for sending an error level
// tag to SentryIO together with the CaptureError(AndWait) functions
// based on the severity input
func SentryIOErrorTag(severity Severity) map[string]string {
	return map[string]string{
		"level": severity.ToString(),
	}
}
