module github.com/crossplane-contrib/provider-newrelic

go 1.16

require (
	github.com/crossplane/crossplane-runtime v0.15.1-0.20210930095326-d5661210733b
	github.com/crossplane/crossplane-tools v0.0.0-20210916125540-071de511ae8e
	github.com/google/go-cmp v0.5.6
	github.com/newrelic/newrelic-client-go v0.68.3
	github.com/onsi/gomega v1.15.0 // indirect
	github.com/openlyinc/pointy v1.1.2
	github.com/pkg/errors v0.9.1
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	k8s.io/apimachinery v0.21.3
	k8s.io/client-go v0.21.3
	sigs.k8s.io/controller-runtime v0.9.6
	sigs.k8s.io/controller-tools v0.6.2
)
