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

package nrqlalertcondition

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/newrelic/newrelic-client-go/v2/pkg/alerts"
	"github.com/openlyinc/pointy"

	"github.com/crossplane/crossplane-runtime/pkg/meta"

	"github.com/crossplane-contrib/provider-newrelic/apis/nrqlalertcondition/v1alpha1"
)

type NrqlAlertConditionModifier func(*v1alpha1.NrqlAlertCondition)

func NrqlAlertCondition(m ...NrqlAlertConditionModifier) *v1alpha1.NrqlAlertCondition {
	cr := &v1alpha1.NrqlAlertCondition{
		Spec: v1alpha1.NrqlAlertConditionSpec{
			ForProvider: v1alpha1.NrqlAlertConditionParameters{
				ID:         "1",
				Name:       "test_nrql",
				Type:       "STATIC",
				RunbookURL: pointy.String("runbookUrl"),
				Enabled:    false,
				//ViolationTimeLimit:        "average", // alerts.ValueFunctionTypes.Average,
				ViolationTimeLimitSeconds: pointy.Int(2592000),
				ValueFunction:             pointy.String("SINGLE_VALUE"),
				Terms: []v1alpha1.NrqlConditionTerm{
					{ThresholdDuration: 60, Operator: "ABOVE", Priority: "WARNING", Threshold: "2", ThresholdOccurrences: "ALL"},
					{ThresholdDuration: 300, Operator: "ABOVE", Priority: "CRITICAL", Threshold: "5", ThresholdOccurrences: "ALL"},
				},
				Signal: v1alpha1.Signal{
					AggregationMethod: pointy.String("EVENT_FLOW"),
					AggregationDelay:  pointy.Int(120),
					AggregationWindow: pointy.Int(60),
					FillOption:        "NONE",
				},
				Expiration: &v1alpha1.Expiration{
					ExpirationDuration:          nil,
					CloseViolationsOnExpiration: false,
					OpenViolationOnExpiration:   false,
				},
				Nrql: v1alpha1.Nrql{Query: "SELECT latest(controller_runtime_reconcile_errors_total) FROM Metric"},
			},
		},
	}
	meta.SetExternalName(cr, "1")
	for _, f := range m {
		f(cr)
	}
	cr.Spec.ForProvider.AlertsPolicyID = "1375108"
	return cr
}

func GetAlertsFillOptionPointer(value string) *alerts.AlertsFillOption {
	o := alerts.AlertsFillOption(value)
	return &o
}

