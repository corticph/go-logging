package main

import (
	"fmt"
	"github.com/corticph/go-logging/pkg/logging"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"time"
)

func main() {

	pflag.String("elk-cloud-addr", "", "URL of Elastic instance to send monitoring data to")
	pflag.String("elk-service", "", "The name of the service, for easier retrieval down the line. e.g. cart")
	pflag.String("elk-user", "", "Username with create access on the correct Elastic Index")
	pflag.String("elk-pass", "", "Password for provided user")
	pflag.Int("log-level", int(logging.INFO), "set the log level to display")
	pflag.Int("processes", 0, "how many processes to use")
	pflag.Bool("stress-test", false, "if to perform some stress test (send a shit ton of logs to elastic search)")

	pflag.Parse()

	if err := viper.BindPFlags(pflag.CommandLine); err != nil {
		panic(err)
	}

	logging.SetLogSeverity(logging.Severity(viper.GetInt("log-level")))

	setUpElasticClient()

	now := time.Now()
	logging.Info("Hello info!")
	logging.Warn("Hello warn")
	logging.Err("Hello error")
	logging.LogAs(logging.FATAL, "Hello fatal!")
	logging.Info("More logging!")

	if viper.GetBool("stress-test") {
		for i := 0; i < 1000; i++ {
			time.Sleep(1)
			logging.Info("Sending stress test logs!!")

		}

	}

	fmt.Println("Sent all log messages to elasticsearch:", time.Since(now))
}

// This function *must* run before any calls to logging package
func setUpElasticClient() {
	if err := logging.SetElasticClient(
		viper.GetInt("processes"),
		viper.GetString("elk-service"),
		elasticsearch.Config{
			Addresses: []string{viper.GetString("elk-cloud-addr")},
			Username:  viper.GetString("elk-user"),
			Password:  viper.GetString("elk-pass"),
		},
	); err != nil {
		panic(err)
	}
}
