package actor

import (
	"fmt"
	"strings"

	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
	"k8s.io/klog"
)

// Build the query to Add a Rule to the Security Policy
func buildArmorQueryAddRule(blockStringArray []string, project string, policy string, action string, description string, priority int32, preview bool) *computepb.AddRuleSecurityPolicyRequest {

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

// Execute the query to add a rule to the armor policy
func (g *GCPArmorActor) executeArmorQueryAddRule(ips []string, priority int32, action, description string, preview bool) error {

	client := g.Connection.Client
	ctx := g.Connection.Ctx

	req := buildArmorQueryAddRule(ips, g.ActorConfig.Project, g.ActorConfig.Policy, action, description, priority, preview)

	_, err := client.AddRule(ctx, req)
	if err != nil {
		if strings.Contains(err.Error(), "Cannot have rules with the same priorities") {
			// reexecute the query with prio + 1 if the prio is already used
			priority++
			req := buildArmorQueryAddRule(ips, g.ActorConfig.Project, g.ActorConfig.Policy, action, description, priority, preview)
			_, err := client.AddRule(ctx, req)
			if err != nil {
				klog.Error("\nError: ", err)
				return err
			}
		} else {
			klog.Error("\nError: ", err)
			return err
		}
	}

	if err != nil {
		return fmt.Errorf("error executing add armor query: %v", err)
	}

	return nil
}

// Execute the query to remove a rule from the armor policy
func (g *GCPArmorActor) executeArmorQueryRemoveRule(priority int32) error {
	client := g.Connection.Client
	ctx := g.Connection.Ctx

	req := &computepb.RemoveRuleSecurityPolicyRequest{
		Project:        g.ActorConfig.Project,
		SecurityPolicy: g.ActorConfig.Policy,
		Priority:       &priority,
	}

	_, err := client.RemoveRule(ctx, req)

	if err != nil {
		return fmt.Errorf("error executing remove armor query: %v", err)
	}

	return nil
}
