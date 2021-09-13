package actor

import (
	"context"
	"fmt"
	"log"
	"os"

	compute "cloud.google.com/go/compute/apiv1"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"

	"github.com/softonic/ip-blocker/app"
)

var (
	stdlog, errlog *log.Logger
)

type GCPArmorActor struct {
	client     *compute.SecurityPoliciesClient
	ctx        context.Context
	k8sProject string
	policy     string
}

func NewGCPArmorActor(project string, policy string) *GCPArmorActor {

	c, ctx := InitConnectiontoActor()

	return &GCPArmorActor{
		client:     c,
		ctx:        ctx,
		k8sProject: project,
		policy:     policy,
	}
}

func InitConnectiontoActor() (*compute.SecurityPoliciesClient, context.Context) {

	ctx := context.Background()
	c, err := compute.NewSecurityPoliciesRESTClient(ctx)
	if err != nil {
		fmt.Println("\nError: ", err)
		os.Exit(1)
	}

	return c, ctx

}

func getIPsAlreadyBlockedFromRules(g *GCPArmorActor, securityPolicy string) ([]string, int32) {

	client := g.client
	ctx := g.ctx

	req := &computepb.GetSecurityPolicyRequest{
		Project:        g.k8sProject,
		SecurityPolicy: securityPolicy,
	}

	//defer client.Close()

	resp, err := client.Get(ctx, req)
	if err != nil {
		fmt.Println("\nError: ", err)
		os.Exit(1)
	}

	var sourceIps computepb.SecurityPolicyRuleMatcherConfig

	var ips []string

	var lastPriority int32

	for _, singleRule := range resp.Rules {

		/* 		rule := computepb.SecurityPolicyRule{
			Action:      singleRule.Action,
			Description: singleRule.Kind,
			Kind:        singleRule.Kind,
			Match:       singleRule.Match,
			Priority:    singleRule.Priority,
		} */

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

	}

	return ips, lastPriority

}

func (g *GCPArmorActor) BlockIPs(sourceIPs []app.IPCount) error {

	client := g.client
	ctx := g.ctx
	project := g.k8sProject

	//defer client.Close()

	actorIPs, prio := getIPsAlreadyBlockedFromRules(g, g.policy)

	candidateIPsBlocked := compareBlockedIps(sourceIPs, actorIPs)

	//var versioned *computepb.SecurityPolicyRuleMatcher_VersionedExpr

	versioned := computepb.SecurityPolicyRuleMatcher_SRC_IPS_V1.Enum()

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

			Project:        project,
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
			return err
		}

		_ = resp

		return nil

	}

	return nil

}

func compareBlockedIps(sourceIPs []app.IPCount, ips []string) []string {

	// compare the array of IPs of ES with the IPs of GCP armor

	var count int
	var ipWithMaskES string
	candidateIPsBlocked := []string{}

	for _, elasticIps := range sourceIPs {
		for _, armorIps := range ips {
			ipWithMaskES = elasticIps.IP + "/32"
			if ipWithMaskES == armorIps {
				fmt.Println("this IPs are already in the armor")
				count++
			}
		}
		fmt.Println("this IP is not in armor:", ipWithMaskES)
		candidateIPsBlocked = append(candidateIPsBlocked, ipWithMaskES)
	}

	return candidateIPsBlocked

}
