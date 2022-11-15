# provider-newrelic

`provider-newrelic` is a [Crossplane](https://crossplane.io/) Provider
that is meant to be used for infrastructure-as-code for New Relic.

See the examples directory for advanced usage.

This provider supports the following:

- `AlertsPolicy` - https://docs.newrelic.com/docs/alerts-applied-intelligence/new-relic-alerts/alert-policies/create-edit-or-find-alert-policy/
- `NrqlAlertCondition` - https://docs.newrelic.com/docs/alerts-applied-intelligence/new-relic-alerts/alert-conditions/create-nrql-alert-conditions/
- `Dashboard` - https://docs.newrelic.com/docs/query-your-data/explore-query-data/dashboards/introduction-dashboards/

- `ProviderConfig` type that only points to a credentials `Secret`.
```---
apiVersion: provider-newrelic.crossplane.io/v1alpha1
kind: ProviderConfig
metadata:
  name: newrelic-provider
spec:
  account_id: "your_nr_account_id"
  credentials:
    source: Secret
    secretRef:
      namespace: crossplane-system
      name: newrelic-creds
      key: key
```
- `Secret` which contains a new relic user token
```
---
apiVersion: v1
data:
  key: NRAK-YOUR_NEW_RELIC_TOKEN_BASE64
kind: Secret
metadata:
  name: newrelic-creds
  namespace: crossplane-system
type: Opaque
```

## Additional Note
Sometimes an `AlertsPolicy` may be deleted, or regenerated, giving it a new ID.  
This can cause issues for any `NrqlAlertCondition` with a reference to that object resulting in errors such as `"error": "Policy with ID 1234567 not found"`
To fix you can simply remove the `policyId` on the `NrqlAlertCondition` to to cause the reference to re-resolve.
(There is no harm in doing this, it will just cause the provider to lookup the new ID.)
```
kubectl -n crossplane-system patch NrqlAlertCondition my-condition-name --type json  --patch='[ { "op": "remove", "path": "/spec/forProvider/policyId" } ]'
```

## Developing

Run against a Kubernetes cluster:

```console
make run
```

Build, push, and install:

```console
make all
```

Build image:

```console
make image
```

Push image:

```console
make push
```

Build binary:

```console
make build
```
