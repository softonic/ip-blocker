package actor

import (
	"strconv"
	"time"

	actorUtils "github.com/softonic/ip-blocker/app/actor/utils"
	globalUtils "github.com/softonic/ip-blocker/app/utils"

	"k8s.io/klog"

	"regexp"

	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
)

type RulesGetter struct {
	actor *GCPArmorActor
}

func NewRulesGetter(actor *GCPArmorActor) *RulesGetter {
	return &RulesGetter{actor: actor}
}

// GetSecurityRules: get Rules that were created by IPBlocker
func (g *RulesGetter) GetSecurityRules() ([]*computepb.SecurityPolicyRule, error) {
	req := &computepb.GetSecurityPolicyRequest{
		Project:        g.actor.ActorConfig.Project,
		SecurityPolicy: g.actor.ActorConfig.Policy,
	}

	resp, err := g.actor.Connection.Client.Get(g.actor.Connection.Ctx, req)
	if err != nil {
		klog.Error("\nError: ", err)
		return nil, err
	}

	var ipBlockerRules []*computepb.SecurityPolicyRule

	for _, singleRule := range resp.Rules {

		match := false
		match, _ = regexp.MatchString("ipblocker:[0-9]{10}", *singleRule.Description)

		if !match {
			continue
		}

		// && singleRule.Match.VersionedExpr == computepb.SecurityPolicyRuleMatcher_SRC_IPS_V1.Enum()

		if *singleRule.Action != "allow" && *singleRule.Match.VersionedExpr == 70925961 {

			ipBlockerRules = append(ipBlockerRules, singleRule)

		}

	}

	return ipBlockerRules, nil
}

// GetBlockedIPsFromActorThatCanBeUnblocked: return IPs that has been blocked for more than ttlRules min
func getBlockedIPsFromActorThatCanBeUnblocked(rules []*computepb.SecurityPolicyRule, ttlRules int) []string {

	now := time.Now()
	secs := now.Unix()

	var ips, restIps []string

	for _, singleRule := range rules {

		// && singleRule.Match.VersionedExpr == computepb.SecurityPolicyRuleMatcher_SRC_IPS_V1.Enum()

		unixTimeStampFromDescString := actorUtils.ExtractFromDescription(*singleRule.Description)
		n, err := strconv.ParseInt(unixTimeStampFromDescString, 10, 64)
		if err != nil {
			continue
		}
		if (secs - n) > int64(ttlRules*60) {

			restIps = singleRule.Match.Config.SrcIpRanges

			ips = append(ips, restIps...)
		} else {
			klog.Infof("This rule with priority %d is still valid", *singleRule.Priority)
		}

	}

	return ips

}

func GetRuleFromIP(rules []*computepb.SecurityPolicyRule, ips []string) ([]int32, error) {

	var prios []int32

	for _, singleRule := range rules {

		for _, k := range singleRule.Match.Config.SrcIpRanges {
			for _, m := range ips {
				if k == m {
					found := globalUtils.Find(prios, *singleRule.Priority)
					// check if prio already exists in array []prio
					// so the ip is in the rule, you can get the prio of this rule
					if !found {
						prios = append(prios, *singleRule.Priority)
					}

				}
			}
		}

	}

	return prios, nil

}
