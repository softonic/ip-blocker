package actor

import (
	"context"
	"errors"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"fmt"

	"k8s.io/klog"

	"gopkg.in/yaml.v2"

	compute "cloud.google.com/go/compute/apiv1"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"

	"github.com/softonic/ip-blocker/app"
)

type GCPArmorActor struct {
	client     *compute.SecurityPoliciesClient
	ctx        context.Context
	k8sProject string
	policy     string
	ttlRules   int
	excludeIPs string
}

type GCPArmorConf struct {
	preview bool   `yaml:"preview"`
	action  string `yaml:"action"`
}

func NewGCPArmorActor(project string, policy string, ttlRules int, excludeIPs string) *GCPArmorActor {

	c, ctx := InitConnectiontoActor()

	return &GCPArmorActor{
		client:     c,
		ctx:        ctx,
		k8sProject: project,
		policy:     policy,
		ttlRules:   ttlRules,
		excludeIPs: excludeIPs,
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

func (c *GCPArmorConf) getConf() *GCPArmorConf {

	yamlFile, err := ioutil.ReadFile("/etc/config/gcp-armor-config.yaml")
	if err != nil {
		klog.Errorf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		klog.Fatalf("Unmarshal: %v", err)
	}

	return c
}

func getIPsAlreadyBlockedFromRules(g *GCPArmorActor, securityPolicy string) ([]string, int32) {

	client := g.client
	ctx := g.ctx

	req := &computepb.GetSecurityPolicyRequest{
		Project:        g.k8sProject,
		SecurityPolicy: g.policy,
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

		match := false
		match, _ = regexp.MatchString("Google suggested rule for attack ID", *singleRule.Description)

		if match {
			continue
		}

		if *singleRule.Action != "allow" && *singleRule.Match.VersionedExpr == 70925961 {

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

func buildQueryObjectArmor(blockStringArray []string, project string, policy string, action string, description string, priority int32, preview bool) *computepb.AddRuleSecurityPolicyRequest {

	versioned := computepb.SecurityPolicyRuleMatcher_SRC_IPS_V1.Enum()

	match := &computepb.SecurityPolicyRuleMatcher{
		Config: &computepb.SecurityPolicyRuleMatcherConfig{
			SrcIpRanges: blockStringArray,
		},
		VersionedExpr: versioned,
	}

	req := &computepb.AddRuleSecurityPolicyRequest{

		Project:        project,
		SecurityPolicy: policy,
		SecurityPolicyRuleResource: &computepb.SecurityPolicyRule{
			Action:      &(action),
			Description: &description,
			Priority:    &priority,
			Preview:     &(preview),
			Match:       match,
		},
	}

	return req

}

func executeQueryArmor(client compute.SecurityPoliciesClient, req *computepb.AddRuleSecurityPolicyRequest, ctx context.Context) error {

	resp, err := client.AddRule(ctx, req)
	if err != nil {
		klog.Error("\nError: ", err)
		return err
	}

	_ = resp

	return nil

}

func (g *GCPArmorActor) BlockIPs(sourceIPs []app.IPCount) error {

	client := g.client
	ctx := g.ctx
	project := g.k8sProject

	var sourceIPstring, candidateWithCird []string

	for _, k := range sourceIPs {
		sourceIPstring = append(sourceIPstring, k.IP)
	}

	data := make(map[interface{}]interface{})

	yamlFile, err := ioutil.ReadFile("/etc/config/gcp-armor-config.yaml")
	if err != nil {
		klog.Errorf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, &data)
	if err != nil {
		klog.Fatalf("Unmarshal: %v", err)
	}

	//defer client.Close()

	actorIPs, lastprio := getIPsAlreadyBlockedFromRules(g, g.policy)

	excludedIpsWellFormated := formatIpsfromStringtoArray(g.excludeIPs)

	klog.Infof("Excluded IPs: %v", excludedIpsWellFormated)

	candidateIPstoBlock := uniqueItems(sourceIPstring, actorIPs)

	candidateAfterExcluded := uniqueItems(candidateIPstoBlock, excludedIpsWellFormated)

	candidateAfterExcluded = removeDuplicateStr(candidateAfterExcluded)

	for _, k := range candidateAfterExcluded {
		candidateWithCird = append(candidateWithCird, k+"/32")
	}

	now := time.Now()
	secs := now.Unix()

	description := strconv.FormatInt(secs, 10)
	priority := lastprio + 1

	action := fmt.Sprintf("%v", data["action"])
	preview, _ := strconv.ParseBool(fmt.Sprintf("%v", data["preview"]))

	if len(candidateIPstoBlock) > 0 {

		if len(candidateWithCird) > 10 {

			var j int
			for i := 0; i < len(candidateWithCird); i += 10 {
				j += 10
				if j > len(candidateWithCird) {
					j = len(candidateWithCird)
				}
				// do what do you want to with the sub-slice
				fmt.Println(candidateWithCird[i:j])
				req := buildQueryObjectArmor(candidateWithCird[i:j], project, g.policy, action, description, priority, preview)
				err := executeQueryArmor(*client, req, ctx)
				if err != nil {
					klog.Error("\nError: ", err)
					return err
				} else {
					klog.Infof("Adding rule with prio: %d", priority)
					klog.Infof("Blocked IPs: %v", candidateWithCird[i:j])
				}
			}

		} else {
			req := buildQueryObjectArmor(candidateWithCird, project, g.policy, action, description, priority, preview)
			err := executeQueryArmor(*client, req, ctx)
			if err != nil {
				klog.Error("\nError: ", err)
				return err
			} else {
				klog.Infof("Adding rule with prio: %d", priority)
				klog.Infof("Blocked IPs: %v", candidateWithCird)
			}
		}

		return nil

	}

	return nil

}

func removeDuplicateStr(strSlice []string) []string {
	allKeys := make(map[string]bool)
	list := []string{}
	for _, item := range strSlice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

func uniqueItems(sourceIPs []string, exceptionsIPs []string) []string {

	var ipWithMaskES string
	candidateIPsBlocked := []string{}

	for _, elasticIps := range sourceIPs {
		count := 0
		for _, armorIps := range exceptionsIPs {
			ipWithMaskES = elasticIps
			if ipWithMaskES == armorIps || ipWithMaskES == armorIps+"/32" || ipWithMaskES+"/32" == armorIps {
				count++
			}
		}
		if count == 0 {
			candidateIPsBlocked = append(candidateIPsBlocked, elasticIps)
		}

	}

	return candidateIPsBlocked

}

// GetBlockedIPsFromActorThatCanBeUnblocked: return IPs that has been blocked for more than ttlRules min
func getBlockedIPsFromActorThatCanBeUnblocked(g *GCPArmorActor) []string {

	client := g.client
	ctx := g.ctx
	ttlRules := g.ttlRules

	req := &computepb.GetSecurityPolicyRequest{
		Project:        g.k8sProject,
		SecurityPolicy: g.policy,
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

		match := false
		match, _ = regexp.MatchString("Google suggested rule for attack ID", *singleRule.Description)

		if match {
			continue
		}

		// && singleRule.Match.VersionedExpr == computepb.SecurityPolicyRuleMatcher_SRC_IPS_V1.Enum()

		if *singleRule.Action != "allow" && *singleRule.Match.VersionedExpr == 70925961 {

			n, err := strconv.ParseInt(*singleRule.Description, 10, 64)
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
			return errors.New("there are no rules in this policy")
		}

		req := &computepb.RemoveRuleSecurityPolicyRequest{

			Project:        project,
			SecurityPolicy: g.policy,
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
		SecurityPolicy: g.policy,
	}

	//defer client.Close()

	resp, err := client.Get(ctx, req)
	if err != nil {
		klog.Error("\nError: ", err)
		os.Exit(1)
	}

	for _, singleRule := range resp.Rules {

		match := false
		match, _ = regexp.MatchString("Google suggested rule for attack ID", *singleRule.Description)

		if match {
			continue
		}

		if *singleRule.Action != "allow" && *singleRule.Match.VersionedExpr == 70925961 {

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

func formatIpsfromStringtoArray(excludeIps string) []string {
	return strings.Split(excludeIps, ",")
}
