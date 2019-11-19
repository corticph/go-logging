package logging

import (
	"testing"

	"github.com/elastic/go-elasticsearch/v8"
	"gotest.tools/assert"
)

func TestSetLogSeverity(t *testing.T) {
	defer reset(t)

	assert.Assert(t, glogger.severity == INFO)
	SetLogSeverity(DEBUG)
	assert.Assert(t, glogger.severity == DEBUG)
}

func reset(t *testing.T) {
	t.Helper()

	glogger.destroy()
	glogger = NewLogger(INFO, 1)
}

func TestSetGlobalElasticClientTwiceNoParameters(t *testing.T) {
	defer reset(t)

	assert.Assert(t, SetElasticClient(0, "go-logging-test", elasticsearch.Config{}) == nil)
	assert.Assert(t, SetElasticClient(0, "go-logging-test", elasticsearch.Config{}) == nil)
}

func TestSetGlobalElasticClientWithParameters(t *testing.T) {
	defer reset(t)

	assert.Assert(t, SetElasticClient(1, "go-logging-test", elasticsearch.Config{
		Username:  "wrong",
		Password:  "credentials",
		Addresses: []string{"http://are.wrong.com"},
	}) == nil)

	SetLogSeverity(DEBUG)
	Debug("message")
	Debugf("mess%s", "age")
	Info("message")
	Infof("mess%s", "age")
	Warnf("mess%s", "age")
	Err("message")
	Errf("mess%s", "age")
	Fatal("message")
	Fatalf("mess%s", "age")
	Log(nil)
}

func TestWithoutElasticSearchInitialised(t *testing.T) {
	SetLogSeverity(DEBUG)
	Debug("message")
	Debugf("mess%s", "age")
	Info("message")
	Infof("mess%s", "age")
	Warnf("mess%s", "age")
	Err("message")
	Errf("mess%s", "age")
	Fatal("message")
	Fatalf("mess%s", "age")
	LogAs(ERROR, "messagge")
	LogfAs(ERROR, "mess%s", "age")
	Log(nil)
}

func TestNewLogger(t *testing.T) {
	logger := NewLogger(INFO, 1)
	assert.Assert(t, logger.severity == INFO)
}

func TestSetElasticClientTwiceNoParameters(t *testing.T) {
	logger := NewLogger(INFO, 1)

	assert.Assert(t, logger.SetElasticClient("go-logging-test", elasticsearch.Config{}) == nil)
	assert.Assert(t, logger.SetElasticClient("go-logging-test", elasticsearch.Config{}) == nil)
}

func TestSetElasticClientWithParameters(t *testing.T) {
	logger := NewLogger(INFO, 1)

	assert.Assert(t, logger.SetElasticClient("go-logging-test", elasticsearch.Config{
		Username:  "wrong",
		Password:  "credentials",
		Addresses: []string{"http://are.wrong.com"},
	}) == nil)

	SetLogSeverity(DEBUG)
	Debug("message")
	Debugf("mess%s", "age")
	Info("message")
	Infof("mess%s", "age")
	Warnf("mess%s", "age")
	Err("message")
	Errf("mess%s", "age")
	Fatal("message")
	Fatalf("mess%s", "age")
	Log(nil)
}
