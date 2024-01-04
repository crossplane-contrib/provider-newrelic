package nr

import (
	"context"
	"strconv"
	"strings"

	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/newrelic/newrelic-client-go/v2/newrelic"
	"github.com/newrelic/newrelic-client-go/v2/pkg/config"
	"github.com/newrelic/newrelic-client-go/v2/pkg/nerdgraph"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apisv1alpha1 "github.com/crossplane-contrib/provider-newrelic/apis/v1alpha1"
)

const (
	errGetCreds     = "cannot get credentials"
	errGetAccountID = "cannot get accountId from ProviderConfig"
)

// ExtractNewRelicAccountID gets the accountID from the provider config
func ExtractNewRelicAccountID(pc *apisv1alpha1.ProviderConfig) (account int, err error) {
	accountID := pc.Spec.AccountID
	account, err = strconv.Atoi(accountID)
	if accountID == "" || err != nil {
		return 0, errors.Wrap(err, errGetAccountID)
	}
	return account, nil
}

func ExtractNewRelicCredentials(ctx context.Context, kube client.Client, pc *apisv1alpha1.ProviderConfig) (client *newrelic.NewRelic, err error) {
	cd := pc.Spec.Credentials
	data, err := resource.CommonCredentialExtractor(ctx, cd.Source, kube, cd.CommonCredentialSelectors)
	if err != nil {
		return nil, errors.Wrap(err, errGetCreds)
	}

	// Extract the region
	region := pc.Spec.Region

	// Create a client using "NEW_RELIC_API_KEY"
	return GetNewRelicClient(strings.TrimSpace(string(data)), region)
}

// GetNewRelicClient gets a new client
// https://github.com/newrelic/newrelic-client-go
// https://pkg.go.dev/github.com/newrelic/newrelic-client-go/v2/pkg/config@v2.23.0#ConfigOption
func GetNewRelicClient(newRelicAPIKey string, region *string) (client *newrelic.NewRelic, err error) {

	var options []newrelic.ConfigOption
	options = append(options,
		newrelic.ConfigPersonalAPIKey(newRelicAPIKey),
	)
	// Add the region, if set
	if region != nil {
		options = append(options, newrelic.ConfigRegion(*region))
	}

	// Initialize the client.
	client, err = newrelic.New(options...)
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
