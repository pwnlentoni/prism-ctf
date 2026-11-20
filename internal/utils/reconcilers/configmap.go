package reconcilers

import (
	"context"
	"fmt"
	"maps"
	"slices"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	// flagConfigMapName is the name of the per-challenge flag configmap (and of the volume mounting it)
	flagConfigMapName = "flag"
	// flagKey holds the flag; it is also the file name the flag is mounted as
	flagKey = "flag"
	// flagLFKey holds the flag followed by a newline
	flagLFKey = "flag_lf"
)

func ReconcileConfigMap(ctx context.Context, c client.Client, namespace string, commonLabels map[string]string, parent metav1.Object, flag *string, exposeMap map[string]string) error {
	l := log.FromContext(ctx)
	{
		cm := &corev1.ConfigMap{
			Namespace: namespace,
			Name:      "services",
		}
		op, err := controllerutil.CreateOrUpdate(ctx, c, cm, func() error {
			if !controllerutil.HasControllerReference(cm) {
				err := controllerutil.SetControllerReference(parent, cm, c.Scheme())
				if err != nil {
					l.Error(err, "failed to set controller reference on service configmap")
				}
			}

			cm.Labels = commonLabels

			if cm.Data == nil {
				cm.Data = map[string]string{}
			}

			touchedKeys := []string{}

			for name, domain := range exposeMap {
				if name == "" {
					name = "chall"
				}
				touchedKeys = append(touchedKeys, name)
				cm.Data[name] = domain
			}

			for k := range maps.Clone(cm.Data) {
				if !slices.Contains(touchedKeys, k) {
					delete(cm.Data, k)
				}
			}

			return nil
		})
		if err != nil {
			l.Error(err, "services configmap reconcile error")
			return fmt.Errorf("services configmap reconcile: %w", err)
		}
		l.Info("services configmap reconciled", "operation", op)
	}
	{
		cm := &corev1.ConfigMap{
			Namespace: namespace,
			Name:      flagConfigMapName,
		}
		if flag != nil {
			op, err := controllerutil.CreateOrUpdate(ctx, c, cm, func() error {
				if !controllerutil.HasControllerReference(cm) {
					err := controllerutil.SetControllerReference(parent, cm, c.Scheme())
					if err != nil {
						l.Error(err, "failed to set controller reference on flag configmap")
					}
				}

				cm.Labels = commonLabels

				if cm.Immutable == nil || !*cm.Immutable {
					cm.Immutable = new(true)
				}
				if cm.Data == nil {
					cm.Data = map[string]string{}
				}
				if _, found := cm.Data[flagKey]; !found {
					cm.Data[flagKey] = *flag
				}
				if _, found := cm.Data[flagLFKey]; !found {
					cm.Data[flagLFKey] = *flag + "\n"
				}
				return nil
			})
			if err != nil {
				l.Error(err, "flag configmap reconcile error")
				return fmt.Errorf("flag configmap reconcile: %w", err)
			}
			l.Info("flag configmap reconciled", "operation", op)
		} else {
			err := client.IgnoreNotFound(c.Delete(ctx, cm))
			if err != nil {
				l.Error(err, "flag configmap delete error")
				return err
			}
		}
	}
	return nil
}
