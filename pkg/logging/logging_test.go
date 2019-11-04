package logging

import (
	"encoding/json"
	"github.com/elastic/go-elasticsearch/v8"
	"io/ioutil"
	"os"
	"testing"
)

func readESConfigs(path string) map[string]string {

	var config map[string]string

	jsonFile, err := os.Open(path)
	// if we os.Open returns an error then handle it
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)

	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(byteValue, &config)
	if err != nil {
		panic(err)
	}

	return config
}

// make sure we can call a bunch of logs
// without the program blocking execution
func TestNonBlocking(t *testing.T) {

	config := readESConfigs("esconfig.json")

	// use too many processes on purpouse to get some errors back
	err := SetElasticClient(
		1000,
		"test",
		elasticsearch.Config{
			Addresses: []string{config["elk-cloud-addr"]},
			Username:  config["elk-user"],
			Password:  config["elk-pass"],
		})

	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 1000; i++ {
		Info("Sending stress test logs!!")

	}
}
