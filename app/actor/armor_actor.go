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

func (g *GCPArmorActor) BlockIPs(sourceIPs []app.IPCount) error {

	var sourceIPstring []string

	for _, k := range sourceIPs {
		sourceIPstring = append(sourceIPstring, k.IP)
	}

	RulesGetter := NewRulesGetter(g)
	rules, err := RulesGetter.GetSecurityRules()
	if err != nil {
		klog.Error("\nError: ", err)
		return err
	}

	ipGetter := NewIPGetter(g)
	alreadyBlockedIPs, err := ipGetter.GetBlockedIPs(rules)
	if err != nil {
		klog.Error("\nError: ", err)
		return err
	}

	lastprio := getLastPriority(rules)

	excludedIPsinArray, err := utils.ConvertCSVToArray(g.ActorConfig.ExcludeIPs)
	if err != nil {
		klog.Error("\nError with exclude IPs function: ", err)
		return err
	}

	handler := utils.UtilsIPListHandler{}
	candidateIPsToBlock := getCandidateIPsToBlock(handler, sourceIPstring, alreadyBlockedIPs, excludedIPsinArray)

	if len(candidateIPsToBlock) == 0 {
		return nil
	}

	now := time.Now()
	secs := now.Unix()

	description := "ipblocker:" + strconv.FormatInt(secs, 10)
	priority := lastprio + 1

	action := fmt.Sprintf("%v", conf.Data["action"])
	preview, _ := strconv.ParseBool(fmt.Sprintf("%v", conf.Data["preview"]))

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
