module github.com/craftypath/sops-operator

go 1.15

require (
	github.com/go-logr/logr v0.1.0
	github.com/golangci/golangci-lint v1.31.0
	github.com/goreleaser/goreleaser v0.143.0
	github.com/magefile/mage v1.10.0
	github.com/operator-framework/operator-lib v0.1.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.6.1
	go.uber.org/zap v1.10.0
	golang.org/x/tools v0.0.0-20200812195022-5ae4c3c160a0
	k8s.io/api v0.18.6
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v0.18.6
	sigs.k8s.io/controller-runtime v0.6.2
)
