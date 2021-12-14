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

package dashboard

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/newrelic/newrelic-client-go/newrelic"
	"github.com/newrelic/newrelic-client-go/pkg/common"
	"github.com/newrelic/newrelic-client-go/pkg/dashboards"
	"github.com/newrelic/newrelic-client-go/pkg/entities"
	"github.com/newrelic/newrelic-client-go/pkg/nrdb"
	"github.com/openlyinc/pointy"
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

	"github.com/crossplane-contrib/provider-newrelic/apis/dashboard/v1alpha1"
	apisv1alpha1 "github.com/crossplane-contrib/provider-newrelic/apis/v1alpha1"
	nr "github.com/crossplane-contrib/provider-newrelic/internal/clients"
)

const (
	errNotDashboard = "managed resource is not a Dashboard custom resource"
	errTrackPCUsage = "cannot track ProviderConfig usage"
	errGetPC        = "cannot get ProviderConfig"
	errGetCreds     = "cannot get credentials"
	errGetAccountID = "cannot get accountId from ProviderConfig"
)

// Setup adds a controller that reconciles a managed resources.
func Setup(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter) error {
	name := managed.ControllerName(v1alpha1.DashboardGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewController(rl),
		}).
		For(&v1alpha1.Dashboard{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.DashboardGroupVersionKind),
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
	cr, ok := mg.(*v1alpha1.Dashboard)
	if !ok {
		return nil, errors.New(errNotDashboard)
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

func (c *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) { // nolint:gocyclo
	cr, ok := mg.(*v1alpha1.Dashboard)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotDashboard)
	}

	externalName := meta.GetExternalName(cr)
	if externalName == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	if cr.Spec.ForProvider.GUID == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	// Get the dashboard by GUID
	entityGUID := common.EntityGUID(cr.Spec.ForProvider.GUID)
	dashboard, err := c.client.Dashboards.GetDashboardEntityWithContext(ctx, entityGUID)

	if err != nil {
		if strings.Contains(err.Error(), "entity not found") {
			return managed.ExternalObservation{
				ResourceExists: false,
			}, nil
		}
		return managed.ExternalObservation{}, err
	}

	if dashboard.GUID == "" {
		return managed.ExternalObservation{ResourceExists: false}, err
	}

	// We have to use the Update Result, not a re-read of the entity as the changes take
	// some amount of time to be re-indexed

	// Update the status
	cr.Status.SetConditions(xpv1.Available())
	cr.Status.AtProvider = v1alpha1.DashboardObservation{
		GUID: string(dashboard.GUID),
	}

	cr.SetConditions(xpv1.Available())
	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  IsUpToDate(cr, *dashboard),
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Dashboard)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotDashboard)
	}
	cr.SetConditions(xpv1.Creating())

	// Create the dashboard
	input := GenerateDashboardInput(cr)

	response, err := c.client.Dashboards.DashboardCreateWithContext(ctx, c.accountID, input)
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	if len(response.Errors) > 0 {
		return managed.ExternalCreation{}, errors.New(response.Errors[0].Description)
	}

	// Set the ID for all pages and widgets
	UpdateGUIDS(ctx, c, cr, response.EntityResult)

	// Set the GUID
	cr.SetConditions(xpv1.Available())

	return managed.ExternalCreation{}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.Dashboard)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotDashboard)
	}

	// Create the dashboard input
	input := GenerateDashboardInput(cr)
	entityGUID := common.EntityGUID(cr.Spec.ForProvider.GUID)

	// See - https://github.com/newrelic/newrelic-client-go/issues/802
	// Updating is causing duplicates right now.
	response, err := c.client.Dashboards.DashboardUpdateWithContext(ctx, input, entityGUID)
	if err != nil {
		return managed.ExternalUpdate{}, err
	}

	// Set the ID for all pages and widgets
	UpdateGUIDS(ctx, c, cr, response.EntityResult)

	cr.SetConditions(xpv1.Available())
	return managed.ExternalUpdate{
		// Optionally return any details that may be required to connect to the
		// external resource. These will be stored as the connection secret.
		ConnectionDetails: managed.ConnectionDetails{},
	}, err
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.Dashboard)
	if !ok {
		return errors.New(errNotDashboard)
	}

	cr.Status.SetConditions(xpv1.Deleting())
	if cr.Spec.ForProvider.GUID == "" {
		fmt.Printf("skipping delete for dashboard %s: guid must be set", cr.Spec.ForProvider.Name)
		return nil
	}

	entityGUID := common.EntityGUID(cr.Spec.ForProvider.GUID)
	_, err := c.client.Dashboards.DashboardDeleteWithContext(ctx, entityGUID)
	if err != nil {
		return err
	}

	return err
}

