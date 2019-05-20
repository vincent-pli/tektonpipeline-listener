/*
Copyright The Kubernetes Authors.

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

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1alpha1 "github.com/vincent-pli/tektonpipeline-listener/pkg/apis/samplecontroller/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeListenerTemplates implements ListenerTemplateInterface
type FakeListenerTemplates struct {
	Fake *FakeSamplecontrollerV1alpha1
	ns   string
}

var listenertemplatesResource = schema.GroupVersionResource{Group: "samplecontroller.k8s.io", Version: "v1alpha1", Resource: "listenertemplates"}

var listenertemplatesKind = schema.GroupVersionKind{Group: "samplecontroller.k8s.io", Version: "v1alpha1", Kind: "ListenerTemplate"}

// Get takes name of the listenerTemplate, and returns the corresponding listenerTemplate object, and an error if there is any.
func (c *FakeListenerTemplates) Get(name string, options v1.GetOptions) (result *v1alpha1.ListenerTemplate, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(listenertemplatesResource, c.ns, name), &v1alpha1.ListenerTemplate{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ListenerTemplate), err
}

// List takes label and field selectors, and returns the list of ListenerTemplates that match those selectors.
func (c *FakeListenerTemplates) List(opts v1.ListOptions) (result *v1alpha1.ListenerTemplateList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(listenertemplatesResource, listenertemplatesKind, c.ns, opts), &v1alpha1.ListenerTemplateList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.ListenerTemplateList{ListMeta: obj.(*v1alpha1.ListenerTemplateList).ListMeta}
	for _, item := range obj.(*v1alpha1.ListenerTemplateList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested listenerTemplates.
func (c *FakeListenerTemplates) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(listenertemplatesResource, c.ns, opts))

}

// Create takes the representation of a listenerTemplate and creates it.  Returns the server's representation of the listenerTemplate, and an error, if there is any.
func (c *FakeListenerTemplates) Create(listenerTemplate *v1alpha1.ListenerTemplate) (result *v1alpha1.ListenerTemplate, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(listenertemplatesResource, c.ns, listenerTemplate), &v1alpha1.ListenerTemplate{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ListenerTemplate), err
}

// Update takes the representation of a listenerTemplate and updates it. Returns the server's representation of the listenerTemplate, and an error, if there is any.
func (c *FakeListenerTemplates) Update(listenerTemplate *v1alpha1.ListenerTemplate) (result *v1alpha1.ListenerTemplate, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(listenertemplatesResource, c.ns, listenerTemplate), &v1alpha1.ListenerTemplate{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ListenerTemplate), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeListenerTemplates) UpdateStatus(listenerTemplate *v1alpha1.ListenerTemplate) (*v1alpha1.ListenerTemplate, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(listenertemplatesResource, "status", c.ns, listenerTemplate), &v1alpha1.ListenerTemplate{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ListenerTemplate), err
}

// Delete takes name of the listenerTemplate and deletes it. Returns an error if one occurs.
func (c *FakeListenerTemplates) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(listenertemplatesResource, c.ns, name), &v1alpha1.ListenerTemplate{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeListenerTemplates) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(listenertemplatesResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.ListenerTemplateList{})
	return err
}

// Patch applies the patch and returns the patched listenerTemplate.
func (c *FakeListenerTemplates) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.ListenerTemplate, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(listenertemplatesResource, c.ns, name, pt, data, subresources...), &v1alpha1.ListenerTemplate{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ListenerTemplate), err
}
