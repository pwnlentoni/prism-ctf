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
	"fmt"
	"github.com/pwnlentoni/prism-ctf/internal/utils"
	"github.com/pwnlentoni/prism-ctf/internal/utils/reconcilers"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	prismctfv1 "github.com/pwnlentoni/prism-ctf/api/v1"
)

// ChallengeInstanceReconciler reconciles a ChallengeInstance object
type ChallengeInstanceReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

func (r *ChallengeInstanceReconciler) internalReconcile(ctx context.Context, namespace string, commonLabels map[string]string, instance *prismctfv1.ChallengeInstance, chal *prismctfv1.IsolatedChallenge) (error, string) {
	l := log.FromContext(ctx)

	err := reconcilers.ReconcileNamespace(ctx, r.Client, namespace, commonLabels, instance)
	if err != nil {
		l.Error(err, "namespace reconcile failed")
		return err, "NamespaceReconcileFailed"
	}

	err = reconcilers.ReconcileNetworkPolicies(ctx, r.Client, namespace, commonLabels, instance)
	if err != nil {
		l.Error(err, "network policy reconcile failed")
		return err, "NetworkPolicyReconcileFailed"
	}

	err = reconcilers.ReconcileConfigMap(ctx, r.Client, namespace, commonLabels, instance, &instance.Spec.Flag)
	if err != nil {
		l.Error(err, "configmap reconcile failed")
		return err, "ConfigMapReconcileFailed"
	}

	statusMap, err := reconcilers.ReconcileContainers(ctx, r.Client, namespace, commonLabels, instance, chal.Spec.Containers, utils.NodeTypeIsolated, &instance.Spec.Flag)
	if err != nil {
		l.Error(err, "containers reconcile failed")
		return err, "ContainersReconcileFailed"
	}

	instance.Status.ExposedUrls, err = reconcilers.ReconcileIngress(ctx, r.Client, namespace, commonLabels, instance, chal.Spec.Exposes, chal.Name, "-"+instance.Spec.RandomId+utils.DomainSuffix(), statusMap)
	if err != nil {
		l.Error(err, "ingress reconcile failed")
		return err, "IngressReconcileFailed"
	}
	return nil, "ReconcileSuccess"
}

// +kubebuilder:rbac:groups=prism-ctf.pwnlentoni.team,resources=isolatedchallenges,verbs=get;list;watch
// +kubebuilder:rbac:groups=prism-ctf.pwnlentoni.team,resources=challengeinstances,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=prism-ctf.pwnlentoni.team,resources=challengeinstances/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=prism-ctf.pwnlentoni.team,resources=challengeinstances/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ChallengeInstance object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.4/pkg/reconcile
func (r *ChallengeInstanceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	l.Info("got reconcile request", "req", req)

	instance := &prismctfv1.ChallengeInstance{}
	if err := r.Get(ctx, req.NamespacedName, instance); err != nil {
		err = client.IgnoreNotFound(err)
		if err == nil {
			l.Info("instance deleted", "req", req)
		} else {
			l.Error(err, "couldn't get instance", "req", req)
		}
		return ctrl.Result{}, err
	}

	if time.Now().After(instance.Spec.Expiration.Time) {
		l.Info("instance expired, deleting", "req", req)
		err := r.Delete(ctx, instance)
		if err != nil {
			l.Error(err, "failed to delete expired instance", "req", req)
		}
		return ctrl.Result{}, err
	}

	chal := &prismctfv1.IsolatedChallenge{}
	err := r.Client.Get(ctx, client.ObjectKey{Name: instance.Spec.Challenge}, chal)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get challenge: %w", err)
	}

	r.Recorder.Event(instance, corev1.EventTypeNormal, "ReconcileStart", "Challenge reconciliation started")

	namespace := utils.IsolatedChallengeNamespace(chal.GetName(), instance.Spec.Team, instance.Spec.RandomId)

	commonLabels := utils.MapMerge(utils.MakeCommonLabels(chal.GetName()), utils.MakeInstancedLabels(instance.Spec.Team))

	err, condition := r.internalReconcile(ctx, namespace, commonLabels, instance, chal)
	instance.Status.Conditions = []metav1.Condition{
		{
			Type:               "Ready",
			Status:             metav1.ConditionTrue,
			ObservedGeneration: instance.ObjectMeta.Generation,
			Reason:             condition,
			LastTransitionTime: metav1.NewTime(time.Now()),
		},
	}
	if err != nil {
		instance.Status.Conditions[0].Status = metav1.ConditionFalse
		instance.Status.Conditions[0].Message = err.Error()
		r.Recorder.Event(instance, corev1.EventTypeWarning, "ReconcileFailed", err.Error())
	} else {
		r.Recorder.Event(instance, corev1.EventTypeNormal, "Reconciled", "Reconciled successfully")
	}
	updErr := r.Status().Update(ctx, instance)
	if updErr != nil {
		l.Error(updErr, "failed to update status")
	}

	return ctrl.Result{
		RequeueAfter: instance.Spec.Expiration.Sub(time.Now()) + 1*time.Second,
	}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ChallengeInstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&prismctfv1.ChallengeInstance{}).
		Owns(&appsv1.Deployment{}).
		Named("challengeinstance").
		Complete(r)
}