// IsUpToDate determines whether the Dashboard needs to be updated
func IsUpToDate(p *v1alpha1.Dashboard, cd entities.DashboardEntity) bool {
	// Convert both objects to the same type
	crObject := GenerateDashboardInput(p)
	nrObject := GenerateDashboardInputFromEntity(cd)

	return cmp.Equal(crObject,
		nrObject,
		cmpopts.EquateEmpty(),
		cmpopts.IgnoreTypes(entities.DashboardWidgetRawConfiguration{}, &entities.DashboardWidgetRawConfiguration{}),
	)
}

// UpdateGUIDS updates all the GUIDs and IDs in our object to match what was created by the API
func UpdateGUIDS(ctx context.Context, c *external, cr *v1alpha1.Dashboard, dashboard dashboards.DashboardEntityResult) { // nolint:gocyclo

	needsKubernetesUpdate := false

	// Set external name if not set
	if meta.GetExternalName(cr) != string(dashboard.GUID) {
		meta.SetExternalName(cr, string(dashboard.GUID))
		needsKubernetesUpdate = true
	}

	// Set the ID for all pages and widgets
	// Without the GUID and ID updates won't work properly and will generate duplicates
	if cr.Spec.ForProvider.GUID != string(dashboard.GUID) {
		cr.Spec.ForProvider.GUID = string(dashboard.GUID)
		needsKubernetesUpdate = true
	}
	for p, crPage := range cr.Spec.ForProvider.Pages {
		for _, page := range dashboard.Pages {
			// Gross but use the page Name as unique identifier
			if crPage.Name == page.Name {
				if crPage.GUID != string(page.GUID) {
					cr.Spec.ForProvider.Pages[p].GUID = string(page.GUID)
					needsKubernetesUpdate = true
				}
				for w, crWidget := range crPage.Widgets {
					for _, widget := range page.Widgets {
						// Try to match identifiers
						if crWidget.Title == widget.Title {
							if crWidget.Layout.Row == widget.Layout.Row && crWidget.Layout.Width == widget.Layout.Width && crWidget.Layout.Column == widget.Layout.Column && crWidget.Layout.Height == widget.Layout.Height {
								if pointy.StringValue(crWidget.ID, "") != widget.ID {
									cr.Spec.ForProvider.Pages[p].Widgets[w].ID = pointy.String(widget.ID)
									needsKubernetesUpdate = true
								}
							}
						}
					}
				}
			}
		}
	}
	if needsKubernetesUpdate {
		_ = c.kube.Update(ctx, cr)
	}
}

// GenerateDashboardInputFromEntity generates an input object
// from new relic output
func GenerateDashboardInputFromEntity(cd entities.DashboardEntity) dashboards.DashboardInput {
	input := dashboards.DashboardInput{Name: cd.Name,
		Description: cd.Description,
		Permissions: cd.Permissions}
	input.Pages = GenerateDashboardPageInputFromEntity(cd.Pages)

	return input
}

// GenerateDashboardPageInputFromEntity generates an input object
func GenerateDashboardPageInputFromEntity(cd []entities.DashboardPage) []dashboards.DashboardPageInput {
	input := make([]dashboards.DashboardPageInput, 0)
	for _, page := range cd {
		pageInput := dashboards.DashboardPageInput{Name: page.Name,
			Description: page.Description,
			GUID:        page.GUID,
		}
		pageInput.Widgets = GenerateDashboardWidgetInputFromEntity(page.Widgets)

		input = append(input, pageInput)
	}

	sort.Slice(input, func(i, j int) bool {
		// Sort first by GUID
		if a, b := input[i].GUID, input[j].GUID; a != b {
			return a < b
		}
		// Then by name
		return input[i].Name < input[j].Name
	})

	return input
}

