module github.com/craftypath/sops-operator

go 1.16

require (
	github.com/golangci/golangci-lint v1.41.1
	github.com/goreleaser/goreleaser v0.183.0
	github.com/magefile/mage v1.11.0
	github.com/stretchr/testify v1.7.0
	github.com/sykesm/zap-logfmt v0.0.4
	go.uber.org/zap v1.19.1
	golang.org/x/tools v0.1.7
	k8s.io/api v0.21.2
	k8s.io/apimachinery v0.21.2
	k8s.io/client-go v0.21.2
	sigs.k8s.io/controller-runtime v0.9.2
	sigs.k8s.io/controller-tools v0.6.1
)
