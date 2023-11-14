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
	"testing"

	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/google/go-cmp/cmp"
	"github.com/newrelic/newrelic-client-go/v2/pkg/alerts"

	"github.com/crossplane-contrib/provider-newrelic/apis/alertspolicy/v1alpha1"
)

// Unlike many Kubernetes projects Crossplane does not use third party testing
// libraries, per the common Go test review comments. Crossplane encourages the
// use of table driven unit tests. The tests of the crossplane-runtime project
// are representative of the testing style Crossplane encourages.
//
// https://github.com/golang/go/wiki/TestComments
// https://github.com/crossplane/crossplane/blob/master/CONTRIBUTING.md#contributing-code

type alertPolicyModifier func(*v1alpha1.AlertsPolicy)

func alertPolicy(m ...alertPolicyModifier) *v1alpha1.AlertsPolicy {
	cr := &v1alpha1.AlertsPolicy{
		Spec: v1alpha1.AlertsPolicySpec{
			ForProvider: v1alpha1.AlertsPolicyParameters{
				ID:                 "1",
				IncidentPreference: string(alerts.IncidentPreferenceTypes.PerCondition),
				Name:               "test_name",
			},
		},
	}
	meta.SetExternalName(cr, "1")
	for _, f := range m {
		f(cr)
	}
	return cr
}

func TestIsUpToDate(t *testing.T) {

	type args struct {
		cr v1alpha1.AlertsPolicy
		nr alerts.AlertsPolicy
	}

	type want struct {
		expected bool
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"DiffIncidentPreference": {
			args: args{cr: *alertPolicy(),
				nr: alerts.AlertsPolicy{
					Name:               "test_name",
					IncidentPreference: alerts.AlertsIncidentPreference(alerts.IncidentPreferenceTypes.PerPolicy),
				},
			},
			want: want{expected: false},
		},
		"DiffName": {
			args: args{cr: *alertPolicy(),
				nr: alerts.AlertsPolicy{
					Name:               "test_name_diff",
					IncidentPreference: alerts.AlertsIncidentPreference(alerts.IncidentPreferenceTypes.PerCondition),
				},
			},
			want: want{expected: false},
		},
		"Same": {
			args: args{cr: *alertPolicy(),
				nr: alerts.AlertsPolicy{
					Name:               "test_name",
					IncidentPreference: alerts.AlertsIncidentPreference(alerts.IncidentPreferenceTypes.PerCondition),
				},
			},
			want: want{expected: true},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := IsUpToDate(&tc.args.cr, tc.args.nr)
			if diff := cmp.Diff(tc.want.expected, got); diff != "" {
				t.Errorf("e.TestIsUpToDate(...): -want, +got:\n%s\n", diff)
			}
		})
	}
}
