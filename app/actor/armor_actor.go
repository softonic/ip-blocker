package actor

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"k8s.io/klog"

	compute "cloud.google.com/go/compute/apiv1"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"

	"github.com/softonic/ip-blocker/app"
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
		klog.Error("\nError: ", err)
		os.Exit(1)
	}

	return c, ctx

}

func getIPsAlreadyBlockedFromRules(g *GCPArmorActor, securityPolicy string) ([]string, int32) {

	client := g.client
	ctx := g.ctx

	req := &computepb.GetSecurityPolicyRequest{
		Project:        g.k8sProject,
		SecurityPolicy: "test",
	}

	//defer client.Close()

	resp, err := client.Get(ctx, req)
	if err != nil {
		klog.Error("\nError: ", err)
		os.Exit(1)
	}

	var sourceIps computepb.SecurityPolicyRuleMatcherConfig

	var ips []string

	var lastPriority int32

	for _, singleRule := range resp.Rules {

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

	actorIPs, _ := getIPsAlreadyBlockedFromRules(g, g.policy)

	candidateIPstoBlock := detectWhichOfTheseIPsAreNotBlocked(sourceIPs, actorIPs)

	versioned := computepb.SecurityPolicyRuleMatcher_SRC_IPS_V1.Enum()

	now := time.Now()
	secs := now.Unix()

	action := "deny(403)"
	description := strconv.FormatInt(secs, 10)
	//priority := prio + 1
	priority := rand.Int31n(100)
	preview := true

	if len(candidateIPstoBlock) > 0 {

		match := &computepb.SecurityPolicyRuleMatcher{
			Config: &computepb.SecurityPolicyRuleMatcherConfig{
				SrcIpRanges: candidateIPstoBlock,
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
			klog.Error("\nError: ", err)
			return err
		} else {
			klog.Infof("Adding rule with prio: %d", priority)
		}

		_ = resp

		return nil

	}

	return nil

}

func detectWhichOfTheseIPsAreNotBlocked(sourceIPs []app.IPCount, actorIPs []string) []string {

	// compare the array of IPs of ES with the IPs of GCP armor

	var ipWithMaskES string
	candidateIPsBlocked := []string{}

	for _, elasticIps := range sourceIPs {
		count := 0
		for _, armorIps := range actorIPs {
			ipWithMaskES = elasticIps.IP + "/32"
			if ipWithMaskES == armorIps {
				count++
			}
		}
		if count == 0 {
			candidateIPsBlocked = append(candidateIPsBlocked, elasticIps.IP+"/32")
		}

	}

	return candidateIPsBlocked

}

// GetBlockedIPsFromActorThatCanBeUnblocked: return IPs that has been blocked for more than N secs
func getBlockedIPsFromActorThatCanBeUnblocked(g *GCPArmorActor) []string {

	client := g.client
	ctx := g.ctx
	project := g.k8sProject
	//securityPolicy := g.policy

	req := &computepb.GetSecurityPolicyRequest{
		Project:        project,
		SecurityPolicy: "test",
	}

	resp, err := client.Get(ctx, req)
	if err != nil {
		klog.Error("\nError:", err)
		os.Exit(1)
	}

	now := time.Now()
	secs := now.Unix()

	var ips, restIps []string

	for _, singleRule := range resp.Rules {

		if *singleRule.Action != "allow" {

			n, err := strconv.ParseInt(*singleRule.Description, 10, 64)
			if err != nil {
				continue
			}
			if (secs - n) > 301 {

				restIps = singleRule.Match.Config.SrcIpRanges

				ips = append(ips, restIps...)
			} else {
				klog.Infof("This rule with priority %d is still valid", *singleRule.Priority)
			}

		}

	}

	return ips

}

func (g *GCPArmorActor) UnBlockIPs() error {

	client := g.client
	ctx := g.ctx
	project := g.k8sProject

	ips := getBlockedIPsFromActorThatCanBeUnblocked(g)

	prios := getRuleFromIP(g, ips)

	for _, prio := range prios {

		if prio == 0 {
			fmt.Println("ips received", ips)
			fmt.Println("the ptio is:", prio)
			return errors.New("There are no rules in this policy")
		}

		req := &computepb.RemoveRuleSecurityPolicyRequest{

			Project:        project,
			SecurityPolicy: "test",
			Priority:       &prio,
		}

		klog.Infof("Removing the rules with %d", prio)

		resp, err := client.RemoveRule(ctx, req)
		if err != nil {
			klog.Error("\nError: ", err)
			return err
		}

		_ = resp

	}

	return nil

}

func getRuleFromIP(g *GCPArmorActor, ips []string) []int32 {

	client := g.client
	ctx := g.ctx

	var prios []int32

	req := &computepb.GetSecurityPolicyRequest{
		Project:        g.k8sProject,
		SecurityPolicy: "test",
	}

	//defer client.Close()

	resp, err := client.Get(ctx, req)
	if err != nil {
		klog.Error("\nError: ", err)
		os.Exit(1)
	}

	for _, singleRule := range resp.Rules {

		if *singleRule.Action != "allow" {

			for _, k := range singleRule.Match.Config.SrcIpRanges {
				for _, m := range ips {
					if k == m {
						found := find(prios, *singleRule.Priority)
						// check if prio already exists in array []prio
						// so the ip is in the rule, you can get the prio of this rule
						if !found {
							prios = append(prios, *singleRule.Priority)
						}

					}
				}
			}

		}

	}

	return prios

}

func find(slice []int32, val int32) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
