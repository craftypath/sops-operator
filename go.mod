module github.com/craftypath/sops-operator

go 1.14

require (
	github.com/go-logr/logr v0.1.0
	github.com/golangci/golangci-lint v1.27.0
	github.com/goreleaser/goreleaser v0.138.0
	github.com/magefile/mage v1.9.0
	github.com/operator-framework/operator-sdk v0.18.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.6.1
	go.mozilla.org/sops/v3 v3.5.0
	go.uber.org/zap v1.14.1
	golang.org/x/tools v0.0.0-20200608174601-1b747fd94509
	k8s.io/api v0.18.2
	k8s.io/apimachinery v0.18.2
	k8s.io/client-go v12.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.6.0
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM
	k8s.io/client-go => k8s.io/client-go v0.18.2 // Required by prometheus-operator
)
