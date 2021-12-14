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
	"context"

	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/reference"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	v1alpha "github.com/crossplane-contrib/provider-newrelic/apis/alertspolicy/v1alpha1"
)

// ResolveReferences of this AlertNrqlCondition
func (mg *NrqlAlertCondition) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.AlertNrqlCondition
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: mg.Spec.ForProvider.AlertsPolicyID,
		Reference:    mg.Spec.ForProvider.AlertsPolicyRef,
		Selector:     mg.Spec.ForProvider.AlertsPolicySelector,
		To:           reference.To{Managed: &v1alpha.AlertsPolicy{}, List: &v1alpha.AlertsPolicyList{}},
		Extract:      AlertPolicyID(),
	})

	if err != nil {
		return errors.Wrap(err, "Spec.ForProvider.AlertPolicyID")
	}

	if rsp.ResolvedValue == "" {
		return errors.New("Spec.ForProvider.AlertPolicyID not yet resolvable")
	}

	mg.Spec.ForProvider.AlertsPolicyID = rsp.ResolvedValue
	mg.Spec.ForProvider.AlertsPolicyRef = rsp.ResolvedReference

	return nil
}

// AlertPolicyID extracts info from a kubernetes referenced object
func AlertPolicyID() reference.ExtractValueFn {
	return func(mg resource.Managed) string {
		cr, _ := mg.(*v1alpha.AlertsPolicy)
		return cr.Spec.ForProvider.ID
	}
}
