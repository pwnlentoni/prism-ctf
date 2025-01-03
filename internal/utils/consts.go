package utils

const SharedChallengesNamespace = "prism-ctf-challenges"

// +kubebuilder:rbac:groups=*,resources=ingresses;ingressroutes;ingressroutetcps;pods;deployments;services,verbs=get;list;watch;create;update;patch;delete,namespace=prism-ctf-challenges
