/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package fake

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
	cluster "sigs.k8s.io/cluster-api/pkg/apis/cluster"
)

// FakeMachineDeployments implements MachineDeploymentInterface
type FakeMachineDeployments struct {
	Fake *FakeCluster
	ns   string
}

var machinedeploymentsResource = schema.GroupVersionResource{Group: "cluster.k8s.io", Version: "", Resource: "machinedeployments"}

var machinedeploymentsKind = schema.GroupVersionKind{Group: "cluster.k8s.io", Version: "", Kind: "MachineDeployment"}

// Get takes name of the machineDeployment, and returns the corresponding machineDeployment object, and an error if there is any.
func (c *FakeMachineDeployments) Get(name string, options v1.GetOptions) (result *cluster.MachineDeployment, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(machinedeploymentsResource, c.ns, name), &cluster.MachineDeployment{})

	if obj == nil {
		return nil, err
	}
	return obj.(*cluster.MachineDeployment), err
}

// List takes label and field selectors, and returns the list of MachineDeployments that match those selectors.
func (c *FakeMachineDeployments) List(opts v1.ListOptions) (result *cluster.MachineDeploymentList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(machinedeploymentsResource, machinedeploymentsKind, c.ns, opts), &cluster.MachineDeploymentList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &cluster.MachineDeploymentList{}
	for _, item := range obj.(*cluster.MachineDeploymentList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested machineDeployments.
func (c *FakeMachineDeployments) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(machinedeploymentsResource, c.ns, opts))

}

// Create takes the representation of a machineDeployment and creates it.  Returns the server's representation of the machineDeployment, and an error, if there is any.
func (c *FakeMachineDeployments) Create(machineDeployment *cluster.MachineDeployment) (result *cluster.MachineDeployment, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(machinedeploymentsResource, c.ns, machineDeployment), &cluster.MachineDeployment{})

	if obj == nil {
		return nil, err
	}
	return obj.(*cluster.MachineDeployment), err
}

// Update takes the representation of a machineDeployment and updates it. Returns the server's representation of the machineDeployment, and an error, if there is any.
func (c *FakeMachineDeployments) Update(machineDeployment *cluster.MachineDeployment) (result *cluster.MachineDeployment, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(machinedeploymentsResource, c.ns, machineDeployment), &cluster.MachineDeployment{})

	if obj == nil {
		return nil, err
	}
	return obj.(*cluster.MachineDeployment), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeMachineDeployments) UpdateStatus(machineDeployment *cluster.MachineDeployment) (*cluster.MachineDeployment, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(machinedeploymentsResource, "status", c.ns, machineDeployment), &cluster.MachineDeployment{})

	if obj == nil {
		return nil, err
	}
	return obj.(*cluster.MachineDeployment), err
}

// Delete takes name of the machineDeployment and deletes it. Returns an error if one occurs.
func (c *FakeMachineDeployments) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(machinedeploymentsResource, c.ns, name), &cluster.MachineDeployment{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeMachineDeployments) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(machinedeploymentsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &cluster.MachineDeploymentList{})
	return err
}

// Patch applies the patch and returns the patched machineDeployment.
func (c *FakeMachineDeployments) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *cluster.MachineDeployment, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(machinedeploymentsResource, c.ns, name, data, subresources...), &cluster.MachineDeployment{})

	if obj == nil {
		return nil, err
	}
	return obj.(*cluster.MachineDeployment), err
}
