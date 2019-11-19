package logging

import (
	"fmt"
	"log"
	"time"
)

// LogLine is a struct containing all information necessary for sending
// messages with sufficient metadata to elasticsearch
type LogLine struct {
	Timestamp   string `json:"@timestamp"`
	Original    string `json:"event.original"`
	Message     string `json:"message"`
	ServiceName string `json:"service.name"`
	LogLevel    string `json:"log.level"`
}

func NewLogLine(msg Message) LogLine {
	return LogLine{
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		Message:     msg.String(),
		ServiceName: glogger.client.name,
		LogLevel:    msg.Severity().ToString(),
	}
}

// String will return a string definition of the LogLine
func (logline LogLine) String() string {
	return fmt.Sprintf("%s [%s] %s", logline.Timestamp, logline.LogLevel, logline.Message)
}

// Log will log the logline to console and return the current object
func (logline LogLine) Log() LogLine {
	log.Printf(logline.String())
	return logline
}
