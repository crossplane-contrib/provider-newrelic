package nr

import (
	"github.com/newrelic/newrelic-client-go/v2/newrelic"
	"github.com/newrelic/newrelic-client-go/v2/pkg/config"
	"github.com/newrelic/newrelic-client-go/v2/pkg/nerdgraph"
	"github.com/newrelic/newrelic-client-go/v2/pkg/region"
)

// GetNewRelicClient gets a new client
// https://github.com/newrelic/newrelic-client-go
func GetNewRelicClient(newRelicAPIKey string, regionStr string) (client *newrelic.NewRelic, err error) {
	// Initialize the client.
	client, err = newrelic.New(newrelic.ConfigPersonalAPIKey(newRelicAPIKey), newrelic.ConfigRegion(regionStr))
	return client, err
}

// GetNerdGraphClient gets a new client
// https://github.com/newrelic/newrelic-client-go
func GetNerdGraphClient(newRelicAPIKey string, regionStr string) (client nerdgraph.NerdGraph, err error) {
	// Initialize the client.
	cfg := config.New()
	cfg.PersonalAPIKey = newRelicAPIKey

	regName, err := region.Parse(regionStr)
	if err == nil {
		reg, err := region.Get(regName)
		if err == nil {
			cfg.SetRegion(reg)
		}
	}
	client = nerdgraph.New(cfg)
	return client, err
}
