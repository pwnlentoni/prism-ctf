package utils

import "flag"

var challengesDomain = flag.String("challs-domain", "challs.test.local", "Domain under which the challenges should be hosted")
var domainSuffix = ""

func DomainSuffix() string {
	if len(domainSuffix) == 0 {
		domainSuffix = "." + *challengesDomain
	}
	return domainSuffix
}
