package digitalocean

import (
	goctx "context"
	"errors"
	"fmt"
	"time"

	"k8s.io/api/core/v1"

	"github.com/digitalocean/godo"
	"github.com/digitalocean/godo/context"
)

const (
	// reach the active state.
	defaultActiveTimeout = 90

	// defaultActiveCheckTick is the number of seconds between load balancer
	// status checks when waiting for activation.
	defaultActiveCheckTick = 5

	// statuses for Digital Ocean load balancer
	lbStatusNew     = "new"
	lbStatusActive  = "active"
	lbStatusErrored = "errored"
)

var errLBNotFound = errors.New("loadbalancer not found")



// EnsureLoadBalancer ensures that the cluster is running a load balancer for
// service.
//
// EnsureLoadBalancer will not modify service or nodes.
func (conn *cloudConnector) EnsureLoadBalancer(clusterName, region string) (*v1.LoadBalancerStatus, error) {
	lbName := fmt.Sprintf("%v-lb", clusterName)
	lbStatus, exists, err := conn.GetLoadBalancer(clusterName)
	if err != nil {
		return nil, err
	}


	if !exists {
		lbRequest := &godo.LoadBalancerRequest{
			Name:            lbName,
			Region:          region,
			ForwardingRules: []godo.ForwardingRule{
				{
					EntryProtocol: "tcp",
					EntryPort: 443,
					TargetPort: 443,
					TargetProtocol: "tcp",

				},
			},
			HealthCheck:      &godo.HealthCheck{
				Protocol:               "TCP",
				Port:                   int(80),
				CheckIntervalSeconds:   3,
				ResponseTimeoutSeconds: 5,
				HealthyThreshold:       5,
				UnhealthyThreshold:     3,
			},
			StickySessions:  &godo.StickySessions{
				Type: "none",
			},
			Algorithm:       "least_connections",
		}

		lb, _, err := conn.client.LoadBalancers.Create(context.TODO(), lbRequest)
		if err != nil {
			return nil, err
		}

		lb, err = conn.waitActive(lb.ID)
		if err != nil {
			return nil, err
		}

		return &v1.LoadBalancerStatus{
			Ingress: []v1.LoadBalancerIngress{
				{
					IP: lb.IP,
				},
			},
		}, nil
	}

	return lbStatus, nil

}


func (conn *cloudConnector) waitActive(lbID string) (*godo.LoadBalancer, error) {
	ctx, cancel := goctx.WithTimeout(goctx.TODO(), time.Second*time.Duration(defaultActiveTimeout))
	defer cancel()
	ticker := time.NewTicker(time.Second * time.Duration(defaultActiveCheckTick))
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			lb, _, err := conn.client.LoadBalancers.Get(ctx, lbID)
			if err != nil {
				return nil, err
			}

			if lb.Status == lbStatusActive {
				return lb, nil
			}
			if lb.Status == lbStatusErrored {
				return nil, fmt.Errorf("error creating DigitalOcean balancer: %q", lbID)
			}
		case <-ctx.Done():
			return nil, fmt.Errorf("load balancer creation for %q timed out", lbID)
		}
	}
}

func (conn *cloudConnector) GetLoadBalancer(clusterName string) (*v1.LoadBalancerStatus, bool, error) {
	lbName := fmt.Sprintf("%v-lb", clusterName)
	lb, err := conn.lbByName(context.TODO(), lbName)
	if err != nil {
		if err == errLBNotFound {
			return nil, false, nil
		}

		return nil, false, err
	}

	if lb.Status != lbStatusActive {
		lb, err = conn.waitActive(lb.ID)
		if err != nil {
			return nil, true, fmt.Errorf("error waiting for load balancer to be active %v", err)
		}
	}

	return &v1.LoadBalancerStatus{
		Ingress: []v1.LoadBalancerIngress{
			{
				IP: lb.IP,
			},
		},
	}, true, nil
}

func (conn *cloudConnector)  lbByName(ctx context.Context, name string) (*godo.LoadBalancer, error) {
	lbs, _, err := conn.client.LoadBalancers.List(ctx, &godo.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, lb := range lbs {
		if lb.Name == name {
			return &lb, nil
		}
	}

	return nil, errLBNotFound
}

func (conn *cloudConnector) addInstanceToPool(ctx context.Context, clusterName string, id int) error {
	lbName := fmt.Sprintf("%v-lb", clusterName)
	lb, err := conn.lbByName(ctx, lbName)
	if err != nil {
		return err
	}
	_, err = conn.client.LoadBalancers.AddDroplets(ctx, lb.ID, id)
	if err != nil {
		return err
	}
	return nil

}