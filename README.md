# provider-newrelic

`provider-newrelic` is a [Crossplane](https://crossplane.io/) Provider
that is meant to be used for infrastructure-as-code for New Relic.
It contains the following:

- A `ProviderConfig` type that only points to a credentials `Secret`.
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
- A `Secret` which contains a new relic user token
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