// GenerateDashboardWidgetInputFromEntity generates an input object
func GenerateDashboardWidgetInputFromEntity(cd []entities.DashboardWidget) []dashboards.DashboardWidgetInput {
	input := make([]dashboards.DashboardWidgetInput, 0)

	for _, widget := range cd {
		widgetInput := dashboards.DashboardWidgetInput{Title: widget.Title, ID: widget.ID}
		widgetInput.Configuration = GenerateDashboardWidgetConfigurationInputFromEntity(widget.Configuration)
		widgetInput.Layout = GenerateDashboardWidgetLayoutInputFromEntity(widget.Layout)
		widgetInput.Visualization = GenerateDashboardWidgetVisualizationInputFromEntity(widget.Visualization)
		// Some types of visualizations use raw configuration
		if widget.RawConfiguration != nil {
			widgetInput.RawConfiguration = widget.RawConfiguration
		}
		input = append(input, widgetInput)
	}
	sort.Slice(input, func(i, j int) bool {
		// Sort first by ID
		if a, b := input[i].ID, input[j].ID; a != b {
			return a < b
		}
		// If there are items with the same ID use the title
		if a, b := input[i].Title, input[j].Title; a != b {
			return a < b
		}
		// Lastly, try sorting by layout as it would be an issue if 2 widgets had the same exact layout
		return strings.Join([]string{strconv.Itoa(input[i].Layout.Row), strconv.Itoa(input[i].Layout.Height), strconv.Itoa(input[i].Layout.Width), strconv.Itoa(input[i].Layout.Column)}, "") < strings.Join([]string{strconv.Itoa(input[j].Layout.Row), strconv.Itoa(input[j].Layout.Height), strconv.Itoa(input[j].Layout.Width), strconv.Itoa(input[j].Layout.Column)}, "")
	})

	return input
}

