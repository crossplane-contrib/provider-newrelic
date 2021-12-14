package nr

import (
	"github.com/newrelic/newrelic-client-go/newrelic"
	"github.com/newrelic/newrelic-client-go/pkg/config"
	"github.com/newrelic/newrelic-client-go/pkg/nerdgraph"
)

// GetNewRelicClient gets a new client
// https://github.com/newrelic/newrelic-client-go
func GetNewRelicClient(newRelicAPIKey string) (client *newrelic.NewRelic, err error) {
	// Initialize the client.
	client, err = newrelic.New(newrelic.ConfigPersonalAPIKey(newRelicAPIKey))
	return client, err
}

// GetNerdGraphClient gets a new client
// https://github.com/newrelic/newrelic-client-go
func GetNerdGraphClient(newRelicAPIKey string) (client nerdgraph.NerdGraph, err error) {
	// Initialize the client.
	cfg := config.New()
	cfg.PersonalAPIKey = newRelicAPIKey
	client = nerdgraph.New(cfg)
	return client, err
}
