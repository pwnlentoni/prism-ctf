package reconcilers

import (
	"context"
	"fmt"
	"google.golang.org/protobuf/proto"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func ReconcileConfigMap(ctx context.Context, c client.Client, namespace string, commonLabels map[string]string, parent metav1.Object, flag *string) error {
	l := log.FromContext(ctx)

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "flag",
		},
	}
	if flag != nil {
		op, err := controllerutil.CreateOrUpdate(ctx, c, cm, func() error {
			if !controllerutil.HasControllerReference(cm) {
				err := controllerutil.SetControllerReference(parent, cm, c.Scheme())
				if err != nil {
					l.Error(err, "failed to set controller reference on configmap")
				}
			}

			cm.Labels = commonLabels

			if cm.Immutable == nil || !*cm.Immutable {
				cm.Immutable = proto.Bool(true)
			}
			if cm.Data == nil {
				cm.Data = map[string]string{}
			}
			if _, found := cm.Data["flag"]; !found {
				cm.Data["flag"] = *flag
			}
			if _, found := cm.Data["flag_lf"]; !found {
				cm.Data["flag_lf"] = *flag + "\n"
			}
			return nil
		})
		if err != nil {
			l.Error(err, "configmap reconcile error")
			return fmt.Errorf("configmap reconcile: %w", err)
		}
		l.Info("configmap reconciled", "operation", op)
	} else {
		err := client.IgnoreNotFound(c.Delete(ctx, cm))
		if err != nil {
			l.Error(err, "cm delete error")
			return err
		}
	}
	return nil
}
