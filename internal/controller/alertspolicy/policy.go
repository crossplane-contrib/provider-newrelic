/*
Copyright 2020 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package alertspolicy

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/newrelic/newrelic-client-go/newrelic"
	"github.com/newrelic/newrelic-client-go/pkg/alerts"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane-contrib/provider-newrelic/apis/alertspolicy/v1alpha1"
	apisv1alpha1 "github.com/crossplane-contrib/provider-newrelic/apis/v1alpha1"
	nr "github.com/crossplane-contrib/provider-newrelic/internal/clients"
)

const (
	errNotPolicy    = "managed resource is not a Policy custom resource"
	errTrackPCUsage = "cannot track ProviderConfig usage"
	errGetPC        = "cannot get ProviderConfig"
	errGetCreds     = "cannot get credentials"
	errGetAccountID = "cannot get accountId from ProviderConfig"
)

// Setup adds a controller that reconciles Policy managed resources.
func Setup(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter) error {
	name := managed.ControllerName(v1alpha1.AlertsPolicyGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewController(rl),
		}).
		For(&v1alpha1.AlertsPolicy{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.AlertsPolicyGroupVersionKind),
			managed.WithExternalConnecter(&connector{
				kube:  mgr.GetClient(),
				usage: resource.NewProviderConfigUsageTracker(mgr.GetClient(), &apisv1alpha1.ProviderConfigUsage{}),
			}),
			managed.WithConnectionPublishers(),
			managed.WithPollInterval(30*time.Minute),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithInitializers(),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

// A connector is expected to produce an ExternalClient when its Connect method
// is called.
type connector struct {
	kube  client.Client
	usage resource.Tracker
}

// Connect typically produces an ExternalClient by:
// 1. Tracking that the managed resource is using a ProviderConfig.
// 2. Getting the managed resource's ProviderConfig.
// 3. Getting the credentials specified by the ProviderConfig.
// 4. Using the credentials to form a client.
func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.AlertsPolicy)
	if !ok {
		return nil, errors.New(errNotPolicy)
	}

	if err := c.usage.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, errTrackPCUsage)
	}

	pc := &apisv1alpha1.ProviderConfig{}
	if err := c.kube.Get(ctx, types.NamespacedName{Name: cr.GetProviderConfigReference().Name}, pc); err != nil {
		return nil, errors.Wrap(err, errGetPC)
	}

	cd := pc.Spec.Credentials
	data, err := resource.CommonCredentialExtractor(ctx, cd.Source, c.kube, cd.CommonCredentialSelectors)
	if err != nil {
		return nil, errors.Wrap(err, errGetCreds)
	}

	account := pc.Spec.AccountID
	accountID, err := strconv.Atoi(account)
	if account == "" || err != nil {
		return nil, errors.Wrap(err, errGetAccountID)
	}

	// Create a client using "NEW_RELIC_API_KEY"
	nrClient, err := nr.GetNewRelicClient(strings.TrimSpace(string(data)))
	if err != nil {
		return nil, err
	}

	return &external{client: nrClient, kube: c.kube, accountID: accountID}, nil
}

// An ExternalClient observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type external struct {
	client    *newrelic.NewRelic
	kube      client.Client
	accountID int
}

func (c *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.AlertsPolicy)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotPolicy)
	}

	// Get the policy by ID, or name since the names should be unique
	policy, err := c.GetAlertsPolicyByIDOrName(ctx, cr)

	if err != nil {
		return managed.ExternalObservation{}, err
	}

	if policy.Name == "" {
		return managed.ExternalObservation{ResourceExists: false}, err
	}

	// Set the ID, if not set
	c.SetExternalNameIfNotSet(ctx, cr, policy)

	// Update the status
	cr.Status.SetConditions(xpv1.Available())
	cr.Status.AtProvider = v1alpha1.AlertsPolicyObservation{
		ID: policy.ID,
	}

	cr.SetConditions(xpv1.Available())
	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  IsUpToDate(cr, *policy),
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.AlertsPolicy)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotPolicy)
	}
	cr.SetConditions(xpv1.Creating())

	input := alerts.AlertsPolicyInput{
		IncidentPreference: alerts.AlertsIncidentPreference(cr.Spec.ForProvider.IncidentPreference),
		Name:               cr.Spec.ForProvider.Name,
	}

	response, err := c.client.Alerts.CreatePolicyMutationWithContext(ctx, c.accountID, input)
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	// Set the ID
	cr.SetConditions(xpv1.Available())
	c.SetExternalNameIfNotSet(ctx, cr, response)

	// Assign to channels
	policyID, err := strconv.Atoi(response.ID)
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	_, err = c.client.Alerts.UpdatePolicyChannelsWithContext(ctx, policyID, cr.Spec.ForProvider.ChannelIDs)
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	cr.SetConditions(xpv1.Available())
	cr.Status.AtProvider.DeepCopyInto(&cr.Status.AtProvider)

	return managed.ExternalCreation{}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.AlertsPolicy)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotPolicy)
	}

	// Update the policy itself
	policy := alerts.AlertsPolicyUpdateInput{
		IncidentPreference: alerts.AlertsIncidentPreference(cr.Spec.ForProvider.IncidentPreference),
		Name:               cr.Spec.ForProvider.Name,
	}
	_, err := c.client.Alerts.UpdatePolicyMutation(c.accountID, cr.Spec.ForProvider.ID, policy)
	if err != nil {
		return managed.ExternalUpdate{}, err
	}

	// Update the channels it's associated with
	policyID, err := strconv.Atoi(cr.Spec.ForProvider.ID)
	if err != nil {
		return managed.ExternalUpdate{}, err
	}
	// ToDo: Convert to nerdgraph, once it's supported
	_, err = c.client.Alerts.UpdatePolicyChannelsWithContext(ctx, policyID, cr.Spec.ForProvider.ChannelIDs)
	if err != nil {
		return managed.ExternalUpdate{}, err
	}

	cr.SetConditions(xpv1.Available())
	return managed.ExternalUpdate{
		// Optionally return any details that may be required to connect to the
		// external resource. These will be stored as the connection secret.
		ConnectionDetails: managed.ConnectionDetails{},
	}, err
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.AlertsPolicy)
	if !ok {
		return errors.New(errNotPolicy)
	}

	cr.Status.SetConditions(xpv1.Deleting())
	if cr.Spec.ForProvider.ID == "" {
		fmt.Printf("skipping delete for policy %s: Id must be set", cr.Spec.ForProvider.Name)
		return nil
	}

	_, err := c.client.Alerts.DeletePolicyMutationWithContext(ctx, c.accountID, cr.Spec.ForProvider.ID)
	if err != nil {
		return err
	}

	return err
}

func (c *external) SetExternalNameIfNotSet(ctx context.Context, cr *v1alpha1.AlertsPolicy, response *alerts.AlertsPolicy) { // nolint:gocyclo
	// Set the ID, if not set
	ext := meta.GetExternalName(cr)
	if cr.Spec.ForProvider.ID == "" || ext == "" || ext != cr.Spec.ForProvider.Name {
		cr.Spec.ForProvider.ID = response.ID
		meta.SetExternalName(cr, cr.Spec.ForProvider.Name)
		_ = c.kube.Update(ctx, cr)
	}
}

// GetAlertsPolicyByIDOrName calls new relic api to get a policy by the ID.  If the ID doesn't exist it will fall back to get by name
func (c *external) GetAlertsPolicyByIDOrName(ctx context.Context, cr *v1alpha1.AlertsPolicy) (*alerts.AlertsPolicy, error) {
	defaultPolicy := &alerts.AlertsPolicy{}

	// Get the policy by ID
	policyByID, err := c.client.Alerts.QueryPolicyWithContext(ctx, c.accountID, cr.Spec.ForProvider.ID)

	// return the policy if found
	if err == nil {
		return policyByID, nil
	}

	if err != nil {
		// If not found, the ID may have changed - attempt to look up the new ID
		if strings.Contains(err.Error(), "Not Found") {
			policyID, err := c.GetAlertsPolicyIDByName(ctx, cr)
			// try to look up the policy using the new ID
			if policyID > 0 {
				// return the policy if found
				policyByID, err := c.client.Alerts.QueryPolicyWithContext(ctx, c.accountID, strconv.Itoa(policyID))
				if err == nil {
					cr.Spec.ForProvider.ID = policyByID.ID
					_ = c.kube.Update(ctx, cr)
					fmt.Println("Updating policy " + policyByID.Name + " with new ID: " + policyByID.ID)
					return policyByID, nil
				}
			}
			return defaultPolicy, err
		}
	}
	return defaultPolicy, nil
}

// GetAlertsPolicyIDByName uses the REST v2 API to return the ID for a given policy name
func (c *external) GetAlertsPolicyIDByName(ctx context.Context, cr *v1alpha1.AlertsPolicy) (int, error) {
	// Otherwise, look up the policy by name instead
	listParams := &alerts.ListPoliciesParams{Name: cr.Spec.ForProvider.Name}
	listPolicies, err := c.client.Alerts.ListPoliciesWithContext(ctx, listParams)
	if len(listPolicies) > 0 && err == nil {
		return listPolicies[0].ID, nil
	}
	return 0, err
}

// IsUpToDate determines whether the AlertsPolicy needs to be updated
func IsUpToDate(p *v1alpha1.AlertsPolicy, cd alerts.AlertsPolicy) bool {
	if !cmp.Equal(p.Spec.ForProvider.Name, cd.Name, cmpopts.EquateEmpty()) {
		return false
	}
	return cmp.Equal(p.Spec.ForProvider.IncidentPreference, string(cd.IncidentPreference), cmpopts.EquateEmpty())
}
