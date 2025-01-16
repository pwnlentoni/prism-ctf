package reconcilers

import (
	"context"
	"fmt"
	"github.com/pwnlentoni/prism-ctf/internal/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func ReconcileNamespace(ctx context.Context, c client.Client, namespace string, commonLabels map[string]string, parent metav1.Object) error {
	l := log.FromContext(ctx)
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{
		Name: namespace,
	}}

	op, err := controllerutil.CreateOrUpdate(ctx, c, ns, func() error {
		if !controllerutil.HasControllerReference(ns) {
			err := controllerutil.SetControllerReference(parent, ns, c.Scheme())
			if err != nil {
				l.Error(err, "failed to set controller reference on namespace")
				return fmt.Errorf("set controller reference: %w", err)
			}
		}
		ns.Labels = utils.MapMerge(utils.MapMerge(map[string]string{
			utils.ChallengeNamespaceLabel: utils.ChallengeNamespaceValue,
		}, commonLabels), ns.Labels)
		return nil
	})
	if err != nil {
		return fmt.Errorf("namespace `%s` reconcile error: %w", ns.Name, err)
	}
	l.Info("namespace reconciled", "operation", op, "name", ns.Name)

	return err
}