func TestIsUpToDate(t *testing.T) {

	type args struct {
		cr v1alpha1.NrqlAlertCondition
		nr alerts.NrqlAlertCondition
	}

	type want struct {
		expected bool
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"DiffType": {
			args: args{cr: *NrqlAlertCondition(),
				nr: alerts.NrqlAlertCondition{
					NrqlConditionBase: alerts.NrqlConditionBase{Name: "test_nrql",
						Type:                      alerts.NrqlConditionTypes.Baseline,
						RunbookURL:                "runbookUrl",
						Enabled:                   false,
						ViolationTimeLimit:        "average",
						ViolationTimeLimitSeconds: 2592000,

						Terms: []alerts.NrqlConditionTerm{
							{ThresholdDuration: 60, Operator: "ABOVE", Priority: "WARNING", Threshold: pointy.Float64(2.0), ThresholdOccurrences: "ALL"},
							{ThresholdDuration: 300, Operator: "ABOVE", Priority: "CRITICAL", Threshold: pointy.Float64(5.0), ThresholdOccurrences: "ALL"},
						},
						Signal: &alerts.AlertsNrqlConditionSignal{
							AggregationDelay:  pointy.Int(120),
							AggregationMethod: &alerts.NrqlConditionAggregationMethodTypes.EventFlow,
							AggregationWindow: pointy.Int(60),
							FillOption:        GetAlertsFillOptionPointer("NONE"),
						},
						Expiration: &alerts.AlertsNrqlConditionExpiration{
							ExpirationDuration:          nil,
							CloseViolationsOnExpiration: false,
							OpenViolationOnExpiration:   false,
						},
						Nrql: alerts.NrqlConditionQuery{Query: "SELECT latest(controller_runtime_reconcile_errors_total) FROM Metric"},
					},
				},
			},
			want: want{expected: false},
		},
		"DiffRunbookUrl": {
			args: args{cr: *NrqlAlertCondition(),
				nr: alerts.NrqlAlertCondition{
					NrqlConditionBase: alerts.NrqlConditionBase{Name: "test_nrql",
						Type:                      alerts.NrqlConditionTypes.Static,
						RunbookURL:                "runbookUrl_is_different",
						Enabled:                   false,
						ViolationTimeLimit:        "average",
						ViolationTimeLimitSeconds: 2592000,

						Terms: []alerts.NrqlConditionTerm{
							{ThresholdDuration: 60, Operator: "ABOVE", Priority: "WARNING", Threshold: pointy.Float64(2.0), ThresholdOccurrences: "ALL"},
							{ThresholdDuration: 300, Operator: "ABOVE", Priority: "CRITICAL", Threshold: pointy.Float64(5.0), ThresholdOccurrences: "ALL"},
						},
						Signal: &alerts.AlertsNrqlConditionSignal{
							AggregationDelay:  pointy.Int(120),
							AggregationMethod: &alerts.NrqlConditionAggregationMethodTypes.EventFlow,
							AggregationWindow: pointy.Int(60),
							FillOption:        GetAlertsFillOptionPointer("NONE"),
						},
						Expiration: &alerts.AlertsNrqlConditionExpiration{
							ExpirationDuration:          nil,
							CloseViolationsOnExpiration: false,
							OpenViolationOnExpiration:   false,
						},
						Nrql: alerts.NrqlConditionQuery{Query: "SELECT latest(controller_runtime_reconcile_errors_total) FROM Metric"},
					},
				},
			},
			want: want{expected: false},
		},
		"DiffEnabled": {
			args: args{cr: *NrqlAlertCondition(),
				nr: alerts.NrqlAlertCondition{
					NrqlConditionBase: alerts.NrqlConditionBase{Name: "test_nrql",
						Type:                      alerts.NrqlConditionTypes.Static,
						RunbookURL:                "runbookUrl",
						Enabled:                   true,
						ViolationTimeLimit:        "average",
						ViolationTimeLimitSeconds: 2592000,
						Terms: []alerts.NrqlConditionTerm{
							{ThresholdDuration: 60, Operator: "ABOVE", Priority: "WARNING", Threshold: pointy.Float64(2.0), ThresholdOccurrences: "ALL"},
							{ThresholdDuration: 300, Operator: "ABOVE", Priority: "CRITICAL", Threshold: pointy.Float64(5.0), ThresholdOccurrences: "ALL"},
						},
						Signal: &alerts.AlertsNrqlConditionSignal{
							AggregationDelay:  pointy.Int(120),
							AggregationMethod: &alerts.NrqlConditionAggregationMethodTypes.EventFlow,
							AggregationWindow: pointy.Int(60),
							FillOption:        GetAlertsFillOptionPointer("NONE"),
						},
						Expiration: &alerts.AlertsNrqlConditionExpiration{
							ExpirationDuration:          nil,
							CloseViolationsOnExpiration: false,
							OpenViolationOnExpiration:   false,
						},
						Nrql: alerts.NrqlConditionQuery{Query: "SELECT latest(controller_runtime_reconcile_errors_total) FROM Metric"},
					},
				},
			},
			want: want{expected: false},
		},
		"DiffViolationTimeLimitSeconds": {
			args: args{cr: *NrqlAlertCondition(),
				nr: alerts.NrqlAlertCondition{
					NrqlConditionBase: alerts.NrqlConditionBase{Name: "test_nrql",
						Type:                      alerts.NrqlConditionTypes.Static,
						RunbookURL:                "runbookUrl",
						Enabled:                   false,
						ViolationTimeLimit:        "average",
						ViolationTimeLimitSeconds: 300,
						Terms: []alerts.NrqlConditionTerm{
							{ThresholdDuration: 60, Operator: "ABOVE", Priority: "WARNING", Threshold: pointy.Float64(2.0), ThresholdOccurrences: "ALL"},
							{ThresholdDuration: 300, Operator: "ABOVE", Priority: "CRITICAL", Threshold: pointy.Float64(5.0), ThresholdOccurrences: "ALL"},
						},
						Signal: &alerts.AlertsNrqlConditionSignal{
							AggregationDelay:  pointy.Int(120),
							AggregationMethod: &alerts.NrqlConditionAggregationMethodTypes.EventFlow,
							AggregationWindow: pointy.Int(60),
							FillOption:        GetAlertsFillOptionPointer("NONE"),
						},
						Expiration: &alerts.AlertsNrqlConditionExpiration{
							ExpirationDuration:          nil,
							CloseViolationsOnExpiration: false,
							OpenViolationOnExpiration:   false,
						},
						Nrql: alerts.NrqlConditionQuery{Query: "SELECT latest(controller_runtime_reconcile_errors_total) FROM Metric"},
					},
				},
			},
			want: want{expected: false},
		},
		"DiffTerms": {
			args: args{cr: *NrqlAlertCondition(),
				nr: alerts.NrqlAlertCondition{
					NrqlConditionBase: alerts.NrqlConditionBase{Name: "test_nrql",
						Type:                      alerts.NrqlConditionTypes.Static,
						RunbookURL:                "runbookUrl",
						Enabled:                   false,
						ViolationTimeLimit:        "average",
						ViolationTimeLimitSeconds: 2592000,
						Terms: []alerts.NrqlConditionTerm{
							{ThresholdDuration: 60, Operator: "ABOVE", Priority: "WARNING", Threshold: pointy.Float64(2.0), ThresholdOccurrences: "ALL"},
							{ThresholdDuration: 300, Operator: "ABOVE", Priority: "CRITICAL", Threshold: pointy.Float64(7.0), ThresholdOccurrences: "ALL"},
						},
						Signal: &alerts.AlertsNrqlConditionSignal{
							AggregationDelay:  pointy.Int(120),
							AggregationMethod: &alerts.NrqlConditionAggregationMethodTypes.EventFlow,
							AggregationWindow: pointy.Int(60),
							FillOption:        GetAlertsFillOptionPointer("NONE"),
						},
						Expiration: &alerts.AlertsNrqlConditionExpiration{
							ExpirationDuration:          nil,
							CloseViolationsOnExpiration: false,
							OpenViolationOnExpiration:   false,
						},
						Nrql: alerts.NrqlConditionQuery{Query: "SELECT latest(controller_runtime_reconcile_errors_total) FROM Metric"},
					},
				},
			},
			want: want{expected: false},
		},
		"DiffNrqlQuery": {
			args: args{cr: *NrqlAlertCondition(),
				nr: alerts.NrqlAlertCondition{
					NrqlConditionBase: alerts.NrqlConditionBase{Name: "test_nrql",
						Type:                      alerts.NrqlConditionTypes.Static,
						RunbookURL:                "runbookUrl",
						Enabled:                   false,
						ViolationTimeLimit:        "average",
						ViolationTimeLimitSeconds: 2592000,
						Terms: []alerts.NrqlConditionTerm{
							{ThresholdDuration: 60, Operator: "ABOVE", Priority: "WARNING", Threshold: pointy.Float64(2.0), ThresholdOccurrences: "ALL"},
							{ThresholdDuration: 300, Operator: "ABOVE", Priority: "CRITICAL", Threshold: pointy.Float64(5.0), ThresholdOccurrences: "ALL"},
						},
						Signal: &alerts.AlertsNrqlConditionSignal{
							AggregationDelay:  pointy.Int(120),
							AggregationMethod: &alerts.NrqlConditionAggregationMethodTypes.EventFlow,
							AggregationWindow: pointy.Int(60),
							FillOption:        GetAlertsFillOptionPointer("NONE"),
						},
						Expiration: &alerts.AlertsNrqlConditionExpiration{
							ExpirationDuration:          nil,
							CloseViolationsOnExpiration: false,
							OpenViolationOnExpiration:   false,
						},
						Nrql: alerts.NrqlConditionQuery{Query: "SELECT latest(controller_runtime_reconcile_errors_total) FROM Metric FACET difference"},
					},
				},
			},
			want: want{expected: false},
		},
		"DiffNrqlAggregationDelay": {
			args: args{cr: *NrqlAlertCondition(),
				nr: alerts.NrqlAlertCondition{
					NrqlConditionBase: alerts.NrqlConditionBase{Name: "test_nrql",
						Type:                      alerts.NrqlConditionTypes.Static,
						RunbookURL:                "runbookUrl",
						Enabled:                   false,
						ViolationTimeLimit:        "average",
						ViolationTimeLimitSeconds: 2592000,
						Terms: []alerts.NrqlConditionTerm{
							{ThresholdDuration: 60, Operator: "ABOVE", Priority: "WARNING", Threshold: pointy.Float64(2.0), ThresholdOccurrences: "ALL"},
							{ThresholdDuration: 300, Operator: "ABOVE", Priority: "CRITICAL", Threshold: pointy.Float64(5.0), ThresholdOccurrences: "ALL"},
						},
						Signal: &alerts.AlertsNrqlConditionSignal{
							AggregationDelay:  pointy.Int(300),
							AggregationMethod: &alerts.NrqlConditionAggregationMethodTypes.EventFlow,
							AggregationWindow: pointy.Int(60),
							FillOption:        GetAlertsFillOptionPointer("NONE"),
						},
						Expiration: &alerts.AlertsNrqlConditionExpiration{
							ExpirationDuration:          nil,
							CloseViolationsOnExpiration: false,
							OpenViolationOnExpiration:   false,
						},
						Nrql: alerts.NrqlConditionQuery{Query: "SELECT latest(controller_runtime_reconcile_errors_total) FROM Metric"},
					},
				},
			},
			want: want{expected: false},
		},
		"DiffNrqlAggregationMethod": {
			args: args{cr: *NrqlAlertCondition(),
				nr: alerts.NrqlAlertCondition{
					NrqlConditionBase: alerts.NrqlConditionBase{Name: "test_nrql",
						Type:                      alerts.NrqlConditionTypes.Static,
						RunbookURL:                "runbookUrl",
						Enabled:                   false,
						ViolationTimeLimit:        "average",
						ViolationTimeLimitSeconds: 2592000,
						Terms: []alerts.NrqlConditionTerm{
							{ThresholdDuration: 60, Operator: "ABOVE", Priority: "WARNING", Threshold: pointy.Float64(2.0), ThresholdOccurrences: "ALL"},
							{ThresholdDuration: 300, Operator: "ABOVE", Priority: "CRITICAL", Threshold: pointy.Float64(5.0), ThresholdOccurrences: "ALL"},
						},
						Signal: &alerts.AlertsNrqlConditionSignal{
							AggregationDelay:  pointy.Int(120),
							AggregationMethod: &alerts.NrqlConditionAggregationMethodTypes.Cadence,
							AggregationWindow: pointy.Int(60),
							FillOption:        GetAlertsFillOptionPointer("NONE"),
						},
						Expiration: &alerts.AlertsNrqlConditionExpiration{
							ExpirationDuration:          nil,
							CloseViolationsOnExpiration: false,
							OpenViolationOnExpiration:   false,
						},
						Nrql: alerts.NrqlConditionQuery{Query: "SELECT latest(controller_runtime_reconcile_errors_total) FROM Metric"},
					},
				},
			},
			want: want{expected: false},
		},
		"TermsOutOfOrder-Same": {
			args: args{cr: *NrqlAlertCondition(),
				nr: alerts.NrqlAlertCondition{
					NrqlConditionBase: alerts.NrqlConditionBase{Name: "test_nrql",
						Type:                      alerts.NrqlConditionTypes.Static,
						RunbookURL:                "runbookUrl",
						Enabled:                   false,
						ViolationTimeLimit:        "average",
						ViolationTimeLimitSeconds: 2592000,
						Terms: []alerts.NrqlConditionTerm{
							{ThresholdDuration: 300, Operator: "ABOVE", Priority: "CRITICAL", Threshold: pointy.Float64(5.0), ThresholdOccurrences: "ALL"},
							{ThresholdDuration: 60, Operator: "ABOVE", Priority: "WARNING", Threshold: pointy.Float64(2.0), ThresholdOccurrences: "ALL"},
						},
						Signal: &alerts.AlertsNrqlConditionSignal{
							AggregationDelay:  pointy.Int(120),
							AggregationMethod: &alerts.NrqlConditionAggregationMethodTypes.EventFlow,
							AggregationWindow: pointy.Int(60),
							FillOption:        GetAlertsFillOptionPointer("NONE"),
						},
						Expiration: &alerts.AlertsNrqlConditionExpiration{
							ExpirationDuration:          nil,
							CloseViolationsOnExpiration: false,
							OpenViolationOnExpiration:   false,
						},
						Nrql: alerts.NrqlConditionQuery{Query: "SELECT latest(controller_runtime_reconcile_errors_total) FROM Metric"},
					},
				},
			},
			want: want{expected: true},
		},
		"Same": {
			args: args{cr: *NrqlAlertCondition(),
				nr: alerts.NrqlAlertCondition{
					NrqlConditionBase: alerts.NrqlConditionBase{Name: "test_nrql",
						Type:                      alerts.NrqlConditionTypes.Static,
						RunbookURL:                "runbookUrl",
						Enabled:                   false,
						ViolationTimeLimit:        "average",
						ViolationTimeLimitSeconds: 2592000,
						Terms: []alerts.NrqlConditionTerm{
							{ThresholdDuration: 60, Operator: "ABOVE", Priority: "WARNING", Threshold: pointy.Float64(2.0), ThresholdOccurrences: "ALL"},
							{ThresholdDuration: 300, Operator: "ABOVE", Priority: "CRITICAL", Threshold: pointy.Float64(5.0), ThresholdOccurrences: "ALL"},
						},
						Signal: &alerts.AlertsNrqlConditionSignal{
							AggregationDelay:  pointy.Int(120),
							AggregationMethod: &alerts.NrqlConditionAggregationMethodTypes.EventFlow,
							AggregationWindow: pointy.Int(60),
							FillOption:        GetAlertsFillOptionPointer("NONE"),
						},
						Expiration: &alerts.AlertsNrqlConditionExpiration{
							ExpirationDuration:          nil,
							CloseViolationsOnExpiration: false,
							OpenViolationOnExpiration:   false,
						},
						Nrql: alerts.NrqlConditionQuery{Query: "SELECT latest(controller_runtime_reconcile_errors_total) FROM Metric"},
					},
				},
			},
			want: want{expected: true},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := IsUpToDate(&tc.args.cr, &tc.args.nr)
			if diff := cmp.Diff(tc.want.expected, got); diff != "" {
				t.Errorf("e.TestIsUpToDate(...): -want, +got:\n%s\n", diff)
			}
		})
	}
}