// GenerateDashboardWidgetConfigurationInputFromEntity generates an input object
func GenerateDashboardWidgetConfigurationInputFromEntity(cd entities.DashboardWidgetConfiguration) dashboards.DashboardWidgetConfigurationInput { // nolint:gocyclo
	input := dashboards.DashboardWidgetConfigurationInput{}

	if cd.Area.NRQLQueries != nil {
		nrqlQueries := make([]dashboards.DashboardWidgetNRQLQueryInput, 0)
		for _, q := range cd.Area.NRQLQueries {
			item := dashboards.DashboardWidgetNRQLQueryInput{AccountID: q.AccountID, Query: q.Query}
			nrqlQueries = append(nrqlQueries, item)
		}
		input.Area = &dashboards.DashboardAreaWidgetConfigurationInput{NRQLQueries: nrqlQueries}
	}

	if cd.Bar.NRQLQueries != nil {
		nrqlQueries := make([]dashboards.DashboardWidgetNRQLQueryInput, 0)
		for _, q := range cd.Bar.NRQLQueries {
			item := dashboards.DashboardWidgetNRQLQueryInput{AccountID: q.AccountID, Query: q.Query}
			nrqlQueries = append(nrqlQueries, item)
		}
		input.Bar = &dashboards.DashboardBarWidgetConfigurationInput{NRQLQueries: nrqlQueries}
	}

	if cd.Billboard.NRQLQueries != nil {
		nrqlQueries := make([]dashboards.DashboardWidgetNRQLQueryInput, 0)
		for _, q := range cd.Billboard.NRQLQueries {
			item := dashboards.DashboardWidgetNRQLQueryInput{AccountID: q.AccountID, Query: q.Query}
			nrqlQueries = append(nrqlQueries, item)
		}

		thresholds := make([]dashboards.DashboardBillboardWidgetThresholdInput, 0)
		alertSeverities := make([]string, 0, len(cd.Billboard.Thresholds))
		for _, threshold := range cd.Billboard.Thresholds {
			alertSeverities = append(alertSeverities, string(threshold.AlertSeverity))
		}
		sort.Strings(alertSeverities)
		for _, alertSeverity := range alertSeverities {
			for _, q := range cd.Billboard.Thresholds {
				if alertSeverity == string(q.AlertSeverity) {
					item := dashboards.DashboardBillboardWidgetThresholdInput{AlertSeverity: q.AlertSeverity, Value: pointy.Float64(q.Value)}
					thresholds = append(thresholds, item)
				}
			}
		}

		input.Billboard = &dashboards.DashboardBillboardWidgetConfigurationInput{NRQLQueries: nrqlQueries, Thresholds: thresholds}
	}

	if cd.Line.NRQLQueries != nil {
		nrqlQueries := make([]dashboards.DashboardWidgetNRQLQueryInput, 0)
		for _, q := range cd.Line.NRQLQueries {
			item := dashboards.DashboardWidgetNRQLQueryInput{AccountID: q.AccountID, Query: q.Query}
			nrqlQueries = append(nrqlQueries, item)
		}
		input.Line = &dashboards.DashboardLineWidgetConfigurationInput{NRQLQueries: nrqlQueries}
	}

	input.Markdown = &dashboards.DashboardMarkdownWidgetConfigurationInput{
		Text: cd.Markdown.Text,
	}

	if cd.Pie.NRQLQueries != nil {
		nrqlQueries := make([]dashboards.DashboardWidgetNRQLQueryInput, 0)
		for _, q := range cd.Pie.NRQLQueries {
			item := dashboards.DashboardWidgetNRQLQueryInput{AccountID: q.AccountID, Query: q.Query}
			nrqlQueries = append(nrqlQueries, item)
		}
		input.Pie = &dashboards.DashboardPieWidgetConfigurationInput{NRQLQueries: nrqlQueries}
	}

	if cd.Table.NRQLQueries != nil {
		nrqlQueries := make([]dashboards.DashboardWidgetNRQLQueryInput, 0)
		for _, q := range cd.Table.NRQLQueries {
			item := dashboards.DashboardWidgetNRQLQueryInput{AccountID: q.AccountID, Query: q.Query}
			nrqlQueries = append(nrqlQueries, item)
		}
		input.Table = &dashboards.DashboardTableWidgetConfigurationInput{NRQLQueries: nrqlQueries}
	}
	return input
}

// GenerateDashboardWidgetLayoutInputFromEntity generates an input object
func GenerateDashboardWidgetLayoutInputFromEntity(cd entities.DashboardWidgetLayout) dashboards.DashboardWidgetLayoutInput {
	input := dashboards.DashboardWidgetLayoutInput{
		Column: cd.Column,
		Row:    cd.Row,
		Height: cd.Height,
		Width:  cd.Width,
	}
	return input
}

// GenerateDashboardWidgetVisualizationInputFromEntity generates an input object
func GenerateDashboardWidgetVisualizationInputFromEntity(cd entities.DashboardWidgetVisualization) dashboards.DashboardWidgetVisualizationInput {
	input := dashboards.DashboardWidgetVisualizationInput{ID: cd.ID}
	return input
}

// GenerateDashboardInput generates an input object
// from our managed object
func GenerateDashboardInput(cr *v1alpha1.Dashboard) dashboards.DashboardInput {

	input := dashboards.DashboardInput{Name: cr.Spec.ForProvider.Name}

	if cr.Spec.ForProvider.Description != nil {
		input.Description = pointy.StringValue(cr.Spec.ForProvider.Description, "")
	}

	inputPermissions := pointy.StringValue(cr.Spec.ForProvider.Permissions, "PUBLIC_READ_WRITE")
	permissions := entities.DashboardPermissions(inputPermissions)
	input.Permissions = permissions

	input.Pages = GenerateDashboardPageInput(cr)

	return input
}

