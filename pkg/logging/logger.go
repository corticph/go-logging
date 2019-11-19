package logging

import (
	"context"
	"fmt"
	"log"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/getsentry/raven-go"
)

// Logger is a struct containing the functionality of the go-logging library
// This can be used as a global logger, or specified class based logger
type Logger struct {
	initialised bool
	logchannel  LogChannel
	severity    Severity
	client      ElasticClient
	cancel      context.CancelFunc
	procs       int
}

// LogChannel is a wrapper around a chan LogLine
type LogChannel struct {
	channel chan LogLine
}

// Send will send the logline input to the contained channel
// if the channel is not nil
func (lchan LogChannel) Send(logline LogLine) {
	logline.Log()

	if lchan.channel == nil {
		return
	}
	lchan.channel <- logline
}

// NewLogger will initialise and return a new loggger
func NewLogger(severity Severity, procs int) *Logger {
	logger := Logger{
		initialised: true,
		severity:    severity,
		procs:       procs,
	}
	logger.initializeLogger()
	return &logger
}

func (logger *Logger) initializeLogger() {
	channel, cancel := logger.initializeChannel()
	logger.logchannel = LogChannel{
		channel: channel,
	}
	logger.cancel = cancel
}

func (logger *Logger) initializeChannel() (chan LogLine, context.CancelFunc) {
	channel := make(chan LogLine)

	ctx, cancel := context.WithCancel(context.Background())
	for i := 0; i < getIntOrDefault(logger.procs, 100); i++ {
		go func() {
			for {
				select {
				case logline := <-channel:
					glogger.client.send(logline)
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

func (logger *Logger) destroy() *Logger {
	if logger.cancel != nil {
		logger.cancel()
	}
	return NewLogger(INFO, 1)
}

// SetElasticClient will set the contained elastic client using the given input parameters
func (logger *Logger) SetElasticClient(service string, config elasticsearch.Config) error {
	if isValidELSConfig(service, config) {
		logger.client = NewElasticClient(service, config)
		return logger.client.err
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

// Log will log all given error messages which are equal to or above the
// current global severity level. Messages determined above the global
// severity level, will be output to console as well as being sent to
// elasicsearch (messages with `ERROR` or higher will be sent to SentryIO)
func (logger *Logger) Log(msgs ...Message) {
	for _, msg := range msgs {
		if msg == nil {
			return
		}
		if msg.Severity() <= ERROR {
			raven.CaptureError(msg, SentryIOErrorTag(msg.Severity()))
		}
		if msg.Severity() <= glogger.severity {
			glogger.logchannel.Send(NewLogLine(msg))

		}
	}
}

// Debug will log a debug message
func (logger *Logger) Debug(msg string) {
	logger.LogAs(DEBUG, msg)
}

// Debugf will log a debug message
func (logger *Logger) Debugf(format string, a ...interface{}) {
	logger.LogfAs(DEBUG, format, a...)
}

// Info will log an info message
func (logger *Logger) Info(msg string) {
	logger.LogAs(INFO, msg)
}

// Infof will log an info message
func (logger *Logger) Infof(format string, a ...interface{}) {
	logger.LogfAs(INFO, format, a...)
}

// Warn will log a warn message
func (logger *Logger) Warn(msg string) {
	logger.LogAs(WARN, msg)
}

// Warnf will log a warn message
func (logger *Logger) Warnf(format string, a ...interface{}) {
	logger.LogfAs(WARN, format, a...)
}

// Err will log an error message
func (logger *Logger) Err(msg string) {
	logger.LogAs(ERROR, msg)
}

// Errf will log an error message
func (logger *Logger) Errf(format string, a ...interface{}) {
	logger.LogfAs(ERROR, format, a...)
}

// Fatal will log a fatal error message
func (logger *Logger) Fatal(msg string) {
	logger.LogAs(FATAL, msg)
}

// Fatalf will log a formatted fatal error message
func (logger *Logger) Fatalf(format string, a ...interface{}) {
	logger.LogfAs(FATAL, format, a...)
}

// LogAs will log an error message with the given severity level
func (logger *Logger) LogAs(severity Severity, msg string) {
	Log(NewLogMsg(msg, severity))
}

// LogfAs will log an error message with the given severity level
func (logger *Logger) LogfAs(severity Severity, format string, a ...interface{}) {
	Log(NewLogMsg(fmt.Sprintf(format, a...), severity))
}
