package reconcilers

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	prismctfv1 "github.com/pwnlentoni/prism-ctf/api/v1"
	"github.com/pwnlentoni/prism-ctf/internal/utils"
	"google.golang.org/protobuf/proto"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"slices"
	"strings"
)

var gatewayNamespace = gatewayv1.Namespace("prism-ctf-system")
var gatewayKind = gatewayv1.Kind("Gateway")
var httpSection = gatewayv1.SectionName("https")
var tcpSection = gatewayv1.SectionName("tls")

func gatewayParentRefHttp() []gatewayv1.ParentReference {
	return []gatewayv1.ParentReference{{
		Name:        "prism-ctf-challenge-gateway",
		Namespace:   &gatewayNamespace,
		SectionName: &httpSection,
		Kind:        &gatewayKind,
	}}
}

func gatewayParentRefTcp() []gatewayv1.ParentReference {
	return []gatewayv1.ParentReference{{
		Name:        "prism-ctf-challenge-gateway",
		Namespace:   &gatewayNamespace,
		SectionName: &tcpSection,
		Kind:        &gatewayKind,
	}}
}

var pathMatchPathPrefix = gatewayv1.PathMatchPathPrefix
var serviceKind = gatewayv1.Kind("Service")

func ReconcileIngress(ctx context.Context, c client.Client, namespace string, commonLabels map[string]string, parent metav1.Object, exposes []prismctfv1.ExposeSpec, challengeName, domain string) ([]prismctfv1.ExposeStatus, error) {
	l := log.FromContext(ctx)

	exposedUrls := make([]prismctfv1.ExposeStatus, 0, len(exposes))

	httpRouteDeleter, err := utils.NewDeleter(ctx, c, &gatewayv1.HTTPRoute{}, namespace)
	if err != nil {
		l.Error(err, "deployment deleter error")
	}

	tlsRouteDeleter, err := utils.NewDeleter(ctx, c, &gatewayv1alpha2.TLSRoute{}, namespace)
	if err != nil {
		l.Error(err, "service deleter error")
	}

	for _, expose := range exposes {
		meta := metav1.ObjectMeta{
			Name:      strings.ToLower(fmt.Sprintf("%s-%s-%d", expose.Container, expose.Protocol, expose.Port)),
			Namespace: namespace,
		}
		exposeHost := challengeName
		if len(expose.Name) != 0 {
			exposeHost += "-" + expose.Name
		}
		exposeHost += domain

		exposedUrls = append(exposedUrls, prismctfv1.ExposeStatus{
			Hostname: exposeHost,
			Protocol: expose.Protocol,
		})

		backendPort := gatewayv1.PortNumber(expose.Port)
		backendNamespace := gatewayv1.Namespace(namespace)
		backendRef := gatewayv1.BackendRef{BackendObjectReference: gatewayv1.BackendObjectReference{
			Kind:      &serviceKind,
			Name:      gatewayv1.ObjectName(expose.Container),
			Port:      &backendPort,
			Namespace: &backendNamespace,
		},
		}
		switch expose.Protocol {
		case prismctfv1.ExposeProtocolHTTP:
			{
				route := &gatewayv1.HTTPRoute{
					ObjectMeta: meta,
				}
				httpRouteDeleter.MarkUsed(ctx, route)
				op, err := controllerutil.CreateOrUpdate(ctx, c, route, func() error {
					if !controllerutil.HasControllerReference(route) {
						err := controllerutil.SetControllerReference(parent, route, c.Scheme())
						if err != nil {
							l.Error(err, "failed to set controller reference on http route")
						}
					}
					route.Labels = commonLabels
					route.Spec.CommonRouteSpec.ParentRefs = gatewayParentRefHttp()
					route.Spec.Hostnames = []gatewayv1.Hostname{
						gatewayv1.Hostname(exposeHost),
					}
					route.Spec.Rules = []gatewayv1.HTTPRouteRule{{
						Matches: []gatewayv1.HTTPRouteMatch{{
							Path: &gatewayv1.HTTPPathMatch{
								Type:  &pathMatchPathPrefix,
								Value: proto.String("/"),
							},
						}},
						BackendRefs: []gatewayv1.HTTPBackendRef{{
							BackendRef: backendRef,
						}},
					}}
					return nil
				})
				if err != nil {
					return nil, fmt.Errorf("http route `%s`: %w", meta.Name, err)
				}
				l.Info("http route reconciled", "operation", op, "route", meta.Name)
			}
		case prismctfv1.ExposeProtocolTCP:
			{
				route := &gatewayv1alpha2.TLSRoute{
					ObjectMeta: meta,
				}
				tlsRouteDeleter.MarkUsed(ctx, route)
				op, err := controllerutil.CreateOrUpdate(ctx, c, route, func() error {
					if !controllerutil.HasControllerReference(route) {
						err := controllerutil.SetControllerReference(parent, route, c.Scheme())
						if err != nil {
							l.Error(err, "failed to set controller reference on tcp route")
						}
					}
					route.Labels = commonLabels
					route.Spec.CommonRouteSpec.ParentRefs = gatewayParentRefTcp()
					route.Spec.Hostnames = []gatewayv1.Hostname{
						gatewayv1.Hostname(exposeHost),
					}
					route.Spec.Rules = []gatewayv1alpha2.TLSRouteRule{{
						BackendRefs: []gatewayv1.BackendRef{backendRef},
					}}
					return nil
				})
				if err != nil {
					return nil, fmt.Errorf("tcp route `%s`: %w", meta.Name, err)
				}
				l.Info("tcp route reconciled", "operation", op, "route", meta.Name)
			}
		case prismctfv1.ExposeProtocolUDP:
			{
				return nil, errors.New("UDP not yet supported") // TODO
			}
		}
	}
	err = httpRouteDeleter.DeleteUnused(ctx)
	if err != nil {
		l.Error(err, "failed to delete unused http routes")
	}
	err = tlsRouteDeleter.DeleteUnused(ctx)
	if err != nil {
		l.Error(err, "failed to delete unused tls routes")
	}
	slices.SortFunc(exposedUrls, func(a, b prismctfv1.ExposeStatus) int {
		return cmp.Compare(a.Hostname, b.Hostname)
	})
	return exposedUrls, nil
}
