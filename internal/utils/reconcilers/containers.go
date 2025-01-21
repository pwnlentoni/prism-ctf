package reconcilers

import (
	"context"
	"fmt"
	prismctfv1 "github.com/pwnlentoni/prism-ctf/api/v1"
	"github.com/pwnlentoni/prism-ctf/internal/utils"
	"google.golang.org/protobuf/proto"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func ReconcileContainers(ctx context.Context, c client.Client, namespace string, commonLabels map[string]string, parent metav1.Object, containers []prismctfv1.ContainerSpec, nodeType utils.NodeType) error {
	l := log.FromContext(ctx)

	deploymentDeleter, err := utils.NewDeleter(ctx, c, &appsv1.Deployment{}, namespace)
	if err != nil {
		l.Error(err, "deployment deleter error")
	}

	serviceDeleter, err := utils.NewDeleter(ctx, c, &corev1.Service{}, namespace)
	if err != nil {
		l.Error(err, "service deleter error")
	}

	for _, container := range containers {
		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      container.Spec.Name,
				Namespace: namespace,
			},
		}
		deploymentDeleter.MarkUsed(ctx, deployment)
		op, err := controllerutil.CreateOrUpdate(ctx, c, deployment, func() error {

			if !controllerutil.HasControllerReference(deployment) {
				err = controllerutil.SetControllerReference(parent, deployment, c.Scheme())
				if err != nil {
					l.Error(err, "failed to set controller reference on deployment")
				}
			}

			deployment.Labels = commonLabels
			if deployment.ObjectMeta.CreationTimestamp.IsZero() {
				deployment.Spec.Selector = &metav1.LabelSelector{MatchLabels: utils.MapMerge(map[string]string{
					utils.ContainerNameLabel: container.Spec.Name,
				}, commonLabels)}
			}

			deployment.Spec.Replicas = proto.Int32(int32(container.Replicas))
			deployment.Spec.Template.ObjectMeta.Labels = utils.MapMerge(utils.MapMerge(map[string]string{
				utils.ContainerNameLabel: container.Spec.Name,
			}, commonLabels), deployment.Spec.Template.ObjectMeta.Labels)
			if container.Egress {
				deployment.Spec.Template.ObjectMeta.Labels[utils.EgressEnabledLabel] = utils.EgressEnabledValue
			} else {
				delete(deployment.Spec.Template.ObjectMeta.Labels, utils.EgressEnabledLabel)
			}
			deployment.Spec.Template.Spec.EnableServiceLinks = proto.Bool(false)
			deployment.Spec.Template.Spec.AutomountServiceAccountToken = proto.Bool(false)
			deployment.Spec.Template.Spec.Containers = []corev1.Container{
				*container.Spec,
			}
			deployment.Spec.Template.Spec.Tolerations = nodeType.Tolerations()
			deployment.Spec.Template.Spec.Affinity = nodeType.Affinity()
			deployment.Spec.Template.Spec.TerminationGracePeriodSeconds = proto.Int64(1)
			return nil
		})
		if err != nil {
			l.Error(err, "deployment reconcile error", "deployment", container.Spec.Name)
			return fmt.Errorf("deployment reconcile `%s`: %w", container.Spec.Name, err)
		}
		l.Info("deployment reconciled", "operation", op, "deployment", container.Spec.Name)

		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      container.Spec.Name,
				Namespace: namespace,
			},
		}
		serviceDeleter.MarkUsed(ctx, svc)
		op, err = controllerutil.CreateOrUpdate(ctx, c, svc, func() error {
			if !controllerutil.HasControllerReference(svc) {
				err := controllerutil.SetControllerReference(parent, svc, c.Scheme())
				if err != nil {
					l.Error(err, "failed to set controller reference on service")
				}
			}
			svc.Labels = commonLabels
			if svc.ObjectMeta.CreationTimestamp.IsZero() {
				svc.Spec.Selector = utils.MapMerge(map[string]string{
					utils.ContainerNameLabel: container.Spec.Name,
				}, commonLabels)
			}
			svc.Spec.Ports = make([]corev1.ServicePort, len(container.Ports))
			for i, port := range container.Ports {
				svc.Spec.Ports[i] = corev1.ServicePort{
					Port:     int32(port.Port),
					Protocol: port.Protocol,
				}
			}
			return nil
		})
		if err != nil {
			l.Error(err, "service reconcile error", "service", container.Spec.Name)
			return fmt.Errorf("service reconcile `%s`: %w", container.Spec.Name, err)
		}
		l.Info("service reconciled", "operation", op, "service", container.Spec.Name)
	}

	err = deploymentDeleter.DeleteUnused(ctx)
	if err != nil {
		l.Error(err, "failed to delete unused deployments")
	}
	err = serviceDeleter.DeleteUnused(ctx)
	if err != nil {
		l.Error(err, "failed to delete unused deployments")
	}

	return nil
}
