package logging

import (
	"bytes"
	"encoding/json"
	"log"

	"github.com/elastic/go-elasticsearch/v8"
)

// ElasticClient is a wrapper aroundd the elasticsearch.Client struct
type ElasticClient struct {
	err      error
	client   *elasticsearch.Client
	name     string
	customer string
}

// NewElasticClient will initialise and returrn a new elastic client
func NewElasticClient(service string, config elasticsearch.Config) ElasticClient {
	client, err := elasticsearch.NewClient(config)
	if err != nil {
		return ElasticClient{
			err: err,
		}
	}
	return ElasticClient{
		customer: "logs-" + config.Username,
		name:     service,
		client:   client,
	}
}

func (client ElasticClient) send(event LogLine) {
	logJSON, err := json.Marshal(event)
	if err != nil {
		log.Printf("got an error while marshling event to json: %v", err)
		return
	}
	if client.client == nil {
		return
	}
	res, err := client.client.Index(client.customer, bytes.NewReader(logJSON))
	if err != nil {
		log.Printf("got an error while sending log to elastic search: %v", err)
		return
	}
	if res.IsError() {
		log.Printf("got an error response after sending logs to elastic search. response was: %v", res)
	}
	res.Body.Close()
}
