/*
Copyright 2022 The Crossplane Authors.

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

package nrqlalertcondition

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/newrelic/newrelic-client-go/v2/newrelic"
	"github.com/newrelic/newrelic-client-go/v2/pkg/alerts"
	"github.com/pkg/errors"
	"go.openly.dev/pointy"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-newrelic/apis/nrqlalertcondition/v1alpha1"
	apisv1alpha1 "github.com/crossplane-contrib/provider-newrelic/apis/v1alpha1"
	nr "github.com/crossplane-contrib/provider-newrelic/pkg/clients"
	"github.com/crossplane-contrib/provider-newrelic/pkg/features"
)

const (
	errNotChannel   = "managed resource is not a Channel custom resource"
	errTrackPCUsage = "cannot track ProviderConfig usage"
	errGetPC        = "cannot get ProviderConfig"
	errGetAccountID = "cannot get accountID from ProviderConfig"
)

// Setup adds a controller that reconciles NrqlAlertCondition.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.NrqlAlertConditionGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), apisv1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithExternalConnecter(&connector{
			kube:  mgr.GetClient(),
			usage: resource.NewProviderConfigUsageTracker(mgr.GetClient(), &apisv1alpha1.ProviderConfigUsage{}),
		}),
		managed.WithConnectionPublishers(),
		managed.WithInitializers(),
		managed.WithPollInterval(o.PollInterval),
		managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.NrqlAlertConditionGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&v1alpha1.NrqlAlertCondition{}).
		Complete(r)
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
	cr, ok := mg.(*v1alpha1.NrqlAlertCondition)
	if !ok {
		return nil, errors.New(errNotChannel)
	}

	if err := c.usage.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, errTrackPCUsage)
	}

	pc := &apisv1alpha1.ProviderConfig{}
	if err := c.kube.Get(ctx, types.NamespacedName{Name: cr.GetProviderConfigReference().Name}, pc); err != nil {
		return nil, errors.Wrap(err, errGetPC)
	}

	// Get the account id from the provider config
	accountID, err := nr.ExtractNewRelicAccountID(pc)
	if err != nil {
		return nil, err
	}

	// Create a client using NR Credentials
	nrClient, err := nr.ExtractNewRelicCredentials(ctx, c.kube, pc)
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
	cr, ok := mg.(*v1alpha1.NrqlAlertCondition)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotChannel)
	}

	condition, err := c.GetNrqlConditionByIDOrName(ctx, cr)
	// nerdgraph errors aren't great
	if err != nil {
		if strings.Contains(err.Error(), "Not Found") {
			return managed.ExternalObservation{
				ResourceExists: false,
			}, nil
		}
		return managed.ExternalObservation{}, err
	}
	// Nerdgraph may return an empty object
	if condition == nil || condition.ID == "" {
		return managed.ExternalObservation{ResourceExists: false}, err
	}
	// Set the ID, if not set
	c.SetExternalNameIfNotSet(ctx, cr, condition)

	// Update the status
	cr.Status.SetConditions(xpv1.Available())
	cr.Status.AtProvider = v1alpha1.NrqlAlertConditionObservation{
		ID: condition.ID,
	}

	// Resource was found
	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  IsUpToDate(cr, condition),
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.NrqlAlertCondition)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotChannel)
	}

	cr.Status.SetConditions(xpv1.Creating())
	// Resolve References to get the policy id
	err := cr.ResolveReferences(ctx, c.kube)
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	_ = c.kube.Update(ctx, cr)

	// Create the condition
	input := GenerateAlertConditionInput(cr)
	response, err := CreateNrqlCondition(ctx, c.client, c.accountID, cr.Spec.ForProvider.AlertsPolicyID, input)

	if err != nil {
		// If the policy is not found, re-run the referencer
		matched, _ := regexp.MatchString(`Policy with ID \d* not found`, err.Error())
		if matched {
			cr.Spec.ForProvider.AlertsPolicyID = ""
			_ = c.kube.Update(ctx, cr)
		}
		return managed.ExternalCreation{}, err
	}

	// Set the ID, if not set
	meta.SetExternalName(cr, cr.Spec.ForProvider.Name)
	c.SetExternalNameIfNotSet(ctx, cr, response)
	cr.Status.SetConditions(xpv1.Available())
	return managed.ExternalCreation{}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.NrqlAlertCondition)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotChannel)
	}

	input := GenerateAlertConditionInput(cr)
	update, uErr := GenerateNrqlConditionUpdateInput(input)
	if uErr != nil {
		return managed.ExternalUpdate{
			ConnectionDetails: managed.ConnectionDetails{},
		}, uErr
	}
	_, err := UpdateNrqlConditionStaticMutationWithContext(ctx, c.client, c.accountID, cr.Spec.ForProvider.ID, update)

	return managed.ExternalUpdate{
		ConnectionDetails: managed.ConnectionDetails{},
	}, err
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.NrqlAlertCondition)
	if !ok {
		return errors.New(errNotChannel)
	}

	cr.Status.SetConditions(xpv1.Deleting())
	if cr.Spec.ForProvider.ID == "" {
		fmt.Printf("skipping delete for nrql condition %s: Id must be set", cr.Spec.ForProvider.Name)
		return nil
	}

	result, err := c.client.Alerts.DeleteNrqlConditionMutationWithContext(ctx, c.accountID, cr.Spec.ForProvider.ID)
	if err != nil {
		fmt.Printf("Unable to delete condition %s: %s", cr.Spec.ForProvider.ID, err)
		if strings.Contains(err.Error(), "Not Found") {
			return nil
		}
	}
	fmt.Printf("Deleted conditionID %s with result of %s", cr.Spec.ForProvider.ID, result)

	return err
}

func (c *external) SetExternalNameIfNotSet(ctx context.Context, cr *v1alpha1.NrqlAlertCondition, response *alerts.NrqlAlertCondition) {
	// Set the ID, if not set
	ext := meta.GetExternalName(cr)
	if cr.Spec.ForProvider.ID == "" || ext == "" || ext != cr.Spec.ForProvider.Name {
		cr.Spec.ForProvider.ID = response.ID
		meta.SetExternalName(cr, cr.Spec.ForProvider.Name)
		_ = c.kube.Update(ctx, cr)
	}
}

// GetNrqlConditionByIDOrName gets a condition by ID and if it doesn't exist it attempts the name
func (c *external) GetNrqlConditionByIDOrName(ctx context.Context, cr *v1alpha1.NrqlAlertCondition) (*alerts.NrqlAlertCondition, error) {
	defaultCondition := &alerts.NrqlAlertCondition{}

	// Get the condition by ID
	conditionByID, err := c.client.Alerts.GetNrqlConditionQueryWithContext(ctx, c.accountID, cr.Spec.ForProvider.ID)
	if err == nil {
		return conditionByID, nil
	}

	if err != nil {
		// If not found, the ID may have changed - attempt to look up the new ID
		if strings.Contains(err.Error(), "Not Found") {
			conditionByName, err := c.GetNrqlConditionByName(ctx, cr)
			if err == nil {
				cr.Spec.ForProvider.ID = conditionByName.ID
				_ = c.kube.Update(ctx, cr)
				fmt.Println("Updating " + cr.Spec.ForProvider.Name + " with new ID: " + conditionByName.ID)
				return conditionByName, nil
			}
			return defaultCondition, err
		}
	}
	return defaultCondition, nil
}

// GetNrqlConditionByName uses the search function to get a condition by name
func (c *external) GetNrqlConditionByName(ctx context.Context, cr *v1alpha1.NrqlAlertCondition) (*alerts.NrqlAlertCondition, error) {
	criteria := alerts.NrqlConditionsSearchCriteria{
		Name:     cr.Spec.ForProvider.Name,
		PolicyID: cr.Spec.ForProvider.AlertsPolicyID,
	}
	// Search for conditions and always use the first one
	conditions, err := c.client.Alerts.SearchNrqlConditionsQueryWithContext(ctx, c.accountID, criteria)
	if len(conditions) > 0 && err == nil {
		return conditions[0], nil
	}
	// Otherwise
	return &alerts.NrqlAlertCondition{}, err
}

// CreateNrqlCondition calls the right API based on the condition type
func CreateNrqlCondition(ctx context.Context, client *newrelic.NewRelic, accountID int, policyID string, input alerts.NrqlConditionCreateInput) (*alerts.NrqlAlertCondition, error) {
	conditionType := input.Type
	input.Type = ""
	// "Argument \"condition\" has invalid value $condition.\nIn field \"type\": Unknown field."}
	if conditionType == "BASELINE" {
		return client.Alerts.CreateNrqlConditionBaselineMutationWithContext(ctx, accountID, policyID, input)
	}
	// Default is static
	return client.Alerts.CreateNrqlConditionStaticMutationWithContext(ctx, accountID, policyID, input)
}

// UpdateNrqlConditionStaticMutationWithContext calls the right API based on the condition type
func UpdateNrqlConditionStaticMutationWithContext(ctx context.Context, client *newrelic.NewRelic, accountID int, conditionID string, input alerts.NrqlConditionUpdateInput) (*alerts.NrqlAlertCondition, error) {
	conditionType := input.Type
	input.Type = ""
	// "Argument \"condition\" has invalid value $condition.\nIn field \"type\": Unknown field."}
	if conditionType == "BASELINE" {
		return client.Alerts.UpdateNrqlConditionBaselineMutationWithContext(ctx, accountID, conditionID, input)
	}
	// Default is static
	return client.Alerts.UpdateNrqlConditionStaticMutationWithContext(ctx, accountID, conditionID, input)
}

// GenerateAlertConditionInput generates an input object
func GenerateAlertConditionInput(cr *v1alpha1.NrqlAlertCondition) alerts.NrqlConditionCreateInput {

	input := alerts.NrqlConditionCreateInput{
		NrqlConditionCreateBase: alerts.NrqlConditionCreateBase{
			Enabled: cr.Spec.ForProvider.Enabled,
			Name:    cr.Spec.ForProvider.Name,
		},
	}
	input.Type = alerts.NrqlConditionType(cr.Spec.ForProvider.Type)
	if cr.Spec.ForProvider.RunbookURL != nil {
		input.RunbookURL = pointy.StringValue(cr.Spec.ForProvider.RunbookURL, "")
	}
	if cr.Spec.ForProvider.Description != nil {
		input.NrqlConditionCreateBase.Description = pointy.StringValue(cr.Spec.ForProvider.Description, "")
	}

	if cr.Spec.ForProvider.ViolationTimeLimitSeconds != nil {
		input.ViolationTimeLimitSeconds = *cr.Spec.ForProvider.ViolationTimeLimitSeconds
	}

	// Expiration
	expiration := GenerateNrqlConditionExpiration(cr)
	input.Expiration = &expiration

	// Nrql
	input.Nrql = alerts.NrqlConditionCreateQuery{Query: cr.Spec.ForProvider.Nrql.Query}

	// Signal
	signal := GenerateNrqlConditionSignal(cr)
	input.Signal = &signal

	// Terms
	terms := GenerateNrqlConditionTerm(cr)
	input.Terms = terms

	// Only valid for "BASELINE" type
	if cr.Spec.ForProvider.BaselineDirection != nil {
		baselineDirection := alerts.NrqlBaselineDirection(*cr.Spec.ForProvider.BaselineDirection)
		input.BaselineDirection = &baselineDirection
	}

	return input
}

// GenerateNrqlConditionTerm generates an input object
func GenerateNrqlConditionTerm(cr *v1alpha1.NrqlAlertCondition) []alerts.NrqlConditionTerm {

	// Handle nil pointer refs
	terms := make([]alerts.NrqlConditionTerm, 0)
	if cr.Spec.ForProvider.Terms != nil {
		for _, v := range cr.Spec.ForProvider.Terms {
			threshold, _ := strconv.ParseFloat(v.Threshold, 64)
			t := alerts.NrqlConditionTerm{
				Operator:             alerts.AlertsNRQLConditionTermsOperator(v.Operator),
				Priority:             alerts.NrqlConditionPriority(v.Priority),
				Threshold:            &threshold,
				ThresholdDuration:    v.ThresholdDuration,
				ThresholdOccurrences: alerts.ThresholdOccurrence(v.ThresholdOccurrences),
			}
			terms = append(terms, t)
		}
	}
	return terms
}

// GenerateNrqlConditionSignal generates an input object
func GenerateNrqlConditionSignal(cr *v1alpha1.NrqlAlertCondition) alerts.AlertsNrqlConditionCreateSignal {

	signal := alerts.AlertsNrqlConditionCreateSignal{}

	if cr.Spec.ForProvider.Signal.AggregationWindow != nil {
		signal.AggregationWindow = cr.Spec.ForProvider.Signal.AggregationWindow
	}

	if cr.Spec.ForProvider.Signal.EvaluationDelay != nil {
		signal.EvaluationDelay = cr.Spec.ForProvider.Signal.EvaluationDelay
	}

	if cr.Spec.ForProvider.Signal.EvaluationOffset != nil {
		signal.EvaluationOffset = cr.Spec.ForProvider.Signal.EvaluationOffset
	}
	fillOption := alerts.AlertsFillOption(cr.Spec.ForProvider.Signal.FillOption)
	signal.FillOption = &fillOption
	if cr.Spec.ForProvider.Signal.FillValue != nil {
		// "NONE" can't have a FillValue
		if fillOption == alerts.AlertsFillOptionTypes.STATIC {
			if cr.Spec.ForProvider.Signal.FillValue != nil {
				fillValue, _ := strconv.ParseFloat(*cr.Spec.ForProvider.Signal.FillValue, 64)
				signal.FillValue = &fillValue
			}
		}
	}

	if cr.Spec.ForProvider.Signal.AggregationMethod != nil {
		aggregationMethod := alerts.NrqlConditionAggregationMethod(*cr.Spec.ForProvider.Signal.AggregationMethod)
		signal.AggregationMethod = &aggregationMethod
	}
	if cr.Spec.ForProvider.Signal.AggregationDelay != nil {
		signal.AggregationDelay = cr.Spec.ForProvider.Signal.AggregationDelay
	}
	if cr.Spec.ForProvider.Signal.AggregationTimer != nil {
		signal.AggregationTimer = cr.Spec.ForProvider.Signal.AggregationTimer
	}

	return signal
}

// GenerateNrqlConditionExpiration generates an input object
func GenerateNrqlConditionExpiration(cr *v1alpha1.NrqlAlertCondition) alerts.AlertsNrqlConditionExpiration {

	if cr.Spec.ForProvider.Expiration == nil {
		return alerts.AlertsNrqlConditionExpiration{}
	}

	expiration := alerts.AlertsNrqlConditionExpiration{
		CloseViolationsOnExpiration: cr.Spec.ForProvider.Expiration.CloseViolationsOnExpiration,
		OpenViolationOnExpiration:   cr.Spec.ForProvider.Expiration.OpenViolationOnExpiration,
	}

	if cr.Spec.ForProvider.Expiration.ExpirationDuration != nil && *cr.Spec.ForProvider.Expiration.ExpirationDuration > 0 {
		expiration.ExpirationDuration = cr.Spec.ForProvider.Expiration.ExpirationDuration
	}

	return expiration
}

// GenerateNrqlConditionUpdateInput generates an input object
func GenerateNrqlConditionUpdateInput(input alerts.NrqlConditionCreateInput) (alerts.NrqlConditionUpdateInput, error) {
	// As part of this change the object required for a nrql condition to be created and updated
	// are now different.  Instead of constructing 2 different types, we will convert a
	// `create` object into an `update` object.  It is easier than having to maintain more code
	// https://github.com/newrelic/newrelic-client-go/pull/805

	var update alerts.NrqlConditionUpdateInput
	out, errMarshal := json.Marshal(input)
	if errMarshal != nil {
		return update, errMarshal
	}

	err := json.Unmarshal(out, &update)
	if err != nil {
		return update, err
	}

	return update, nil
}

// IsUpToDate performs comparison
func IsUpToDate(p *v1alpha1.NrqlAlertCondition, cd *alerts.NrqlAlertCondition) bool {

	input := GenerateAlertConditionInput(p)

	if !cmp.Equal(input.Name, cd.Name, cmpopts.EquateEmpty()) {
		return false
	}
	if !cmp.Equal(input.Type, cd.Type, cmpopts.EquateEmpty()) {
		return false
	}

	if !cmp.Equal(input.RunbookURL, cd.RunbookURL, cmpopts.EquateEmpty()) {
		return false
	}

	if !cmp.Equal(input.Enabled, cd.Enabled, cmpopts.EquateEmpty()) {
		return false
	}

	if !cmp.Equal(input.ViolationTimeLimitSeconds, cd.ViolationTimeLimitSeconds, cmpopts.EquateEmpty()) {
		return false
	}

	// Compare whether the Terms are equal with custom function, ignore ordering
	if !termsAreEqual(input.Terms, cd.Terms) {
		return false
	}

	if !cmp.Equal(input.Nrql.Query, cd.Nrql.Query, cmpopts.EquateEmpty()) {
		return false
	}

	// Compare whether the Signals are equal with custom function
	if !signalsAreEqual(*input.Signal, *cd.Signal) {
		return false
	}

	if !expirationsAreEqual(input.Expiration, cd.Expiration) {
		return false
	}
	return true
}

// expirationsAreEqual compares Expiration
func expirationsAreEqual(expiration *alerts.AlertsNrqlConditionExpiration, nrExpiration *alerts.AlertsNrqlConditionExpiration) bool { //nolint:gocyclo

	if (expiration == nil && nrExpiration != nil) || (expiration != nil && nrExpiration == nil) {
		return false
	}

	if (expiration.ExpirationDuration == nil && nrExpiration.ExpirationDuration != nil) || (expiration.ExpirationDuration != nil && nrExpiration.ExpirationDuration == nil) {
		return false
	}

	if expiration.ExpirationDuration != nil && nrExpiration.ExpirationDuration != nil {
		if !cmp.Equal(pointy.IntValue(expiration.ExpirationDuration, 0), pointy.IntValue(nrExpiration.ExpirationDuration, 0), cmpopts.EquateEmpty()) {
			return false
		}
	}
	if !cmp.Equal(expiration.OpenViolationOnExpiration, nrExpiration.OpenViolationOnExpiration, cmpopts.EquateEmpty()) {
		return false
	}
	if !cmp.Equal(expiration.CloseViolationsOnExpiration, nrExpiration.CloseViolationsOnExpiration, cmpopts.EquateEmpty()) {
		return false
	}
	return true
}

// signalsAreEqual compares signals
func signalsAreEqual(signal alerts.AlertsNrqlConditionCreateSignal, nrSignal alerts.AlertsNrqlConditionSignal) bool { //nolint:gocyclo

	if nrSignal.FillOption == nil {
		nrSignal.FillOption = &alerts.AlertsFillOptionTypes.NONE
	}
	if signal.FillOption == nil {
		signal.FillOption = &alerts.AlertsFillOptionTypes.NONE
	}
	if !cmp.Equal(*signal.FillOption, *nrSignal.FillOption, cmpopts.EquateEmpty()) {
		return false
	}

	if signal.FillValue != nil {
		// "NONE" can't have a FillValue
		if *signal.FillOption == alerts.AlertsFillOptionTypes.STATIC {
			if !cmp.Equal(*signal.FillValue, *nrSignal.FillValue, cmpopts.EquateEmpty()) {
				return false
			}
		}
	}

	if signal.AggregationMethod == nil && nrSignal.AggregationMethod != nil {
		return false
	}
	if signal.AggregationMethod != nil && nrSignal.AggregationMethod == nil {
		return false
	}
	if signal.AggregationMethod != nil && nrSignal.AggregationMethod != nil {
		crAggregationMethod := *signal.AggregationMethod
		nrAggregationMethod := *nrSignal.AggregationMethod
		if !cmp.Equal(crAggregationMethod, nrAggregationMethod, cmpopts.EquateEmpty()) {
			return false
		}
	}

	if !cmp.Equal(pointy.IntValue(signal.EvaluationDelay, 0), pointy.IntValue(nrSignal.EvaluationDelay, 0), cmpopts.EquateEmpty()) {
		return false
	}

	if !cmp.Equal(pointy.IntValue(signal.EvaluationOffset, 0), pointy.IntValue(nrSignal.EvaluationOffset, 0), cmpopts.EquateEmpty()) {
		return false
	}

	if !cmp.Equal(pointy.IntValue(signal.AggregationDelay, 0), pointy.IntValue(nrSignal.AggregationDelay, 0), cmpopts.EquateEmpty()) {
		return false
	}

	if !cmp.Equal(pointy.IntValue(signal.AggregationTimer, 0), pointy.IntValue(nrSignal.AggregationTimer, 0), cmpopts.EquateEmpty()) {
		return false
	}

	return true
}

// termsAreEqual compares terms
func termsAreEqual(terms []alerts.NrqlConditionTerm, nrTerms []alerts.NrqlConditionTerm) bool {
	stringTerms := make([]string, 0)
	stringNrTerms := make([]string, 0)
	for _, i := range terms {
		stringTerm := fmt.Sprintf("%s_%d_%s_%s_%f", i.Priority, i.ThresholdDuration, i.ThresholdOccurrences, i.Operator, *i.Threshold)
		stringTerms = append(stringTerms, stringTerm)
	}
	for _, i := range nrTerms {
		stringNrTerm := fmt.Sprintf("%s_%d_%s_%s_%f", i.Priority, i.ThresholdDuration, i.ThresholdOccurrences, i.Operator, *i.Threshold)
		stringNrTerms = append(stringNrTerms, stringNrTerm)
	}
	// Ignore sort order
	sortCmp := cmpopts.SortSlices(func(i, j string) bool {
		return i < j
	})
	return cmp.Equal(stringTerms, stringNrTerms, sortCmp, cmpopts.EquateEmpty())
}
