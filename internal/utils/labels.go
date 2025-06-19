package utils

const (
	ManagedByLabel          = "app.kubernetes.io/managed-by"
	ManagedByValue          = "prism-ctf"
	ChallengeLabel          = "prism-ctf.pwnlentoni.team/challenge"
	EgressEnabledLabel      = "prism-ctf.pwnlentoni.team/egress-enabled"
	EgressEnabledValue      = "true"
	ChallengeNamespaceLabel = "prism-ctf.pwnlentoni.team/challenge-namespace"
	ChallengeNamespaceValue = "true"
	ContainerNameLabel      = "prism-ctf.pwnlentoni.team/container"
	// GatewayAllowLabel is also used in config/ingress/gateway.yaml and config/placeholders/namespace.yaml
	GatewayAllowLabel  = "prism-ctf.pwnlentoni.team/allow-gateway-routes"
	GatewayAllowValue  = "true"
	ChallengeTeamLabel = "prism-ctf.pwnlentoni.team/team"

	// AccessibleByChallengesLabel is used to mark pods not managed by prism-ctf that should be accessible to challenges
	AccessibleByChallengesLabel = "prism-ctf.pwnlentoni.team/accessible-by-challenges"
	AccessibleByChallengesValue = "true"
)

func MakeCommonLabels(challengeName string) map[string]string {
	return map[string]string{
		ManagedByLabel:    ManagedByValue,
		ChallengeLabel:    challengeName,
		GatewayAllowLabel: GatewayAllowValue,
	}
}

func MakeInstancedLabels(team string) map[string]string {
	return map[string]string{
		ChallengeTeamLabel: team,
	}
}
