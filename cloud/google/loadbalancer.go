package google

import (
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	compute "google.golang.org/api/compute/v1"
	"fmt"
	"strings"
	"os"
)
//go-bindata -ignore=\\.go -ignore=\\.DS_Store -mode=0644 -modtime=1453795200 -o bindata.go -pkg etcd .
func (gce *GCEClient) CreateLoadbalancer(machine *clusterv1.Machine) error {
	config, err := gce.providerconfig(machine.Spec.ProviderConfig)
	if err != nil {
		return fmt.Errorf("Cannot unmarshal providerConfig field: %v", err)
	}

	zoneSplit := strings.Split(config.Zone, "-")
	region := strings.Join(zoneSplit[:2], "-")

	firewall := fmt.Sprintf("www-firewall-network-lb-%v", machine.ClusterName)
	tag := fmt.Sprintf("%v-lb-tag", machine.ClusterName)
	f, err := gce.service.Firewalls.Get(config.Project, firewall).Do()
	if f== nil {
		_, err = gce.service.Firewalls.Insert(config.Project, &compute.Firewall{
			Name:       firewall,
			TargetTags: []string{tag},
			Allowed: []*compute.FirewallAllowed{
				{
					IPProtocol: "tcp",
					Ports:      []string{"443"},
				},
			},
		}).Do()
	}
	if err != nil {
		return err
	}

	ipAddress := fmt.Sprintf("%v-lb-ip", machine.ClusterName)
	_, err = gce.service.Addresses.Get(config.Project, region, ipAddress).Do()
	if err != nil {
		_, err := gce.service.Addresses.Insert(config.Project, region, &compute.Address{
			Name:   ipAddress,
			Region: region,
		}).Do()
		if err != nil {
			return err
		}
	}

	addr, err := gce.service.Addresses.Get(config.Project, region, ipAddress).Do()
	if err != nil {
		return err
	}

	healthName := fmt.Sprintf("%v-health", machine.ClusterName)
	_, err  = gce.service.HttpHealthChecks.Get(config.Project, healthName).Do()
	if err != nil {
		_, err := gce.service.HttpHealthChecks.Insert(config.Project, &compute.HttpHealthCheck{
			Name:        healthName,
			Port:        80,
			RequestPath: "/",
			TimeoutSec: 1,
			UnhealthyThreshold: 5,
			HealthyThreshold: 1,
		}).Do()
		if err != nil {
			os.Exit(1)
		}
	}
	health, err  := gce.service.HttpHealthChecks.Get(config.Project, healthName).Do()

	poolName := fmt.Sprintf("%v-pool", machine.ClusterName)
	_, err = gce.service.TargetPools.Get(config.Project, region, poolName).Do()
	if err != nil {
		_, err = gce.service.TargetPools.Insert(config.Project, region, &compute.TargetPool{
			Name:   poolName,
			Region: region,
			HealthChecks: []string{health.SelfLink},
			SessionAffinity: "NONE",
		}).Do()
		if err != nil {
			return err
		}
	}

	pool, err := gce.service.TargetPools.Get(config.Project, region, poolName).Do()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
		return err
	}

	ruleName := fmt.Sprintf("%v-rule", machine.ClusterName)
	fr, err := gce.service.ForwardingRules.Get(config.Project, region, ruleName).Do()
	if fr == nil {
		op, err := gce.service.ForwardingRules.Insert(config.Project, region, &compute.ForwardingRule{
			Name:                ruleName,
			IPAddress:           addr.SelfLink,
			LoadBalancingScheme: "EXTERNAL",
			PortRange:           "443-443",
			Region:              region,
			Target:              pool.SelfLink,
		}).Do()
		fmt.Println(err, "................<><")
		if err != nil {
			return err
		}
		fmt.Println(op.Description)
		//return gce.waitForOperation(config, op)
	}
	return nil

}

func (gce *GCEClient)  addInstanceToPool(machine *clusterv1.Machine) error {
	config, err := gce.providerconfig(machine.Spec.ProviderConfig)
	if err != nil {
		return fmt.Errorf("Cannot unmarshal providerConfig field: %v", err)
	}
	zoneSplit := strings.Split(config.Zone, "-")
	region := strings.Join(zoneSplit[:2], "-")
	url := fmt.Sprintf(" https://www.googleapis.com/compute/v1/projects/%v/zones/%v/instances/%v", config.Project, config.Zone, machine.Name)
	clusterName := machine.ClusterName
	if clusterName == "" {
		clusterName = machine.Spec.Etcd.ClusterName
	}
	poolName := fmt.Sprintf("%v-pool", clusterName)
	_, err = gce.service.TargetPools.AddInstance(config.Project, region, poolName, &compute.TargetPoolsAddInstanceRequest{
		Instances: []*compute.InstanceReference{
			{
				Instance: url,
			},
		},
	}).Do()
	if err != nil {
		return err
	}
	return nil //gce.waitForOperation(config, op)
}

func (gce *GCEClient) getPublicIP(clusterName string, machine *clusterv1.Machine) (string, error) {
	config, err := gce.providerconfig(machine.Spec.ProviderConfig)
	if err != nil {
		return "", fmt.Errorf("Cannot unmarshal providerConfig field: %v", err)
	}
	zoneSplit := strings.Split(config.Zone, "-")
	region := strings.Join(zoneSplit[:2], "-")
	//addressUrl := fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%v/regions/%v/addresses/%v", config.Project, config.Zone, "network-lb-ip-ha")
	ipAddress := fmt.Sprintf("%v-lb-ip", clusterName)
	addr, err := gce.service.Addresses.Get(config.Project, region, ipAddress).Do()
	if err != nil {
		return "", err
	}
	fmt.Println(addr,"...........^^^^^^^^^^^^^^^^^^")
	return addr.Address, nil
}


