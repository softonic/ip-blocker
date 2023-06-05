package actor

import (
	"errors"
	"strconv"
	"time"

	"fmt"

	"k8s.io/klog"

	"github.com/softonic/ip-blocker/app"
	"github.com/softonic/ip-blocker/app/actor/utils"
)

type GCPArmorActor struct {
	Connection  *GCPArmorConnection // client to connect to GCPArmor
	RulesConfig *GCPArmorRulesConf  // rules to execute in GCPArmor
	ActorConfig *ActorConfig        // actor configuration (excludeIPs, preview, action)
}

func NewGCPArmorActor(config *ActorConfig) (*GCPArmorActor, error) {

	connection, err := NewGCPArmorConnection()
	if err != nil {
		return nil, err
	}

	RulesConfig := &GCPArmorRulesConf{}
	err = RulesConfig.LoadConfig()
	if err != nil {
		return nil, err
	}

	return &GCPArmorActor{
		Connection:  connection,
		RulesConfig: RulesConfig,
		ActorConfig: config,
	}, nil
}

// GetIPsToBlock: this function will get the IPs candidate to Block by BlockIPs function
func (g *GCPArmorActor) getBlockIPs(sourceIPs []app.IPCount) ([]string, int32, error) {

	var sourceIPstring []string

	for _, k := range sourceIPs {
		sourceIPstring = append(sourceIPstring, k.IP)
	}

	RulesGetter := NewRulesGetter(g)
	rules, err := RulesGetter.GetSecurityRules()
	if err != nil {
		klog.Error("\nError: ", err)
		return nil, 0, err
	}

	ipGetter := NewIPGetter(g)
	alreadyBlockedIPs, err := ipGetter.GetBlockedIPs(rules)
	if err != nil {
		klog.Error("\nError: ", err)
		return nil, 0, err
	}

	lastprio := getLastPriority(rules)

	excludedIPsinArray, err := utils.ConvertCSVToArray(g.ActorConfig.ExcludeIPs)
	if err != nil {
		klog.Error("\nError with exclude IPs function: ", err)
		return nil, 0, err
	}

	handler := utils.UtilsIPListHandler{}
	candidateIPsToBlock := getCandidateIPsToBlock(handler, sourceIPstring, alreadyBlockedIPs, excludedIPsinArray)

	if len(candidateIPsToBlock) == 0 {
		return nil, 0, nil
	}

	return candidateIPsToBlock, lastprio, nil

}

// BlockIPs: this function will block the IPs
func (g *GCPArmorActor) BlockIPs(sourceIPs []app.IPCount) error {

	candidateIPsToBlock, lastprio, err := g.getBlockIPs(sourceIPs)
	if err != nil {
		return err
	}

	description := setDescriptionForNewRules()

	priority := lastprio + 1

	action := fmt.Sprintf("%v", conf.Data["action"])
	preview, _ := strconv.ParseBool(fmt.Sprintf("%v", conf.Data["preview"]))

	err = addNewFirewallRules(g, candidateIPsToBlock, priority, action, description, preview)
	if err != nil {
		return err
	}

	return nil

}

// this function will compute the timestamp to be used in the description of the new rules
func setDescriptionForNewRules() string {

	now := time.Now()
	secs := now.Unix()

	description := "ipblocker:" + strconv.FormatInt(secs, 10)

	return description

}

// this function will add the new rules to GCPArmor
func addNewFirewallRules(g *GCPArmorActor, candidateIPsToBlock []string, priority int32, action string, description string, preview bool) error {

	chunkSize := 10
	for i := 0; i < len(candidateIPsToBlock); i += chunkSize {
		end := i + chunkSize
		if end > len(candidateIPsToBlock) {
			end = len(candidateIPsToBlock)
		}

		ipsChunk := candidateIPsToBlock[i:end]

		err := g.executeArmorQueryAddRule(ipsChunk, priority, action, description, preview)

		if err != nil {
			return err
		}
		klog.Infof("Adding rule with priority: %d", priority)
		klog.Infof("Blocked IPs: %v", ipsChunk)
	}

	return nil
}

func getFieldFromData(confData string) interface{} {

	return fmt.Sprintf("%v", conf.Data[confData])

}

func (g *GCPArmorActor) UnBlockIPs() error {

	RulesGetter := NewRulesGetter(g)
	// Return all the security rules
	rules, err := RulesGetter.GetSecurityRules()
	if err != nil {
		klog.Error("\nError: ", err)
		return err
	}

	// Get the IPs that can be unblocked depending on the ttl.
	// Rules passed by argument are already the list of ipBlocker rules
	ips := getBlockedIPsFromActorThatCanBeUnblocked(rules, g.ActorConfig.TTLRules)

	// Get the priorities of the rules that will be removed.
	// rules passed by argument are already the list of ipBlocker rules
	prios, err := GetRuleFromIP(rules, ips)

	for _, prio := range prios {

		if prio == 0 {
			return errors.New("there are no rules in this policy")
		}

		err := g.executeArmorQueryRemoveRule(prio)
		if err != nil {
			klog.Error("\nError: ", err)
			return err
		}
	}

	return nil

}
