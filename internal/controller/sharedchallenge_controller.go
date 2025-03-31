/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	prismctfv1 "github.com/pwnlentoni/prism-ctf/api/v1"
	"github.com/pwnlentoni/prism-ctf/internal/utils"
	"github.com/pwnlentoni/prism-ctf/internal/utils/reconcilers"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"time"
)

// SharedChallengeReconciler reconciles a SharedChallenge object
type SharedChallengeReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=prism-ctf.pwnlentoni.team,resources=sharedchallenges,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=prism-ctf.pwnlentoni.team,resources=sharedchallenges/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=prism-ctf.pwnlentoni.team,resources=sharedchallenges/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=events,verbs=create;patch

func (r *SharedChallengeReconciler) internalReconcile(ctx context.Context, namespace string, commonLabels map[string]string, chal *prismctfv1.SharedChallenge) (error, string) {
	l := log.FromContext(ctx)

	err := reconcilers.ReconcileNamespace(ctx, r.Client, namespace, commonLabels, chal)
	if err != nil {
		l.Error(err, "namespace reconcile failed")
		return err, "NamespaceReconcileFailed"
	}

	err = reconcilers.ReconcileNetworkPolicies(ctx, r.Client, namespace, commonLabels, chal)
	if err != nil {
		l.Error(err, "network policy reconcile failed")
		return err, "NetworkPolicyReconcileFailed"
	}

	err = reconcilers.ReconcileConfigMap(ctx, r.Client, namespace, commonLabels, chal, chal.Spec.Flag)
	if err != nil {
		l.Error(err, "configmap reconcile failed")
		return err, "ConfigMapReconcileFailed"
	}

	statusMap, err := reconcilers.ReconcileContainers(ctx, r.Client, namespace, commonLabels, chal, chal.Spec.Containers, utils.NodeTypeShared, chal.Spec.Flag)
	if err != nil {
		l.Error(err, "containers reconcile failed")
		return err, "ContainersReconcileFailed"
	}

	chal.Status.ExposedUrls, err = reconcilers.ReconcileIngress(ctx, r.Client, namespace, commonLabels, chal, chal.Spec.Exposes, chal.Name, utils.DomainSuffix(), statusMap)
	if err != nil {
		l.Error(err, "ingress reconcile failed")
		return err, "IngressReconcileFailed"
	}
	return nil, "ReconcileSuccess"
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the SharedChallenge object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.1/pkg/reconcile
func (r *SharedChallengeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	l.Info("got reconcile request", "req", req)

	chal := &prismctfv1.SharedChallenge{}
	if err := r.Get(ctx, req.NamespacedName, chal); err != nil {
		err = client.IgnoreNotFound(err)
		if err == nil {
			l.Info("challenge spec deleted", "chall", req)
		} else {
			l.Error(err, "couldn't get challenge spec")
		}
		return ctrl.Result{}, err
	}
	r.Recorder.Event(chal, corev1.EventTypeNormal, "ReconcileStart", "Challenge reconciliation started")

	namespace := utils.SharedChallengeNamespace(chal.GetName())

	commonLabels := utils.MakeCommonLabels(chal.GetName())

	err, condition := r.internalReconcile(ctx, namespace, commonLabels, chal)
	chal.Status.Conditions = []metav1.Condition{
		{
			Type:               "Ready",
			Status:             metav1.ConditionTrue,
			ObservedGeneration: chal.ObjectMeta.Generation,
			Reason:             condition,
			LastTransitionTime: metav1.NewTime(time.Now()),
		},
	}
	if err != nil {
		chal.Status.Conditions[0].Status = metav1.ConditionFalse
		chal.Status.Conditions[0].Message = err.Error()
		r.Recorder.Event(chal, corev1.EventTypeWarning, "ReconcileFailed", err.Error())
	} else {
		r.Recorder.Event(chal, corev1.EventTypeNormal, "Reconciled", "Reconciled successfully")
	}
	updErr := r.Status().Update(ctx, chal)
	if updErr != nil {
		l.Error(updErr, "failed to update status")
	}

	return ctrl.Result{}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *SharedChallengeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&prismctfv1.SharedChallenge{}).
		Owns(&appsv1.Deployment{}).
		Named("sharedchallenge").
		Complete(r)
}
