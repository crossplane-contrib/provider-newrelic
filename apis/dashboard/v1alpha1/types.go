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

	// Dashboard variables
	Variables []DashboardVariable `json:"variables,omitempty"`
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

// DashboardWidget - Widgets in a Dashboard Page.
type DashboardWidget struct {
	// id
	ID *string `json:"id,omitempty"`
	// layout
	Layout DashboardWidgetLayout `json:"layout,omitempty"`
	// Untyped configuration
	RawConfiguration *DashboardWidgetRawConfiguration `json:"rawConfiguration,omitempty"`
	// title
	Title string `json:"title,omitempty"`
	// Specifies how this widget will be visualized.
	Visualization DashboardWidgetVisualization `json:"visualization"`
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

// DashboardWidgetRawConfiguration represents the configuration for widgets, it's a replacement for configuration field
type DashboardWidgetRawConfiguration struct {
	// Used by all widgets
	NRQLQueries     *[]DashboardWidgetNRQLQueryInput `json:"nrqlQueries,omitempty"`
	PlatformOptions *RawConfigurationPlatformOptions `json:"platformOptions,omitempty"`

	// Used by viz.bullet
	Limit *float64 `json:"limit,omitempty"`

	// Used by viz.markdown
	Text *string `json:"text,omitempty"`

	// Used by viz.billboard
	Thresholds []DashboardBillboardWidgetThresholdInput `json:"thresholds,omitempty"`
}

// DashboardBillboardWidgetThresholdInput - used by Billboard Widgets
type DashboardBillboardWidgetThresholdInput struct {
	// alert severity.
	// +kubebuilder:validation:Enum=CRITICAL;NOT_ALERTING;WARNING
	AlertSeverity *string `json:"alertSeverity,omitempty"`
	// value.
	Value *float64 `json:"value,omitempty"`
}

// DashboardWidgetNRQLQueryInput - NRQL query used by a widget
type DashboardWidgetNRQLQueryInput struct {
	// accountId
	AccountID int `json:"accountId"`
	// NRQL formatted query
	Query string `json:"query"`
}

// RawConfigurationPlatformOptions represents the platform widget options
type RawConfigurationPlatformOptions struct {
	IgnoreTimeRange bool `json:"ignoreTimeRange,omitempty"`
}

// DashboardVariable - Definition of a variable that is local to this dashboard. Variables are placeholders for dynamic values in widget NRQLs.
type DashboardVariable struct {
	// Default values for this variable. The actual value to be used will depend on the type.
	DefaultValues *[]DashboardVariableDefaultItem `json:"defaultValues,omitempty"`
	// Indicates whether this variable supports multiple selection or not. Only applies to variables of type NRQL or ENUM.
	IsMultiSelection bool `json:"isMultiSelection,omitempty"`
	// List of possible values for variables of type ENUM.
	Items []DashboardVariableEnumItem `json:"items,omitempty"`
	// Configuration for variables of type NRQL.
	NRQLQuery *DashboardVariableNRQLQuery `json:"nrqlQuery,omitempty"`
	// Variable identifier.
	Name string `json:"name,omitempty"`
	// Indicates the strategy to apply when replacing a variable in a NRQL query.
	// +kubebuilder:validation:Enum=DEFAULT;IDENTIFIER;NUMBER;STRING
	ReplacementStrategy string `json:"replacementStrategy,omitempty"`
	// Human-friendly display string for this variable.
	Title string `json:"title,omitempty"`
	// Specifies the data type of the variable and where its possible values may come from.
	// +kubebuilder:validation:Enum=ENUM;NRQL;STRING
	Type string `json:"type,omitempty"`
}

// DashboardVariableDefaultItem - Represents a possible default value item.
type DashboardVariableDefaultItem struct {
	// The value of this default item.
	Value DashboardVariableDefaultValue `json:"value,omitempty"`
}

// DashboardVariableDefaultValue - Specifies a default value for variables.
type DashboardVariableDefaultValue struct {
	// Default string value.
	String string `json:"string,omitempty"`
}

// DashboardVariableEnumItem - Represents a possible value for a variable of type ENUM.
type DashboardVariableEnumItem struct {
	// A human-friendly display string for this value.
	Title string `json:"title,omitempty"`
	// A possible variable value.
	Value string `json:"value,omitempty"`
}

// DashboardVariableNRQLQuery - Configuration for variables of type NRQL.
type DashboardVariableNRQLQuery struct {
	// New Relic account ID(s) to issue the query against.
	AccountIDs []int `json:"accountIds,omitempty"`
	// NRQL formatted query.
	Query string `json:"query"`
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

// DashboardList contains a list of Dashboard
type DashboardList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Dashboard `json:"items"`
}
