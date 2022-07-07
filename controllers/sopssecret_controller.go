/*
Copyright The SOPS Operator Authors.

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

package controllers

import (
	"context"
	"fmt"
	"math"
	"time"
	"unicode"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	craftypathgithubiov1alpha1 "github.com/riskalyze/sops-operator/api/v1alpha1"
)

type Decryptor interface {
	Decrypt(fileName string, encrypted string) ([]byte, error)
}

// SopsSecretReconciler reconciles a SopsSecret object
type SopsSecretReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	Recorder  record.EventRecorder
	Decryptor Decryptor
}

//+kubebuilder:rbac:groups=craftypath.github.io,resources=sopssecrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=craftypath.github.io,resources=sopssecrets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=craftypath.github.io,resources=sopssecrets/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=coordination.k8s.io,resources=leases,verbs=get;list;create;update

func (r *SopsSecretReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reqLogger := log.FromContext(ctx)
	reqLogger.Info("reconciling SopsSecret")

	// Fetch the SopsSecret instance
	instance := &craftypathgithubiov1alpha1.SopsSecret{}
	if err := r.Get(ctx, req.NamespacedName, instance); err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
		},
	}

	result, err := ctrl.CreateOrUpdate(ctx, r.Client, secret, func() error {
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

func (r *SopsSecretReconciler) update(ctx context.Context, secret *corev1.Secret, sopsSecret *craftypathgithubiov1alpha1.SopsSecret) error {
	logger := log.FromContext(ctx)
	logger.Info("handling Secret update")

	data := make(map[string][]byte, len(sopsSecret.Spec.StringData))
	for fileName, encryptedContents := range sopsSecret.Spec.StringData {
		logger.Info("decrypting data", "fileName", fileName)
		decrypted, err := r.Decryptor.Decrypt(fileName, encryptedContents)
		if err != nil {
			return err
		}
		data[fileName] = decrypted
	}

	secret.Annotations = sopsSecret.Spec.Metadata.Annotations
	secret.Labels = sopsSecret.Spec.Metadata.Labels
	secret.Data = data
	if sopsSecret.Spec.Type != "" {
		secret.Type = sopsSecret.Spec.Type
	}

	logger.Info("setting controller reference")
	if err := ctrl.SetControllerReference(sopsSecret, secret, r.Scheme); err != nil {
		return fmt.Errorf("unable to set ownerReference: %w", err)
	}
	return nil
}

func (r *SopsSecretReconciler) manageError(ctx context.Context, instance *craftypathgithubiov1alpha1.SopsSecret, issue error) (reconcile.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("handling reconciliation error")

	r.Recorder.Event(instance, "Warning", "ProcessingError", capitalizeFirst(issue.Error()))

	lastUpdate := instance.Status.LastUpdate
	lastStatus := instance.Status.Status

	status := craftypathgithubiov1alpha1.SopsSecretStatus{
		LastUpdate: metav1.Now(),
		Reason:     issue.Error(),
		Status:     "Failure",
	}
	instance.Status = status

	if err := r.Status().Update(ctx, instance); err != nil {
		logger.Error(err, "unable to update status")
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
	logger.Error(issue, "failed to reconcile SopsSecret", "reqeueAfter", reqeueAfter)
	return reconcile.Result{
		RequeueAfter: reqeueAfter,
		Requeue:      true,
	}, nil
}

func (r *SopsSecretReconciler) manageSuccess(ctx context.Context, instance *craftypathgithubiov1alpha1.SopsSecret, result controllerutil.OperationResult) (reconcile.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("handling reconciliation success")

	if result == controllerutil.OperationResultNone {
		return reconcile.Result{}, nil
	}

	status := craftypathgithubiov1alpha1.SopsSecretStatus{
		LastUpdate: metav1.Now(),
		Reason:     "",
		Status:     "Success",
	}
	instance.Status = status

	if err := r.Status().Update(ctx, instance); err != nil {
		logger.Error(err, "unable to update status")
		r.Recorder.Event(instance, "Warning", "ProcessingError", "Unable to update status")
		return reconcile.Result{
			RequeueAfter: time.Second,
			Requeue:      true,
		}, nil
	}

	opResult := capitalizeFirst(string(result))
	msg := fmt.Sprintf("%s secret: %s", opResult, instance.Name)
	logger.Info("status updated successfully: " + msg)
	r.Recorder.Event(instance, "Normal", opResult, msg)
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

// SetupWithManager sets up the controller with the Manager.
func (r *SopsSecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&craftypathgithubiov1alpha1.SopsSecret{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}
