package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/getsentry/raven-go"
)

var (
	logSeverity   = INFO
	esClient      *elasticsearch.Client
	customerIndex string
	serviceName   string
	loggerSet     bool
	logchannel    chan LogLine
	cancel        context.CancelFunc
)

func init() {
	log.SetOutput(os.Stdout)
}

func resetLoggingSettings() {
	if cancel != nil {
		cancel()
	}

	logSeverity = INFO
	esClient = nil
	customerIndex = ""
	serviceName = ""
	loggerSet = false
	logchannel = nil
	cancel = nil
}

// SetLogSeverity will determine which log to output to the console, as well as elasticsearch.
// Anything equal to or higher in severity will be logged. In other words, if the severity
// `ERROR` is passed to this function, the global logging level will be set to `ERROR` and will
// therefore never log `DEBUG`, `INFO` & `WARNING`.
func SetLogSeverity(severity Severity) {
	LogfAs(INFO, "Log level set to: %s", severity.ToString())
	logSeverity = severity
}

// SetElasticClient will create an elasticsearch logger client, which the information given
// on function invokation. You cannot instatiate the elastic client more than once, and any
// attempt of setting it more than once, will produce an error. This is to avoid unwanted
// processor instatiation.
func SetElasticClient(processors int, service string, config elasticsearch.Config) error {
	if !loggerSet { // ensure that logger is only set once per execution
		if err := setElasticClient(service, config); err != nil {
			return err
		}
		logchannel, cancel = newLogListener(processors)
		loggerSet = true
	}
	Info("initialized elastic client without errors")
	return nil
}

func setElasticClient(service string, config elasticsearch.Config) error {
	if isValidELSConfig(service, config) {
		customerIndex = "logs-" + config.Username
		serviceName = service
		var err error
		esClient, err = elasticsearch.NewClient(config)

		return err
	}
	return nil
}

// evaluate if we have all the flags needed to setup the elastic search client.
func isValidELSConfig(service string, config elasticsearch.Config) bool {
	if service == "" || len(config.Addresses) == 0 || config.Username == "" || config.Password == "" {
		Warn("missing parameters for the elastic search client, skipping logging to it.")
		Warnf("%+v", config)
		return false
	}
	return true
}

func newLogListener(procs int) (chan LogLine, context.CancelFunc) {
	channel := make(chan LogLine)
	ctx, cancel := context.WithCancel(context.Background())
	for i := 0; i < getIntOrDefault(procs, 100); i++ {
		go func() {
			for {
				select {
				case logline := <-channel:
					sendToElasticServer(logline)
				case <-ctx.Done():
					log.Println("Stopping logging service, due to a forced cancel")
					return
				}
			}
		}()
	}
	return channel, cancel
}

func getIntOrDefault(i, defaultInt int) int {
	if i == 0 {
		return defaultInt
	}
	return i
}

func sendToElasticServer(event LogLine) {
	logJSON, err := json.Marshal(event)
	if err != nil {
		log.Printf("got an error while marshling event to json: %v", err)
		return
	}
	if esClient == nil {
		return
	}
	res, err := esClient.Index(customerIndex, bytes.NewReader(logJSON))
	if err != nil {
		log.Printf("got an error while sending log to elastic search: %v", err)
		return
	}
	if res.IsError() {
		log.Printf("got an error response after sending logs to elastic search. response was: %v", res)
	}
	res.Body.Close()
}

// Debug will log a debug message
func Debug(msg string) {
	LogAs(DEBUG, msg)
}

// Debugf will log a debug message
func Debugf(format string, a ...interface{}) {
	LogfAs(DEBUG, format, a...)
}

// Info will log an info message
func Info(msg string) {
	LogAs(INFO, msg)
}

// Infof will log an info message
func Infof(format string, a ...interface{}) {
	LogfAs(INFO, format, a...)
}

// Warn will log a warn message
func Warn(msg string) {
	LogAs(WARN, msg)
}

// Warnf will log a warn message
func Warnf(format string, a ...interface{}) {
	LogfAs(WARN, format, a...)
}

// Err will log an error message
func Err(msg string) {
	LogAs(ERROR, msg)
}

// Errf will log an error message
func Errf(format string, a ...interface{}) {
	LogfAs(ERROR, format, a...)
}

// Fatal will log a fatal error message
func Fatal(msg string) {
	LogAs(FATAL, msg)
}

// Fatalf will log a formatted fatal error message
func Fatalf(format string, a ...interface{}) {
	LogfAs(FATAL, format, a...)
}

// LogAs will log an error message with the given severity level
func LogAs(severity Severity, msg string) {
	Log(NewLogMsg(msg, severity))
}

// LogfAs will log an error message with the given severity level
func LogfAs(severity Severity, format string, a ...interface{}) {
	Log(NewLogMsg(fmt.Sprintf(format, a...), severity))
}

// Log will log all given error messages which are equal to or above the
// current global severity level. Messages determined above the global
// severity level, will be output to console as well as being sent to
// elasicsearch (messages with `ERROR` or higher will be sent to SentryIO)
func Log(msgs ...Message) {
	for _, msg := range msgs {
		if msg == nil {
			return
		}
		if msg.Severity() <= ERROR {
			raven.CaptureError(msg, SentryIOErrorTag(msg.Severity()))
		}
		if msg.Severity() <= logSeverity {
			logline := LogLine{
				Timestamp:   time.Now().UTC().Format(time.RFC3339),
				Message:     msg.String(),
				ServiceName: serviceName,
				LogLevel:    msg.Severity().ToString(),
			}
			log.Printf(logline.String())

			if logchannel != nil {
				logchannel <- logline
			}
		}
	}
}

// LogLine is a struct containing all information necessary for sending
// messages with sufficient metadata to elasticsearch
type LogLine struct {
	Timestamp   string `json:"@timestamp"`
	Original    string `json:"event.original"`
	Message     string `json:"message"`
	ServiceName string `json:"service.name"`
	LogLevel    string `json:"log.level"`
}

// String will return a string definition of the LogLine
func (logline LogLine) String() string {
	return fmt.Sprintf("%s [%s] %s", logline.Timestamp, logline.LogLevel, logline.Message)
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
