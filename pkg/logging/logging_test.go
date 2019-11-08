package logging

import (
	"testing"

	"github.com/elastic/go-elasticsearch/v8"
	"gotest.tools/assert"
)

func TestSetLogSeverity(t *testing.T) {
	defer resetLoggingSettings()

	SetLogSeverity(DEBUG)
	assert.Assert(t, logSeverity == DEBUG)
}

func TestSetElasticClient(t *testing.T) {
	defer resetLoggingSettings()

	assert.Assert(t, loggerSet == false)
	assert.Assert(t, SetElasticClient(0, "go-logging-test", elasticsearch.Config{}) == nil)
	assert.Assert(t, loggerSet == true)
}

func TestSetElasticClientErrror(t *testing.T) {
	defer resetLoggingSettings()

	assert.Assert(t, loggerSet == false)
	assert.Assert(t, SetElasticClient(1, "go-logging-test", elasticsearch.Config{
		Username:  "wrong",
		Password:  "credentials",
		Addresses: []string{"http://are.wrong.com"},
	}) == nil)
	assert.Assert(t, loggerSet == true)

	Debug("message")
	Debugf("mess%s", "age")
	Infof("mess%s", "age")
	Err("message")
	Errf("mess%s", "age")
	Fatal("message")
	Fatalf("mess%s", "age")
	Log(nil)

}
