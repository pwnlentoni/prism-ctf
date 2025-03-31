package utils

import (
	"errors"
	"flag"
	"github.com/go-logr/logr"
)

var challengesDomain = flag.String("challs-domain", "challs.test.local", "Domain under which the challenges should be hosted")
var domainSuffix = ""

func DomainSuffix() string {
	if len(domainSuffix) == 0 {
		domainSuffix = "." + *challengesDomain
	}
	return domainSuffix
}

func ChallengesDomain() string {
	return *challengesDomain
}

var MaxInstancesPerTeam = flag.Int("max-team-instances", 4, "Max instances per team. Set to 0 to disable limit.")
var RandomTokenLength = flag.Int("random-token-len", 4, "Random token length.")
var NodePortMode = flag.Bool("nodeport-mode", false, "Expose challenges via a NodePort instead of TLS.")
var UseAffinity = flag.Bool("use-affinity", true, "Use node affinity for challenges")

func ValidateConfig(l logr.Logger) error {
	if *MaxInstancesPerTeam < 0 {
		return errors.New("max instances per team is less than 0")
	}
	if *MaxInstancesPerTeam == 0 {
		l.Info("max instances per team is set to 0, don't do this in production!")
	}
	if *RandomTokenLength <= 0 {
		return errors.New("random token length must be greater than 0")
	}
	if *RandomTokenLength > 8 {
		l.Info("random token length might be too long", "length", *RandomTokenLength)
	}
	return nil
}
