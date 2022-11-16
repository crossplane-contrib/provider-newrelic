# newrelic_configuration

## Currently supported NewRelic Configuration-as-code
* Policies
* Nrql Conditions
* Dashboards

## Tips on generating Policies and Nrql Conditions

* Create, test, and save an alert condition in the New Relic UI
    * https://docs.newrelic.com/docs/alerts-applied-intelligence/new-relic-alerts/alert-conditions/create-nrql-alert-conditions/
* Record the ID of the condition

* Now that you have a condition, the next step is to export it so it's properties can be copied into a configuration file


* Navigate to the graphql explorer
    * https://api.newrelic.com/graphiql

* You may use the explorer to check off the properties you want the API to return, or you can paste the query below into the search box
    * Be sure to replace the conditionId with the ID from the condition you created in the NewRelic UI

* Export your condition to json


* Example API Explorer Query
```
{
  actor {
    account(id: 111111) {
      alerts {
        nrqlCondition(id: "11111111") {
          description
          enabled
          id
          nrql {
            query
          }
          name
          policyId
          runbookUrl
          signal {
            aggregationDelay
            aggregationMethod
            aggregationTimer
            aggregationWindow
            fillOption
            fillValue
          }
          terms {
            operator
            priority
            threshold
            thresholdDuration
            thresholdOccurrences
          }
          type
          violationTimeLimitSeconds
          expiration {
            closeViolationsOnExpiration
            expirationDuration
            openViolationOnExpiration
          }
          ... on AlertsNrqlStaticCondition {
            valueFunction
          }
        }
      }
    }
  }
}
```


* Convert to yaml format (can use the script below)
    * `python3 -c 'import yaml, sys; print(yaml.dump(yaml.load(open(sys.argv[1])), default_flow_style=False))' conditions.json`


* The output can be used as the base for your manifest or helm chart

## Tips on schema
* All nrql conditions must match the `apiVersion` and `kind` defined below
```yaml
---
apiVersion: nrqlalertcondition.provider-newrelic.crossplane.io/v1alpha1
kind: NrqlAlertCondition
metadata:
  name: {{ $env_name }}-your-unique-name
  {{- if $.Values.annotations }}
  annotations:  {{ toYaml $.Values.annotations | nindent 4 }}
  {{- end }}
```

* The rest of the schema (everything under `forProvider`) should be a 1:1 match to the NewRelic API spec for a nrql condition.  Exceptions are noted below
```yaml
spec:
  forProvider:
```

* References a policy by name instead of needing to hard-code a policy ID
```yaml
  forProvider:
    alertsPolicyRef:
      name: "{{ $env_name }}-k8s-policy"
```

* This is a reference to the provider/pod which is called to make the API request to NewRelic.  Unless noted, it should always be set to `newrelic-provider`
```yaml
  providerConfigRef:
    name: {{ $.Values.providerConfigRef.name }}
```

## Post Deployment Verification
* Verify your objects sync'ed properly 
```text
kubectl -n crossplane-system get NrqlAlertCondition -l app.cloudcheckr.com/appname=cc-provider-newrelic-rds-starvation-alarms-prod-corp
NAME                           ID         READY   SYNCED   AGE
dev-rds-server-high-cpu        10000001   True    True     13m
qa-rds-server-high-cpu         10000002   True    True     13m
prod-us-rds-server-high-cpu    10000003   True    True     13m
prod-au-rds-server-high-cpu    10000004   True    True     13m
prod-eu-rds-server-high-cpu    10000005   True    True     13m
```

* Each resource has a status with more detailed info
    * The corresponding NewRelic ID
    * When the object was last sync'ed
    * Whether it was created and sync'ed successfully
    * Any errors returned by the NewRelic API
```yaml
status:
  atProvider:
    id: '1234567'
  conditions:
    - lastTransitionTime: '2022-11-15T01:34:47Z'
      reason: ReconcileSuccess
      status: 'True'
      type: Synced
    - lastTransitionTime: '2022-11-15T01:34:51Z'
      reason: Available
      status: 'True'
      type: Ready
```
