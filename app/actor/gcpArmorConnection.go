package actor

import (
	"context"

	compute "cloud.google.com/go/compute/apiv1"
	"k8s.io/klog"
)

// GCPArmorConnection is a struct to connect to GCP Armor
type GCPArmorConnection struct {
	Client *compute.SecurityPoliciesClient
	Ctx    context.Context
}

func NewGCPArmorConnection() (*GCPArmorConnection, error) {
	ctx := context.Background()
	client, err := compute.NewSecurityPoliciesRESTClient(ctx)
	if err != nil {
		klog.Error("\nError: ", err)
		return nil, err
	}

	return &GCPArmorConnection{
		Client: client,
		Ctx:    ctx,
	}, nil
}
