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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// https://rpm.newrelic.com/api/explore/alerts_policies/list

// DashboardParameters are the configurable fields of a Policy.
type DashboardParameters struct {
	// Account ID.
	AccountID int `json:"accountId,omitempty"`

	// Dashboard description.
	Description *string `json:"description,omitempty"`
	// Unique entity identifier.
	GUID string `json:"guid,omitempty"`
	// Dashboard name.
	Name string `json:"name,omitempty"`

	// Dashboard pages.
	Pages []DashboardPage `json:"pages,omitempty"`
	// Dashboard permissions configuration.
	// +kubebuilder:validation:Enum=PUBLIC_READ_WRITE;PUBLIC_READ_ONLY;PRIVATE
	Permissions *string `json:"permissions,omitempty"`
}

// DashboardPage is a type of resource
type DashboardPage struct {
	// Page description.
	Description *string `json:"description,omitempty"`
	// Unique entity identifier.
	GUID string `json:"guid,omitempty"`
	// Page name.
	Name string `json:"name,omitempty"`
	// Page widgets.
	Widgets []DashboardWidget `json:"widgets,omitempty"`
}

// DashboardWidgetConfiguration - Typed configuration for known visualizations. Only one (at most) will be populated for a given widget.
type DashboardWidgetConfiguration struct {
	// Configuration for visualization type 'viz.area'
	Area *DashboardAreaWidgetConfiguration `json:"area,omitempty"`
	// Configuration for visualization type 'viz.bar'
	Bar *DashboardBarWidgetConfiguration `json:"bar,omitempty"`
	// Configuration for visualization type 'viz.billboard'
	Billboard *DashboardBillboardWidgetConfiguration `json:"billboard,omitempty"`
	// Configuration for visualization type 'viz.line'
	Line *DashboardLineWidgetConfiguration `json:"line,omitempty"`
	// Configuration for visualization type 'viz.markdown'
	// +optional
	Markdown DashboardMarkdownWidgetConfiguration `json:"markdown"`
	// Configuration for visualization type 'viz.pie'
	Pie *DashboardPieWidgetConfiguration `json:"pie,omitempty"`
	// Configuration for visualization type 'viz.table'
	Table *DashboardTableWidgetConfiguration `json:"table,omitempty"`
}

// DashboardWidget - Widgets in a Dashboard Page.
type DashboardWidget struct {
	// Typed configuration
	Configuration DashboardWidgetConfiguration `json:"configuration,omitempty"`
	// id
	ID *string `json:"id,omitempty"`
	// layout
	Layout DashboardWidgetLayout `json:"layout,omitempty"`

	// Untyped configuration
	RawConfiguration *string `json:"rawConfiguration,omitempty"`
	// title
	Title string `json:"title,omitempty"`
	// Specifies how this widget will be visualized.
	Visualization DashboardWidgetVisualization `json:"visualization"`
}

// DashboardPieWidgetConfiguration - Configuration for visualization type 'viz.pie'
type DashboardPieWidgetConfiguration struct {
	// nrql queries
	NRQLQueries []DashboardWidgetNRQLQuery `json:"nrqlQueries,omitempty"`
}

// DashboardTableWidgetConfiguration - Configuration for visualization type 'viz.table'
type DashboardTableWidgetConfiguration struct {
	// nrql queries
	NRQLQueries []DashboardWidgetNRQLQuery `json:"nrqlQueries,omitempty"`
}

// DashboardMarkdownWidgetConfiguration - Configuration for visualization type 'viz.markdown'
type DashboardMarkdownWidgetConfiguration struct {
	// Markdown content of the widget
	// +kubebuilder:default:=""
	// +nullable
	Text *string `json:"text"`
}

// DashboardLineWidgetConfiguration - Configuration for visualization type 'viz.line'
type DashboardLineWidgetConfiguration struct {
	// nrql queries
	NRQLQueries []DashboardWidgetNRQLQuery `json:"nrqlQueries,omitempty"`
}

// DashboardWidgetVisualization - Visualization configuration
type DashboardWidgetVisualization struct {
	// Nerdpack artifact ID
	// +kubebuilder:validation:Enum=viz.area;viz.bar;viz.billboard;viz.bullet;viz.funnel;viz.heatmap;viz.histogram;viz.json;viz.line;viz.markdown;viz.pie;viz.stacked-bar;viz.table
	ID string `json:"id,omitempty"`
}

// DashboardWidgetLayout - Widget layout.
type DashboardWidgetLayout struct {
	// Column
	// +kubebuilder:validation:Minimum=1
	Column int `json:"column,omitempty"`
	// Height
	// +kubebuilder:validation:Minimum=1
	Height int `json:"height,omitempty"`
	// Row
	// +kubebuilder:validation:Minimum=1
	Row int `json:"row,omitempty"`
	// Width
	// +kubebuilder:validation:Minimum=1
	Width int `json:"width,omitempty"`
}

// DashboardAreaWidgetConfiguration - Configuration for visualization type 'viz.area'
type DashboardAreaWidgetConfiguration struct {
	// nrql queries
	NRQLQueries []DashboardWidgetNRQLQuery `json:"nrqlQueries,omitempty"`
}

// DashboardWidgetNRQLQuery - Single NRQL query for a widget.
type DashboardWidgetNRQLQuery struct {
	// accountId
	AccountID int `json:"accountId"`
	// NRQL formatted query
	Query string `json:"query"`
}

// DashboardBarWidgetConfiguration - Configuration for visualization type 'viz.bar'
type DashboardBarWidgetConfiguration struct {
	// nrql queries
	NRQLQueries []DashboardWidgetNRQLQuery `json:"nrqlQueries,omitempty"`
}

// DashboardBillboardWidgetConfiguration - Configuration for visualization type 'viz.billboard'
type DashboardBillboardWidgetConfiguration struct {
	// nrql queries
	NRQLQueries []DashboardWidgetNRQLQuery `json:"nrqlQueries,omitempty"`
	// Thresholds
	Thresholds []DashboardBillboardWidgetThreshold `json:"thresholds,omitempty"`
}

// DashboardBillboardWidgetThreshold - Billboard widget threshold.
type DashboardBillboardWidgetThreshold struct {
	// Alert severity.
	AlertSeverity string `json:"alertSeverity,omitempty"`
	// Alert value.
	Value string `json:"value,omitempty"`
}

// DashboardObservation are the observable fields of a Policy.
type DashboardObservation struct {
	// The stable and unique string guid from NewRelic.
	GUID            string `json:"guid,omitempty"`
	ObservableField string `json:"observableField,omitempty"`
}

// A DashboardSpec defines the desired state of a Policy.
type DashboardSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       DashboardParameters `json:"forProvider"`
}

// A DashboardStatus represents the observed state of a Policy.
type DashboardStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          DashboardObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Dashboard is an example API type.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="GUID",type="string",JSONPath=".status.atProvider.guid"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,newrelic}
type Dashboard struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DashboardSpec   `json:"spec"`
	Status DashboardStatus `json:"status,omitempty"`
}

// SetPublishConnectionDetailsTo is a func for connection details
func (in *Dashboard) SetPublishConnectionDetailsTo(r *xpv1.PublishConnectionDetailsTo) {
	// TODO implement me
	panic("implement me")
}

// GetPublishConnectionDetailsTo is a func for connection details
func (in *Dashboard) GetPublishConnectionDetailsTo() *xpv1.PublishConnectionDetailsTo {
	// TODO implement me
	panic("implement me")
}

// +kubebuilder:object:root=true

// DashboardList contains a list of Policy
type DashboardList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Dashboard `json:"items"`
}
