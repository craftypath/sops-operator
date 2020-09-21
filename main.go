/*
Copyright The SOPS Operator Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"fmt"
	"os"
	rt "runtime"
	"strings"
	"time"

	"github.com/craftypath/sops-operator/pkg/sops"
	"github.com/craftypath/sops-operator/pkg/version"
	"github.com/spf13/pflag"
	"sigs.k8s.io/controller-runtime/pkg/cache"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	craftypathv1alpha1 "github.com/craftypath/sops-operator/api/v1alpha1"
	"github.com/craftypath/sops-operator/controllers"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

const controllerName string = "sopssecret-controller"

type leaderElectionFlags struct {
	leaderElection          bool
	leaderElectionID        string
	leaderElectionNamespace string
	leaseDuration           time.Duration
	renewDeadline           time.Duration
	retryPeriod             time.Duration
}

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(craftypathv1alpha1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")

	opts := zap.Options{}
	opts.BindFlags(flag.CommandLine)

	// Add flags registered by imported packages (e.g. glog and
	// controller-runtime)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	lef := leaderElectionFlags{}
	pflag.CommandLine.AddFlagSet(lef.flagSet())

	pflag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	printVersion()

	watchNamespace := os.Getenv("WATCH_NAMESPACE")

	options := ctrl.Options{
		Scheme:                  scheme,
		MetricsBindAddress:      metricsAddr,
		Port:                    9443,
		Namespace:               watchNamespace,
		LeaderElection:          lef.leaderElection,
		LeaderElectionID:        lef.leaderElectionID,
		LeaderElectionNamespace: lef.leaderElectionNamespace,
		LeaseDuration:           &lef.leaseDuration,
		RenewDeadline:           &lef.renewDeadline,
		RetryPeriod:             &lef.retryPeriod,
	}

	// Add support for MultiNamespace set in WATCH_NAMESPACE (e.g ns1,ns2)
	if strings.Contains(watchNamespace, ",") {
		setupLog.Info("manager will be watching multiple namespace", "namespaces", watchNamespace)
		// configure cluster-scoped with MultiNamespacedCacheBuilder
		options.Namespace = ""
		options.NewCache = cache.MultiNamespacedCacheBuilder(strings.Split(watchNamespace, ","))
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), options)
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.SopsSecretReconciler{
		Client:    mgr.GetClient(),
		Log:       ctrl.Log.WithName("controllers").WithName("SopsSecret"),
		Scheme:    mgr.GetScheme(),
		Recorder:  mgr.GetEventRecorderFor(controllerName),
		Decryptor: &sops.Decryptor{},
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "SopsSecret")
		os.Exit(1)
	}

	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func (l *leaderElectionFlags) flagSet() *pflag.FlagSet {
	leaderElectionFlags := pflag.NewFlagSet("leader-election", pflag.ExitOnError)
	leaderElectionFlags.StringVar(&l.leaderElectionID, "leader-election-id", "sops-operator-lock", "The name of the configmap that leader election will use for holding the leader lock")
	leaderElectionFlags.BoolVar(&l.leaderElection, "leader-election", true, "Enable leader election")
	leaderElectionFlags.StringVar(&l.leaderElectionNamespace, "leader-election-namespace", "", "The namespace in which the leader election configmap will be created")
	leaderElectionFlags.DurationVar(&l.leaseDuration, "lease-duration", 15*time.Second, "The duration that non-leader candidates will wait to force acquire leadership")
	leaderElectionFlags.DurationVar(&l.renewDeadline, "renew-deadline", 10*time.Second, "The duration that the acting master will retry refreshing leadership before giving up")
	leaderElectionFlags.DurationVar(&l.retryPeriod, "retry-duration", 2*time.Second, "Te duration the LeaderElector clients should wait between tries of actions")
	return leaderElectionFlags
}

func printVersion() {
	setupLog.Info(fmt.Sprintf("Operator Version: %s", version.Version))
	setupLog.Info(fmt.Sprintf("Git Commit: %s", version.GitCommit))
	setupLog.Info(fmt.Sprintf("Build Date: %s", version.BuildDate))
	setupLog.Info(fmt.Sprintf("Go Version: %s", rt.Version()))
	setupLog.Info(fmt.Sprintf("Go OS/Arch: %s/%s", rt.GOOS, rt.GOARCH))
}
