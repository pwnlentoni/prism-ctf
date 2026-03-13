package reconcilers

import (
	"context"
	"fmt"
	"slices"

	v1 "github.com/pwnlentoni/prism-ctf/api/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// +kubebuilder:rbac:groups=prism-ctf.pwnlentoni.team,resources=prismserviceaccounts,verbs=get;list;watch
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles,verbs=bind,resourceNames=prism-ctf-challenge-manager-role
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create

func ReconcileRoleBinding(ctx context.Context, c client.Client, namespace string, commonLabels map[string]string, parent metav1.Object) error {
	l := log.FromContext(ctx)

	r := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "prism-ctf-challenge-manager-rolebinding",
			Namespace: namespace,
		},
	}
	op, err := controllerutil.CreateOrUpdate(ctx, c, r, func() error {
		if !controllerutil.HasControllerReference(r) {
			err := controllerutil.SetControllerReference(parent, r, c.Scheme())
			if err != nil {
				l.Error(err, "failed to set controller reference on rolebinding")
			}
		}
		r.Labels = commonLabels

		r.RoleRef.APIGroup = "rbac.authorization.k8s.io"
		r.RoleRef.Kind = "ClusterRole"
		r.RoleRef.Name = "prism-ctf-challenge-manager-role"

		sal := &v1.PrismServiceAccountList{}

		err := c.List(ctx, sal)
		if err != nil {
			return fmt.Errorf("failed to list prism service accounts: %w", err)
		}

		subjects := make([]rbacv1.Subject, 0, len(sal.Items))
		for _, sa := range sal.Items {
			if sa.Spec.Name == "" || sa.Spec.Namespace == "" {
				continue
			}
			subjects = append(subjects, rbacv1.Subject{
				Kind:      "ServiceAccount",
				Name:      sa.Spec.Name,
				Namespace: sa.Spec.Namespace,
			})
		}

		slices.SortFunc(subjects, func(a, b rbacv1.Subject) int {
			if a.Namespace != b.Namespace {
				if a.Namespace < b.Namespace {
					return -1
				}
				return 1
			}
			if a.Name != b.Name {
				if a.Name < b.Name {
					return -1
				}
				return 1
			}
			if a.Kind != b.Kind {
				if a.Kind < b.Kind {
					return -1
				}
				return 1
			}
			if a.APIGroup < b.APIGroup {
				return -1
			}
			if a.APIGroup > b.APIGroup {
				return 1
			}
			return 0
		})

		if !slices.Equal(r.Subjects, subjects) {
			r.Subjects = subjects
		}

		return nil
	})
	if err != nil {
		l.Error(err, "rolebinding reconcile error")
		return fmt.Errorf("rolebinding reconcile: %w", err)
	}
	l.Info("rolebinding reconciled", "operation", op)
	return nil

}
