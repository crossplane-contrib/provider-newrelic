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

// AlertsPolicyParameters are the configurable fields of a Policy.
type AlertsPolicyParameters struct {
	ID string `json:"id,omitempty"`
	// +kubebuilder:validation:Enum=PER_CONDITION;PER_CONDITION_AND_TARGET;PER_POLICY
	IncidentPreference string `json:"incidentPreference"`
	Name               string `json:"name"`
	ChannelIDs         []int  `json:"channelIds"`
}

// AlertsPolicyObservation are the observable fields of a Policy.
type AlertsPolicyObservation struct {
	// The stable and unique string id from NewRelic.
	ID              string `json:"id,omitempty"`
	ObservableField string `json:"observableField,omitempty"`
}

// A AlertsPolicySpec defines the desired state of a Policy.
type AlertsPolicySpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       AlertsPolicyParameters `json:"forProvider"`
}

// A AlertsPolicyStatus represents the observed state of a Policy.
type AlertsPolicyStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          AlertsPolicyObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A AlertsPolicy is an example API type.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="ID",type="string",JSONPath=".status.atProvider.id"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,newrelic}
type AlertsPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AlertsPolicySpec   `json:"spec"`
	Status AlertsPolicyStatus `json:"status,omitempty"`
}

// SetPublishConnectionDetailsTo is a func for connection details
func (in *AlertsPolicy) SetPublishConnectionDetailsTo(r *xpv1.PublishConnectionDetailsTo) {
	// TODO implement me
	panic("implement me")
}

// GetPublishConnectionDetailsTo is a func for connection details
func (in *AlertsPolicy) GetPublishConnectionDetailsTo() *xpv1.PublishConnectionDetailsTo {
	// TODO implement me
	panic("implement me")
}

// +kubebuilder:object:root=true

// AlertsPolicyList contains a list of Policy
type AlertsPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AlertsPolicy `json:"items"`
}
