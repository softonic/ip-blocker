package app

import (
	"context"
	"fmt"
	"os"

	compute "cloud.google.com/go/compute/apiv1"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
)

// conectar a GCP
// usar json para private key
// get rules de armor GCP
// comprobar si las IPs que hemos sacado estan ya en las rules

func ConnectGCP() ([]string, int32) {

	ctx := context.Background()
	c, err := compute.NewSecurityPoliciesRESTClient(ctx)

	if err != nil {
		// TODO: Handle error.
		fmt.Println("\nError: ", err)
		os.Exit(1)
	}
	defer c.Close()

	req := &computepb.GetSecurityPolicyRequest{
		// TODO: Fill request struct fields.
		// See https://pkg.go.dev/google.golang.org/genproto/googleapis/cloud/compute/v1#GetSecurityPolicyRequest.
		Project:        "kubertonic",
		SecurityPolicy: "global-loadbalancer-rules",
	}

	resp, err := c.Get(ctx, req)
	if err != nil {
		// TODO: Handle error.
		errlog.Println("\nError: ", err)
		os.Exit(1)
	}
	// TODO: Use resp.

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

	return ips, lastPriority

}

func BlockedIPs(prio int32) string {

	ctx := context.Background()
	c, err := compute.NewSecurityPoliciesRESTClient(ctx)

	if err != nil {
		// TODO: Handle error.
		fmt.Println("\nError: ", err)
		os.Exit(1)
	}
	defer c.Close()

	action := "deny(403)"
	description := "ip blocker"
	priority := prio + 1
	preview := true

	match := &computepb.SecurityPolicyRuleMatcher{
		Config: &computepb.SecurityPolicyRuleMatcherConfig{
			SrcIpRanges: []string{
				"1.1.1.1/32",
			},
		},
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

	resp, err := c.AddRule(ctx, req)
	if err != nil {
		// TODO: Handle error.
		fmt.Println("\nError: ", err)
		os.Exit(1)
	}

	_ = resp

	return "done"

}