// GenerateDashboardPageInput generates an input object
func GenerateDashboardPageInput(cr *v1alpha1.Dashboard) []dashboards.DashboardPageInput {
	input := make([]dashboards.DashboardPageInput, 0)
	for _, page := range cr.Spec.ForProvider.Pages {
		pageInput := dashboards.DashboardPageInput{Name: page.Name}
		pageInput.GUID = common.EntityGUID(page.GUID)

		if page.Description != nil {
			pageInput.Description = pointy.StringValue(page.Description, "")
		}

		pageInput.Widgets = GenerateDashboardWidgetInput(page)

		input = append(input, pageInput)
	}

	sort.Slice(input, func(i, j int) bool {
		// Sort first by GUID
		if a, b := input[i].GUID, input[j].GUID; a != b {
			return a < b
		}
		// Then by name
		return input[i].Name < input[j].Name
	})

	return input
}

// GenerateDashboardWidgetInput generates an input object
func GenerateDashboardWidgetInput(cr v1alpha1.DashboardPage) []dashboards.DashboardWidgetInput { // nolint:gocyclo
	input := make([]dashboards.DashboardWidgetInput, 0)

	for _, widget := range cr.Widgets {
		widgetInput := dashboards.DashboardWidgetInput{Title: widget.Title}
		// Set the ID, if exists
		ID := pointy.StringValue(widget.ID, "")
		if ID != "" {
			widgetInput.ID = ID
		}
		widgetInput.Configuration = GenerateDashboardWidgetConfigurationInput(widget.Configuration)
		widgetInput.Layout = GenerateDashboardWidgetLayoutInput(widget.Layout)
		widgetInput.Visualization = GenerateDashboardWidgetVisualizationInput(widget.Visualization)
		// Some types of visualizations use raw configuration
		if widget.RawConfiguration != nil {
			widgetInput.RawConfiguration = entities.DashboardWidgetRawConfiguration(pointy.StringValue(widget.RawConfiguration, "{}"))
		}
		input = append(input, widgetInput)
	}

	sort.Slice(input, func(i, j int) bool {
		// Sort first by ID
		if a, b := input[i].ID, input[j].ID; a != b {
			return a < b
		}
		// If there are items with the same ID use the title
		if a, b := input[i].Title, input[j].Title; a != b {
			return a < b
		}
		// Lastly, try sorting by layout as it would be an issue if 2 widgets had the same exact layout
		return strings.Join([]string{strconv.Itoa(input[i].Layout.Row), strconv.Itoa(input[i].Layout.Height), strconv.Itoa(input[i].Layout.Width), strconv.Itoa(input[i].Layout.Column)}, "") < strings.Join([]string{strconv.Itoa(input[j].Layout.Row), strconv.Itoa(input[j].Layout.Height), strconv.Itoa(input[j].Layout.Width), strconv.Itoa(input[j].Layout.Column)}, "")
	})

	return input
}

