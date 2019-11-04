package logging

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/getsentry/raven-go"
	"log"
	"os"
	"time"
)

var logSeverity = INFO
var esClient *elasticsearch.Client
var customerIndex string
var serviceName string
var loggerSet bool

var logchannel chan LogLine

func init() {
	log.SetOutput(os.Stdout)
}

func SetLogSeverity(severity Severity) {
	LogfAs(INFO, "Log level set to: %s", severity.ToString())
	logSeverity = severity
}

func SetElasticClient(processors int, service string, config elasticsearch.Config) error {
	if !loggerSet { // ensure that logger is only set once per execution
		if err := setElasticClient(service, config); err != nil {
			return err
		}
		logchannel = newLogListener(processors)
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

func newLogListener(procs int) chan LogLine {
	channel := make(chan LogLine)
	for i := 0; i < getIntOrDefault(procs, 100); i++ {
		go func() {
			for logline := range channel {
				sendToElasticServer(logline)
			}
		}()
	}
	return channel
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

// evaluate if we have all the flags needed to setup the elastic search client.
func isValidELSConfig(service string, config elasticsearch.Config) bool {
	if service == "" || len(config.Addresses) == 0 || config.Username == "" || config.Password == "" {
		Warn("missing parameters for the elastic search client, skipping logging to it.")
		return false
	}
	return true
}

func Debug(msg string) {
	LogAs(DEBUG, msg)
}

func Debugf(format string, a ...interface{}) {
	LogfAs(DEBUG, format, a...)
}

func Info(msg string) {
	LogAs(INFO, msg)
}

func Infof(format string, a ...interface{}) {
	LogfAs(INFO, format, a...)
}

func Warn(msg string) {
	LogAs(WARN, msg)
}

func Warnf(format string, a ...interface{}) {
	LogfAs(WARN, format, a...)
}

func Err(msg string) {
	LogAs(ERROR, msg)
}

func Errf(format string, a ...interface{}) {
	LogfAs(ERROR, format, a...)
}
func LogAs(severity Severity, msg string) {
	Log(NewLogMsg(msg, severity))
}

func LogfAs(severity Severity, format string, a ...interface{}) {
	Log(NewLogMsg(fmt.Sprintf(format, a...), severity))
}

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

			if esClient != nil {
				logchannel <- logline
			}
		}
	}
}

type LogLine struct {
	Timestamp   string `json:"@timestamp"`
	Original    string `json:"event.original"`
	Message     string `json:"message"`
	ServiceName string `json:"service.name"`
	LogLevel    string `json:"log.level"`
}

func (logline LogLine) String() string {
	return fmt.Sprintf("%s [%s] %s", logline.Timestamp, logline.LogLevel, logline.Message)
}

type Message interface {
	Severity() Severity
	String() string
	Error() string
}

type LogMessage struct {
	severity Severity
	text     string
}

func NewLogMsg(text string, severity Severity) LogMessage {
	return LogMessage{
		severity: severity,
		text:     text,
	}
}

func (msg LogMessage) Severity() Severity {
	return msg.severity
}

func (msg LogMessage) String() string {
	return msg.text
}

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
	FATAL Severity = 0

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
