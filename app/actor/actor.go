package actor

import (
	"github.com/softonic/ip-blocker/app/actor/gcp"
)

type Actor struct {
	GCP gcp.GCPArmor
}

func NewActor() *Actor {
	return &Actor{
		GCP: gcp.GCPArmor{},
	}
}

func (a *Actor) InitConnectiontoActor() {

	a.GCP = gcp.GCPConnection()

}

func (a *Actor) GetIPsfromRulesActor() ([]string, int32) {

	ips, lastPriority := gcp.GetArmorRules(a.GCP)

	return ips, lastPriority

}

func (a *Actor) BlockIPsFromActor(prio int32, candidateIPsBlocked []string) (error, string) {

	err, status := gcp.BlockIPsFromArmor(a.GCP, prio, candidateIPsBlocked)

	return err, status

}
