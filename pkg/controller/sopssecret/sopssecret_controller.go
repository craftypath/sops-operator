// Copyright The SOPS Operator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sopssecret

import (
	"context"
	"encoding/base64"
	"fmt"
	"math"
	"time"
	"unicode"

	"github.com/go-logr/logr"

	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/craftypath/sops-operator/pkg/sops"

	craftypathv1alpha1 "github.com/craftypath/sops-operator/pkg/apis/craftypath/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type loggerKey string

const (
	reqLoggerKey   loggerKey = "logger"
	controllerName string    = "sopssecret-controller"
)

var log = logf.Log.WithName(controllerName)

type Decryptor interface {
	Decrypt(fileName string, encrypted string) ([]byte, error)
}

// Add creates a new SopsSecret Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileSopsSecret{
		client:    mgr.GetClient(),
		scheme:    mgr.GetScheme(),
		recorder:  mgr.GetEventRecorderFor(controllerName),
		decryptor: &sops.Decryptor{},
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New(controllerName, mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource SopsSecret
	if err = c.Watch(&source.Kind{Type: &craftypathv1alpha1.SopsSecret{}}, &handler.EnqueueRequestForObject{}); err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileSopsSecret implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileSopsSecret{}

// ReconcileSopsSecret reconciles a SopsSecret object
type ReconcileSopsSecret struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client    client.Client
	scheme    *runtime.Scheme
	recorder  record.EventRecorder
	decryptor Decryptor
}

// Reconcile reads that state of the cluster for a SopsSecret object and makes changes based on the state read
// and what is in the SopsSecret.Spec
func (r *ReconcileSopsSecret) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	ctx := context.WithValue(context.Background(), reqLoggerKey, reqLogger)

	reqLogger.Info("reconciling SopsSecret")

	// Fetch the SopsSecret instance
	instance := &craftypathv1alpha1.SopsSecret{}
	if err := r.client.Get(ctx, request.NamespacedName, instance); err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
		},
	}

	result, err := ctrl.CreateOrUpdate(ctx, r.client, secret, func() error {
		if !secret.CreationTimestamp.IsZero() {
			if !metav1.IsControlledBy(secret, instance) {
				return fmt.Errorf("secret already exists and not owned by sops-operator")
			}
		}
		if err := r.update(ctx, secret, instance); err != nil {
			return fmt.Errorf("failed to update secret: %w", err)
		}
		return nil
	})
	if err != nil {
		return r.manageError(ctx, instance, err)
	}

	return r.manageSuccess(ctx, instance, result)
}

func (r *ReconcileSopsSecret) update(ctx context.Context, secret *corev1.Secret, sopsSecret *craftypathv1alpha1.SopsSecret) error {
	logger(ctx).V(1).Info("handling Secret update")

	data := make(map[string][]byte, len(sopsSecret.Spec.StringData))
	for fileName, encryptedContents := range sopsSecret.Spec.StringData {
		logger(ctx).V(1).Info("decrypting data", "fileName", fileName)
		decrypted, err := r.decryptor.Decrypt(fileName, encryptedContents)
		if err != nil {
			return err
		}
		logger(ctx).V(1).Info("base64-encoding data", "fileName", fileName)
		buf := make([]byte, base64.StdEncoding.EncodedLen(len(decrypted)))
		base64.StdEncoding.Encode(buf, decrypted)
		data[fileName] = buf
	}

	secret.ObjectMeta.Annotations = sopsSecret.Annotations
	secret.ObjectMeta.Labels = sopsSecret.Labels
	secret.Data = data
	if sopsSecret.Spec.Type != "" {
		secret.Type = sopsSecret.Spec.Type
	}

	logger(ctx).V(1).Info("setting controller reference")
	if err := ctrl.SetControllerReference(sopsSecret, secret, r.scheme); err != nil {
		return fmt.Errorf("unable to set ownerReference: %w", err)
	}
	return nil
}

func (r *ReconcileSopsSecret) manageError(ctx context.Context, instance *craftypathv1alpha1.SopsSecret, issue error) (reconcile.Result, error) {
	logger(ctx).V(1).Info("handling reconciliation error")

	r.recorder.Event(instance, "Warning", "ProcessingError", capitalizeFirst(issue.Error()))

	lastUpdate := instance.Status.LastUpdate
	lastStatus := instance.Status.Status

	status := craftypathv1alpha1.SopsSecretStatus{
		LastUpdate: metav1.Now(),
		Reason:     issue.Error(),
		Status:     "Failure",
	}
	instance.Status = status

	if err := r.client.Status().Update(ctx, instance); err != nil {
		log.Error(err, "unable to update status")
		return reconcile.Result{
			RequeueAfter: time.Second,
			Requeue:      true,
		}, nil
	}

	var retryInterval time.Duration
	if lastUpdate.IsZero() || lastStatus == "Success" {
		retryInterval = time.Second
	} else {
		retryInterval = status.LastUpdate.Sub(lastUpdate.Time.Round(time.Second))
	}

	reqeueAfter := time.Duration(math.Min(float64(retryInterval.Nanoseconds()*2), float64(time.Hour.Nanoseconds()*6)))
	logger(ctx).Error(issue, "failed to reconcile SopsSecret", "reqeueAfter", reqeueAfter)
	return reconcile.Result{
		RequeueAfter: reqeueAfter,
		Requeue:      true,
	}, nil
}

func (r *ReconcileSopsSecret) manageSuccess(ctx context.Context, instance *craftypathv1alpha1.SopsSecret, result controllerutil.OperationResult) (reconcile.Result, error) {
	logger(ctx).V(1).Info("handling reconciliation success")

	status := craftypathv1alpha1.SopsSecretStatus{
		LastUpdate: metav1.Now(),
		Reason:     "",
		Status:     "Success",
	}
	instance.Status = status

	if err := r.client.Status().Update(ctx, instance); err != nil {
		logger(ctx).Error(err, "unable to update status")
		r.recorder.Event(instance, "Warning", "ProcessingError", "Unable to update status")
		return reconcile.Result{
			RequeueAfter: time.Second,
			Requeue:      true,
		}, nil
	}

	opResult := capitalizeFirst(string(result))
	msg := fmt.Sprintf("%s secret: %s", opResult, instance.Name)
	logger(ctx).Info("status updated successfully: " + msg)
	r.recorder.Event(instance, "Normal", opResult, msg)
	return reconcile.Result{}, nil
}

func capitalizeFirst(s string) string {
	if len(s) == 0 {
		return ""
	}
	tmp := []rune(s)
	tmp[0] = unicode.ToUpper(tmp[0])
	return string(tmp)
}

func logger(ctx context.Context) logr.Logger {
	return ctx.Value(reqLoggerKey).(logr.Logger)
}
