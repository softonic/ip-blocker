package actor

import (
	"github.com/softonic/ip-blocker/app/utils"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"

	"k8s.io/klog"
)

type IPGetter struct {
	actor *GCPArmorActor
}

func NewIPGetter(actor *GCPArmorActor) *IPGetter {
	return &IPGetter{actor: actor}
}

func (g *IPGetter) GetBlockedIPs() ([]string, error) {

	var sourceIps computepb.SecurityPolicyRuleMatcherConfig

	var ips []string

	RulesGetter := NewRulesGetter(g.actor)
	rules, err := RulesGetter.GetSecurityRules()
	if err != nil {
		klog.Error("\nError: ", err)
		return nil, err
	}

	for _, singleRule := range rules {

		sourceIps = computepb.SecurityPolicyRuleMatcherConfig{
			SrcIpRanges: singleRule.Match.Config.SrcIpRanges,
		}

		ips = append(ips, sourceIps.SrcIpRanges...)

	}

	return ips, nil

}

// getLastPriority: this function will get the last priority of the rules
func getLastPriority(rules []*computepb.SecurityPolicyRule) int32 {

	var lastPriority int32

	for _, singleRule := range rules {

		if *singleRule.Priority > lastPriority {
			lastPriority = *singleRule.Priority
		}

	}

	if lastPriority == 0 {
		lastPriority = 1000
	}

	return lastPriority

}

func getCandidateIPsToBlock(sourceIPs, actorIPs, excludedIPs []string) []string {
	candidateIPsToBlock := utils.UniqueItems(sourceIPs, actorIPs)
	candidateAfterExcluded := utils.UniqueItems(candidateIPsToBlock, excludedIPs)
	candidateAfterExcluded = utils.RemoveDuplicateStr(candidateAfterExcluded)

	var candidateWithCidr []string
	for _, k := range candidateAfterExcluded {
		candidateWithCidr = append(candidateWithCidr, k+"/32")
	}
	return candidateWithCidr
}
