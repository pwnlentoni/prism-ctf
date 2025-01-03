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
	"github.com/pwnlentoni/prism-ctf/internal/utils"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	prismctfv1 "github.com/pwnlentoni/prism-ctf/api/v1"
)

const finalizerName = "shared-challenges.prism-ctf.pwnlentoni.team/finalizer"

// SharedChallengeReconciler reconciles a SharedChallenge object
type SharedChallengeReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=prism-ctf.pwnlentoni.team,resources=sharedchallenges,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=prism-ctf.pwnlentoni.team,resources=sharedchallenges/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=prism-ctf.pwnlentoni.team,resources=sharedchallenges/finalizers,verbs=update

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
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if chal.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(chal, finalizerName) {
			controllerutil.AddFinalizer(chal, finalizerName)
			if err := r.Update(ctx, chal); err != nil {
				l.Error(err, "finalizer add error")
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
	} else {
		l.Info("challenge spec deleted", "chall", chal.GetName())
		if controllerutil.ContainsFinalizer(chal, finalizerName) {
			if err := r.cleanupChallenge(ctx, chal); err != nil {
				l.Error(err, "challenge cleanup failed")
				return ctrl.Result{}, err
			}

			controllerutil.RemoveFinalizer(chal, finalizerName)
			if err := r.Update(ctx, chal); err != nil {
				l.Error(err, "finalizer remove failed")
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	l.Info("challenge spec created/updated", "chall", chal.GetName())

	doc, err := utils.RenderSharedTemplate(chal.Spec.Template, "challs.pwnlentoni.team")
	if err != nil {
		return ctrl.Result{}, err
	}

	objs, err := utils.GetObjectsFromTemplate(r.Client, ctx, doc)
	if err != nil {
		return ctrl.Result{}, err
	}

	for _, obj := range objs {
		err = r.Client.Create(ctx, &obj)
		if err != nil {
			if client.IgnoreAlreadyExists(err) != nil {
				l.Error(err, "obj create error", "obj", obj)
				return ctrl.Result{}, err
			} else {
				gotObj := unstructured.Unstructured{Object: make(map[string]interface{})}
				gotObj.SetGroupVersionKind(obj.GroupVersionKind())
				err = r.Client.Get(ctx, client.ObjectKeyFromObject(&obj), &gotObj)
				if err != nil {
					l.Error(err, "obj update get error", "obj", obj)
					return ctrl.Result{}, err
				}
				obj.SetResourceVersion(gotObj.GetResourceVersion())
				err = r.Client.Update(ctx, &obj)
				if err != nil {
					l.Error(err, "obj update error", "obj", obj)
					return ctrl.Result{}, err
				}
			}
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SharedChallengeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&prismctfv1.SharedChallenge{}).
		Named("sharedchallenge").
		Complete(r)
}

func (r *SharedChallengeReconciler) cleanupChallenge(ctx context.Context, chal *prismctfv1.SharedChallenge) error {
	l := log.FromContext(ctx)
	l.Info("cleaning up challenge", "chall", chal.GetName())
	doc, err := utils.RenderSharedTemplate(chal.Spec.Template, "challs.pwnlentoni.team")
	if err != nil {
		return err
	}

	objs, err := utils.GetObjectsFromTemplate(r.Client, ctx, doc)
	if err != nil {
		return err
	}

	for _, obj := range objs {
		err = r.Client.Delete(ctx, &obj)
		if err != nil {
			if client.IgnoreNotFound(err) == nil {
				l.Info("deleting not found object", "name", obj.GetNamespace()+"/"+obj.GetName(), "kind", obj.GetKind())
			} else {
				l.Error(err, "object delete error", "name", obj.GetNamespace()+"/"+obj.GetName(), "kind", obj.GetKind())
				return err
			}
		}
	}

	return nil
}
