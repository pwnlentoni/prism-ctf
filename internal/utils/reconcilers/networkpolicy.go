package reconcilers

import (
	"context"
	"fmt"
	ciliumv2 "github.com/cilium/cilium/pkg/k8s/apis/cilium.io/v2"
	slim_metav1 "github.com/cilium/cilium/pkg/k8s/slim/k8s/apis/meta/v1"
	ciliumapi "github.com/cilium/cilium/pkg/policy/api"
	"github.com/pwnlentoni/prism-ctf/internal/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"maps"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func ReconcileNetworkPolicies(ctx context.Context, c client.Client, namespace string, commonLabels map[string]string, parent metav1.Object) error {
	l := log.FromContext(ctx)

	isolatePolicy := &ciliumv2.CiliumNetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "isolate",
		},
	}
	prefixedLabels := make(map[string]string)
	for k, v := range commonLabels {
		prefixedLabels["k8s."+k] = v
	}
	op, err := controllerutil.CreateOrUpdate(ctx, c, isolatePolicy, func() error {
		if !controllerutil.HasControllerReference(isolatePolicy) {
			err := controllerutil.SetControllerReference(parent, isolatePolicy, c.Scheme())
			if err != nil {
				l.Error(err, "failed to set controller reference on isolate policy")
			}
		}
		ingressRules := []ciliumapi.IngressRule{
			{ // allow ingress from other challenge containers
				IngressCommonRule: ciliumapi.IngressCommonRule{FromEndpoints: []ciliumapi.EndpointSelector{
					{
						LabelSelector: &slim_metav1.LabelSelector{
							MatchLabels: maps.Clone(prefixedLabels),
						},
					},
				}},
			},
		}
		// These two rules require that the cluster isn't pwned, specifically:
		// 	- attackers can't create NodePort services
		// 	- attackers can't edit traefik config
		if *utils.NodePortMode {
			ingressRules = append(ingressRules, ciliumapi.IngressRule{ // allow ingress from world
				IngressCommonRule: ciliumapi.IngressCommonRule{FromEntities: []ciliumapi.Entity{
					ciliumapi.EntityWorld,
				}},
			})
		} else {
			ingressRules = append(ingressRules, ciliumapi.IngressRule{ // allow ingress from traefik
				IngressCommonRule: ciliumapi.IngressCommonRule{FromEndpoints: []ciliumapi.EndpointSelector{
					{
						LabelSelector: &slim_metav1.LabelSelector{
							MatchLabels: map[string]string{
								"k8s.io.kubernetes.pod.namespace": "traefik",
							},
						},
					},
				}},
			})
		}
		isolatePolicy.Spec = ciliumapi.NewRule().WithEndpointSelector(ciliumapi.WildcardEndpointSelector).WithIngressRules(ingressRules).WithEgressRules([]ciliumapi.EgressRule{
			{
				EgressCommonRule: ciliumapi.EgressCommonRule{ToEndpoints: []ciliumapi.EndpointSelector{
					{ // allow egress to other challenge containers
						LabelSelector: &slim_metav1.LabelSelector{
							MatchLabels: maps.Clone(prefixedLabels),
						},
					},
				}},
			},
			{
				EgressCommonRule: ciliumapi.EgressCommonRule{ToEndpoints: []ciliumapi.EndpointSelector{
					{ // allow egress to explicitly marked pods
						LabelSelector: &slim_metav1.LabelSelector{
							MatchLabels: map[string]string{
								"k8s." + utils.AccessibleByChallengesLabel: utils.AccessibleByChallengesValue,
							},
							MatchExpressions: []slim_metav1.LabelSelectorRequirement{{
								Key:      "k8s.io.kubernetes.pod.namespace",
								Operator: slim_metav1.LabelSelectorOpExists,
							}},
						},
					},
				}},
			},
			{ // allow egress to kube-dns
				EgressCommonRule: ciliumapi.EgressCommonRule{
					ToEndpoints: []ciliumapi.EndpointSelector{
						{
							LabelSelector: &slim_metav1.LabelSelector{
								MatchLabels: map[string]string{
									"k8s.io.kubernetes.pod.namespace": "kube-system",
									"k8s-app":                         "kube-dns",
								},
							},
						},
					}},
				ToPorts: []ciliumapi.PortRule{{
					Ports: []ciliumapi.PortProtocol{{Port: "53", Protocol: ciliumapi.ProtoUDP}},
					Rules: &ciliumapi.L7Rules{
						DNS: []ciliumapi.PortRuleDNS{{MatchPattern: "*"}},
					},
				}},
			},
		})
		return nil
	})
	if err != nil {
		return fmt.Errorf("isolate policy: %w", err)
	}
	l.Info("isolate policy reconciled", "operation", op)

	egressPolicy := &ciliumv2.CiliumNetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "allow-egress",
		},
	}
	op, err = controllerutil.CreateOrUpdate(ctx, c, egressPolicy, func() error {
		if !controllerutil.HasControllerReference(egressPolicy) {
			err := controllerutil.SetControllerReference(parent, egressPolicy, c.Scheme())
			if err != nil {
				l.Error(err, "failed to set controller reference on egress policy")
			}
		}
		egressPolicy.Spec = ciliumapi.NewRule().WithEndpointSelector(ciliumapi.EndpointSelector{
			LabelSelector: &slim_metav1.LabelSelector{
				MatchLabels: map[string]string{
					"k8s." + utils.EgressEnabledLabel: utils.EgressEnabledValue,
				}}}).WithEgressRules([]ciliumapi.EgressRule{
			{ // allow egress to everything except private IPs
				EgressCommonRule: ciliumapi.EgressCommonRule{
					ToCIDRSet: []ciliumapi.CIDRRule{
						{
							Cidr: "0.0.0.0/0",
							ExceptCIDRs: []ciliumapi.CIDR{
								"10.0.0.0/8",
								"192.168.0.0/16",
								"172.16.0.0/20",
								"169.254.169.0/24",
							},
						},
					},
				},
			},
		})
		return nil
	})
	if err != nil {
		return fmt.Errorf("egress policy: %w", err)
	}
	l.Info("egress policy reconciled", "operation", op)

	return nil
}
