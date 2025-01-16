package utils

import corev1 "k8s.io/api/core/v1"

type NodeType string

const (
	NodeTypeShared   NodeType = "shared-challs"
	NodeTypeIsolated NodeType = "isolated-challs"
)

const nodeRoleKey = "infra.pwnlentoni.team/node-role"

func (nt NodeType) Tolerations() []corev1.Toleration {
	switch nt {
	case NodeTypeShared:
		fallthrough
	case NodeTypeIsolated:
		return []corev1.Toleration{
			{
				Key:      nodeRoleKey,
				Operator: corev1.TolerationOpEqual,
				Value:    string(nt),
				Effect:   corev1.TaintEffectNoSchedule,
			},
		}
	default:
		panic("bad nodetype: " + nt)
	}
}

func (nt NodeType) Affinity() *corev1.Affinity {
	switch nt {
	case NodeTypeShared:
		fallthrough
	case NodeTypeIsolated:
		return &corev1.Affinity{
			NodeAffinity: &corev1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
					NodeSelectorTerms: []corev1.NodeSelectorTerm{
						{
							MatchExpressions: []corev1.NodeSelectorRequirement{
								{
									Key:      nodeRoleKey,
									Operator: corev1.NodeSelectorOpIn,
									Values:   []string{string(nt)},
								},
							},
						},
					},
				},
			},
		}
	default:
		panic("bad nodetype: " + nt)
	}
}
