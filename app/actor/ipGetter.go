package actor

import (
	"github.com/softonic/ip-blocker/app/actor/utils"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
)

type IPGetter struct {
	actor *GCPArmorActor
}

func NewIPGetter(actor *GCPArmorActor) *IPGetter {
	return &IPGetter{actor: actor}
}

// GetIPsToBlock: this function will get the IPs already blocked
func (g *IPGetter) GetBlockedIPs(rules []*computepb.SecurityPolicyRule) ([]string, error) {

	var sourceIps computepb.SecurityPolicyRuleMatcherConfig

	var ips []string

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

func getCandidateIPsToBlock(handler utils.IPListHandler, candidateIPs, alreadyBlockedIPs, excludedIPs []string) []string {
	candidateIPs = handler.UniqueItems(candidateIPs, alreadyBlockedIPs)
	candidateAfterExcluded := handler.UniqueItems(candidateIPs, excludedIPs)
	candidateAfterExcluded = handler.RemoveDuplicateStr(candidateAfterExcluded)

	return addCidrMaskToIPs(candidateAfterExcluded)
}

// This function add the cidr mask to the IPs
func addCidrMaskToIPs(ips []string) []string {
	var ipsWithCidr []string
	for _, k := range ips {
		ipsWithCidr = append(ipsWithCidr, k+"/32")
	}
	return ipsWithCidr
}
