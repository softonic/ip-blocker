package actor

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

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
		SecurityPolicy: "test",
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

	fmt.Println("These are the IPs already banned in armor that return from func getIPsAlreadyBlockedFromRules:", ips)

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
			fmt.Println("\nError: ", err)
			return err
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

	fmt.Println("These are the IPs from source", sourceIPs)
	fmt.Println("These are the IPs from Armor", actorIPs)

	for _, elasticIps := range sourceIPs {
		count := 0
		for _, armorIps := range actorIPs {
			ipWithMaskES = elasticIps.IP + "/32"
			if ipWithMaskES == armorIps {
				fmt.Println("this IPs are already in the armor", ipWithMaskES)
				count++
			}
		}
		if count == 0 {
			fmt.Println("this IP is not in armor:", elasticIps.IP+"/32")
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
		fmt.Println("\nError: ", err)
		os.Exit(1)
	}

	now := time.Now()
	secs := now.Unix()

	var ips, restIps []string

	for _, singleRule := range resp.Rules {

		if *singleRule.Action != "allow" {

			n, err := strconv.ParseInt(*singleRule.Description, 10, 64)
			fmt.Println("this rule has a epoch of:", n)
			if err != nil {
				fmt.Println("this is not a description")
				continue
			}
			if (secs - n) > 301 {
				fmt.Println("Its been more than 1 day, we can unban this")

				restIps = singleRule.Match.Config.SrcIpRanges

				ips = append(ips, restIps...)
			} else {
				fmt.Println("Still valid, dont unblock")
			}

		}

	}
	fmt.Println("These are the IPs I should ban", ips)

	return ips

}

func (g *GCPArmorActor) UnBlockIPs() error {

	client := g.client
	ctx := g.ctx
	project := g.k8sProject

	ips := getBlockedIPsFromActorThatCanBeUnblocked(g)

	prios := getRuleFromIP(g, ips)

	fmt.Println("the prios gotten from getRuleFromIP", prios)

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

		fmt.Println("we are unblocking the following prios/rules", prio)

		resp, err := client.RemoveRule(ctx, req)
		if err != nil {
			fmt.Println("\nError: ", err)
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
		fmt.Println("\nError: ", err)
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