func TestGenerateNrqlConditionUpdateInput(t *testing.T) {

	type args struct {
		cr v1alpha1.NrqlAlertCondition
	}

	type want struct {
		expected alerts.NrqlConditionUpdateInput
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"Same": {
			args: args{cr: *NrqlAlertCondition()},
			want: want{expected: alerts.NrqlConditionUpdateInput{
				ValueFunction: &alerts.NrqlConditionValueFunctions.SingleValue,
				NrqlConditionUpdateBase: alerts.NrqlConditionUpdateBase{Name: "test_nrql",
					Type:                      alerts.NrqlConditionTypes.Static,
					RunbookURL:                "runbookUrl",
					Enabled:                   false,
					ViolationTimeLimit:        "",
					ViolationTimeLimitSeconds: 2592000,
					Terms: []alerts.NrqlConditionTerm{
						{ThresholdDuration: 60, Operator: "ABOVE", Priority: "WARNING", Threshold: pointy.Float64(2.0), ThresholdOccurrences: "ALL"},
						{ThresholdDuration: 300, Operator: "ABOVE", Priority: "CRITICAL", Threshold: pointy.Float64(5.0), ThresholdOccurrences: "ALL"},
					},
					Signal: &alerts.AlertsNrqlConditionUpdateSignal{
						AggregationDelay:  pointy.Int(120),
						AggregationMethod: &alerts.NrqlConditionAggregationMethodTypes.EventFlow,
						AggregationWindow: pointy.Int(60),
						FillOption:        GetAlertsFillOptionPointer("NONE"),
					},
					Expiration: &alerts.AlertsNrqlConditionExpiration{
						ExpirationDuration:          nil,
						CloseViolationsOnExpiration: false,
						OpenViolationOnExpiration:   false,
					},
					Nrql: alerts.NrqlConditionUpdateQuery{Query: "SELECT latest(controller_runtime_reconcile_errors_total) FROM Metric"},
				},
			}},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			input := GenerateAlertConditionInput(&tc.args.cr)

			got, err := GenerateNrqlConditionUpdateInput(input)
			if err != nil {
				t.Errorf("e.TestIsUpToDate(...): -want, +got:\n%s\n", err)
			}
			if diff := cmp.Diff(tc.want.expected, got); diff != "" {
				t.Errorf("e.TestIsUpToDate(...): -want, +got:\n%s\n", diff)
			}
		})
	}
}
