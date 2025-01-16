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
	GatewayAllowLabel       = "prism-ctf.pwnlentoni.team/allow-gateway-routes"
	GatewayAllowValue       = "true"
)
