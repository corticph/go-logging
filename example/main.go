package main

import (
	"github.com/corticph/go-logging/pkg/logging"
)

func main() {

	pflag.String("elk-cloud-addr", "", "URL of Elastic instance to send monitoring data to")
	pflag.String("elk-service", "", "The name of the service, for easier retrieval down the line. e.g. cart")
	pflag.String("elk-user", "", "Username with create access on the correct Elastic Index")
	pflag.String("elk-pass", "", "Password for provided user")
	pflag.String("elk-customer", "", "The name of the customer, for easier retrieval")

	pflag.Parse()

	if err := viper.BindPFlags(pflag.CommandLine); err != nil {
		panic(err)
	}

	setUpElasticClient()

	logging.Log("Hello world!")

}

// This function *must* run before any calls to logging package
func setUpElasticClient() {
	if err := logging.SetElasticClient(
		viper.GetString("elk-service"),
		viper.GetString("elk-customer"),
		elasticsearch.Config{
			Addresses: []string{viper.GetString("elk-cloud-addr")},
			Username:  viper.GetString("elk-user"),
			Password:  viper.GetString("elk-pass"),
		},
	); err != nil {
		panic(err)
	}
}
