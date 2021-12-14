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

// https://rpm.newrelic.com/api/explore/alerts_nrql_conditions/list

// NrqlAlertConditionParameters are the configurable fields of a Condition
type NrqlAlertConditionParameters struct {
	ID string `json:"id,omitempty"`
	// +kubebuilder:validation:Enum=STATIC;BASELINE;OUTLIER
	Type       string  `json:"type,omitempty"`
	Name       string  `json:"name"`
	RunbookURL *string `json:"runbookUrl,omitempty"`
	Enabled    bool    `json:"enabled"`
	// +kubebuilder:validation:Maximum=2592000
	// +kubebuilder:validation:Minimum=300
	ViolationTimeLimitSeconds *int `json:"violationTimeLimitSeconds,omitempty"`
	// +kubebuilder:validation:MaxItems=2
	Terms       []NrqlConditionTerm `json:"terms"`
	Nrql        Nrql                `json:"nrql"`
	Signal      Signal              `json:"signal"`
	Expiration  *Expiration         `json:"expiration,omitempty"`
	Description *string             `json:"description,omitempty"`
	// +kubebuilder:validation:Enum=SINGLE_VALUE;SUM
	ValueFunction *string `json:"valueFunction,omitempty"`

	// +kubebuilder:validation:Enum=LOWER_ONLY;UPPER_AND_LOWER;UPPER_ONLY
	BaselineDirection *string `json:"baselineDirection,omitempty"`

	// Below are referenced items
	AlertsPolicyID string `json:"policyId,omitempty"`

	// AlertPolicyRef is a reference to an AlertPolicy used to set
	// the PolicyID.
	// +optional
	AlertsPolicyRef *xpv1.Reference `json:"alertsPolicyRef,omitempty"`

	// AlertPolicySelector selects references to an AlertPolicy used
	// to set the AlertPolicyID.
	// +optional
	AlertsPolicySelector *xpv1.Selector `json:"alertsPolicySelector,omitempty"`
}

// NrqlConditionTerm are the configurable fields of a Condition
type NrqlConditionTerm struct {
	// +kubebuilder:validation:Enum=ABOVE;BELOW;EQUALS
	Operator string `json:"operator,omitempty"`
	// +kubebuilder:validation:Enum=CRITICAL;WARNING
	Priority  string `json:"priority,omitempty"`
	Threshold string `json:"threshold"`
	// +kubebuilder:validation:Maximum=7200
	// +kubebuilder:validation:Minimum=60
	// +kubebuilder:validation:MultipleOf=60
	ThresholdDuration int `json:"thresholdDuration,omitempty"`
	// +kubebuilder:validation:Enum=ALL;AT_LEAST_ONCE
	ThresholdOccurrences string `json:"thresholdOccurrences,omitempty"`
}

// Nrql are the configurable fields of a Condition
type Nrql struct {
	Query string `json:"query"`
}

// Signal are the configurable fields of a Condition
type Signal struct {
	// +kubebuilder:validation:Minimum=60
	// +kubebuilder:validation:MultipleOf=60
	AggregationWindow *int `json:"aggregationWindow,omitempty"`
	// +kubebuilder:validation:Enum=LAST_VALUE;NONE;STATIC
	FillOption string  `json:"fillOption"`
	FillValue  *string `json:"fillValue,omitempty"`
	// +kubebuilder:validation:Enum=CADENCE;EVENT_FLOW;EVENT_TIMER
	AggregationMethod *string `json:"aggregationMethod,omitempty"`
	AggregationDelay  *int    `json:"aggregationDelay,omitempty"`
	AggregationTimer  *int    `json:"aggregationTimer,omitempty"`
}

// Expiration are the configurable fields of a Condition
type Expiration struct {
	ExpirationDuration          *int `json:"expirationDuration,omitempty"`
	OpenViolationOnExpiration   bool `json:"openViolationOnExpiration"`
	CloseViolationsOnExpiration bool `json:"closeViolationsOnExpiration"`
}

// NrqlAlertConditionObservation are the observable fields of a Condition.
type NrqlAlertConditionObservation struct {
	// The stable and unique string id from NewRelic.
	ID              string `json:"id,omitempty"`
	ObservableField string `json:"observableField,omitempty"`
}

// A NrqlAlertConditionSpec defines the desired state of a Condition.
type NrqlAlertConditionSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       NrqlAlertConditionParameters `json:"forProvider"`
}

// A NrqlAlertConditionStatus represents the observed state of a Condition.
type NrqlAlertConditionStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          NrqlAlertConditionObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A NrqlAlertCondition is an example API type.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="ID",type="string",JSONPath=".status.atProvider.id"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,newrelic}
type NrqlAlertCondition struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NrqlAlertConditionSpec   `json:"spec"`
	Status NrqlAlertConditionStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NrqlAlertConditionList contains a list of Condition
type NrqlAlertConditionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NrqlAlertCondition `json:"items"`
}