// GenerateDashboardWidgetConfigurationInput generates an input object
func GenerateDashboardWidgetConfigurationInput(cr v1alpha1.DashboardWidgetConfiguration) dashboards.DashboardWidgetConfigurationInput { // nolint:gocyclo
	input := dashboards.DashboardWidgetConfigurationInput{}

	if cr.Area != nil {
		if cr.Area.NRQLQueries != nil {
			nrqlQueries := make([]dashboards.DashboardWidgetNRQLQueryInput, 0)
			for _, q := range cr.Area.NRQLQueries {
				item := dashboards.DashboardWidgetNRQLQueryInput{AccountID: q.AccountID, Query: nrdb.NRQL(q.Query)}
				nrqlQueries = append(nrqlQueries, item)
			}
			input.Area = &dashboards.DashboardAreaWidgetConfigurationInput{NRQLQueries: nrqlQueries}
		}
	}

	if cr.Bar != nil {
		if cr.Bar.NRQLQueries != nil {
			nrqlQueries := make([]dashboards.DashboardWidgetNRQLQueryInput, 0)
			for _, q := range cr.Bar.NRQLQueries {
				item := dashboards.DashboardWidgetNRQLQueryInput{AccountID: q.AccountID, Query: nrdb.NRQL(q.Query)}
				nrqlQueries = append(nrqlQueries, item)
			}
			input.Bar = &dashboards.DashboardBarWidgetConfigurationInput{NRQLQueries: nrqlQueries}
		}
	}

	if cr.Billboard != nil {
		if cr.Billboard.NRQLQueries != nil {
			nrqlQueries := make([]dashboards.DashboardWidgetNRQLQueryInput, 0)
			for _, q := range cr.Billboard.NRQLQueries {
				item := dashboards.DashboardWidgetNRQLQueryInput{AccountID: q.AccountID, Query: nrdb.NRQL(q.Query)}
				nrqlQueries = append(nrqlQueries, item)
			}

			// Add sorted thresholds
			thresholds := make([]dashboards.DashboardBillboardWidgetThresholdInput, 0)
			alertSeverities := make([]string, 0, len(cr.Billboard.Thresholds))
			for _, threshold := range cr.Billboard.Thresholds {
				alertSeverities = append(alertSeverities, threshold.AlertSeverity)
			}
			sort.Strings(alertSeverities)
			for _, alertSeverity := range alertSeverities {
				for _, q := range cr.Billboard.Thresholds {
					if alertSeverity == q.AlertSeverity {
						thresholdValue, _ := strconv.ParseFloat(q.Value, 64)
						item := dashboards.DashboardBillboardWidgetThresholdInput{AlertSeverity: entities.DashboardAlertSeverity(q.AlertSeverity), Value: pointy.Float64(thresholdValue)}
						thresholds = append(thresholds, item)
					}
				}
			}

			input.Billboard = &dashboards.DashboardBillboardWidgetConfigurationInput{NRQLQueries: nrqlQueries, Thresholds: thresholds}
		}
	}

	if cr.Line != nil {
		if cr.Line.NRQLQueries != nil {
			nrqlQueries := make([]dashboards.DashboardWidgetNRQLQueryInput, 0)
			for _, q := range cr.Line.NRQLQueries {
				item := dashboards.DashboardWidgetNRQLQueryInput{AccountID: q.AccountID, Query: nrdb.NRQL(q.Query)}
				nrqlQueries = append(nrqlQueries, item)
			}
			input.Line = &dashboards.DashboardLineWidgetConfigurationInput{NRQLQueries: nrqlQueries}
		}
	}

	input.Markdown = &dashboards.DashboardMarkdownWidgetConfigurationInput{
		Text: pointy.StringValue(cr.Markdown.Text, ""),
	}

	if cr.Pie != nil {
		if cr.Pie.NRQLQueries != nil {
			nrqlQueries := make([]dashboards.DashboardWidgetNRQLQueryInput, 0)
			for _, q := range cr.Pie.NRQLQueries {
				item := dashboards.DashboardWidgetNRQLQueryInput{AccountID: q.AccountID, Query: nrdb.NRQL(q.Query)}
				nrqlQueries = append(nrqlQueries, item)
			}
			input.Pie = &dashboards.DashboardPieWidgetConfigurationInput{NRQLQueries: nrqlQueries}
		}
	}

	if cr.Table != nil {
		if cr.Table.NRQLQueries != nil {
			nrqlQueries := make([]dashboards.DashboardWidgetNRQLQueryInput, 0)
			for _, q := range cr.Table.NRQLQueries {
				item := dashboards.DashboardWidgetNRQLQueryInput{AccountID: q.AccountID, Query: nrdb.NRQL(q.Query)}
				nrqlQueries = append(nrqlQueries, item)
			}
			input.Table = &dashboards.DashboardTableWidgetConfigurationInput{NRQLQueries: nrqlQueries}
		}
	}
	return input
}

// GenerateDashboardWidgetLayoutInput generates an input object
func GenerateDashboardWidgetLayoutInput(cr v1alpha1.DashboardWidgetLayout) dashboards.DashboardWidgetLayoutInput {
	input := dashboards.DashboardWidgetLayoutInput{
		Column: cr.Column,
		Row:    cr.Row,
		Height: cr.Height,
		Width:  cr.Width,
	}
	return input
}

// GenerateDashboardWidgetVisualizationInput generates an input object
func GenerateDashboardWidgetVisualizationInput(cr v1alpha1.DashboardWidgetVisualization) dashboards.DashboardWidgetVisualizationInput {
	input := dashboards.DashboardWidgetVisualizationInput{ID: cr.ID}
	return input
}
