module github.com/craftypath/sops-operator

go 1.14

require (
	github.com/go-logr/logr v0.1.0
	github.com/golangci/golangci-lint v1.25.0
	github.com/goreleaser/goreleaser v0.132.1 // indirect
	github.com/magefile/mage v1.9.0
	github.com/operator-framework/operator-sdk v0.17.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.5.1
	go.mozilla.org/sops/v3 v3.5.0
	go.uber.org/zap v1.14.1
	golang.org/x/tools v0.0.0-20200422022333-3d57cf2e726e
	k8s.io/api v0.17.4
	k8s.io/apimachinery v0.17.4
	k8s.io/client-go v12.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.5.2
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM
	k8s.io/client-go => k8s.io/client-go v0.17.4 // Required by prometheus-operator
)
