package utils

import "strings"

const SharedChallengesNamespacePrefix = "prism-ctf-shared-"
const IsolatedChallengesNamespacePrefix = "prism-ctf-isolated-"

// +kubebuilder:rbac:groups=*,resources=namespaces;httproutes;tlsroutes;deployments;services;ciliumnetworkpolicies,verbs=get;list;watch;create;update;patch;delete

func SharedChallengeNamespace(challenge string) string {
	return SharedChallengesNamespacePrefix + challenge
}

func IsolatedChallengeNamespace(challenge string, team string) string {
	b := strings.Builder{}
	b.Grow(len(IsolatedChallengesNamespacePrefix) + len(challenge) + len(team) + 1)
	b.WriteString(IsolatedChallengesNamespacePrefix)
	b.WriteString(challenge)
	b.WriteRune('-')
	b.WriteString(team)
	return b.String()
}
