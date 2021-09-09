package gcp

import (
	"context"
	"fmt"
	"log"
	"os"

	compute "cloud.google.com/go/compute/apiv1"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
)

var (
	stdlog, errlog *log.Logger
)

type GCPArmor struct {
	Client *compute.SecurityPoliciesClient
	Ctx    context.Context
}

func GCPConnection() GCPArmor {

	ctx := context.Background()
	c, err := compute.NewSecurityPoliciesRESTClient(ctx)
	if err != nil {
		fmt.Println("\nError: ", err)
		os.Exit(1)
	}

	return GCPArmor{
		Client: c,
		Ctx:    ctx,
	}

}

func GetArmorRules(c GCPArmor) ([]string, int32) {

	client := c.Client
	ctx := c.Ctx

	req := &computepb.GetSecurityPolicyRequest{
		Project:        "kubertonic",
		SecurityPolicy: "global-loadbalancer-rules",
	}

	defer client.Close()

	resp, err := client.Get(ctx, req)
	if err != nil {
		// TODO: Handle error.
		errlog.Println("\nError: ", err)
		os.Exit(1)
	}

	rules := make([]computepb.SecurityPolicyRule, 10)

	configs := make([]computepb.SecurityPolicyRuleMatcherConfig, 100)

	var sourceIps computepb.SecurityPolicyRuleMatcherConfig

	var ips []string

	var lastPriority int32

	for _, singleRule := range resp.Rules {

		rule := computepb.SecurityPolicyRule{
			Action:      singleRule.Action,
			Description: singleRule.Kind,
			Kind:        singleRule.Kind,
			Match:       singleRule.Match,
			Priority:    singleRule.Priority,
		}

		if *singleRule.Action != "allow" {

			sourceIps = computepb.SecurityPolicyRuleMatcherConfig{
				SrcIpRanges: singleRule.Match.Config.SrcIpRanges,
			}

			ips = append(ips, sourceIps.SrcIpRanges...)

			// get last priority

			if *singleRule.Priority > lastPriority {
				lastPriority = *singleRule.Priority
			}

		}
		rules = append(rules, rule)
		configs = append(configs, sourceIps)

	}

	fmt.Println(rules)

	fmt.Println(ips)

	fmt.Println(lastPriority)

	return ips, lastPriority

}

func BlockIPsFromArmor(c GCPArmor, prio int32, candidateIPsBlocked []string) (error, string) {

	client := c.Client
	ctx := c.Ctx

	defer client.Close()

	var versioned *computepb.SecurityPolicyRuleMatcher_VersionedExpr

	versioned = computepb.SecurityPolicyRuleMatcher_SRC_IPS_V1.Enum()

	action := "deny(403)"
	description := "ip blocker"
	priority := prio + 1
	preview := true

	if len(candidateIPsBlocked) > 0 {

		match := &computepb.SecurityPolicyRuleMatcher{
			Config: &computepb.SecurityPolicyRuleMatcherConfig{
				SrcIpRanges: candidateIPsBlocked,
			},
			VersionedExpr: versioned,
		}

		req := &computepb.AddRuleSecurityPolicyRequest{

			Project:        "kubertonic",
			SecurityPolicy: "test",
			SecurityPolicyRuleResource: &computepb.SecurityPolicyRule{
				Action:      &action,
				Description: &description,
				Priority:    &priority,
				Preview:     &preview,
				Match:       match,
			},
		}

		resp, err := client.AddRule(ctx, req)
		if err != nil {
			// TODO: Handle error.
			fmt.Println("\nError: ", err)
			os.Exit(1)
		}

		return nil, *resp.Proto().StatusMessage

	}

	return nil, ""

}
