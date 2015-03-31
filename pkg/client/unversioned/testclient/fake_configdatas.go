/*
Copyright 2015 The Kubernetes Authors All rights reserved.

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

package testclient

import (
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/watch"
)

const (
	configDataResourceName string = "configDatas"
)

// Fake implements ConfigDataInterface. Meant to be embedded into a struct to get a default
// implementation. This makes faking out just the method you want to test easier.
type FakeConfigDatas struct {
	Fake      *Fake
	Namespace string
}

func (c *FakeConfigDatas) Get(name string) (*extensions.ConfigData, error) {
	obj, err := c.Fake.Invokes(NewGetAction(configDataResourceName, c.Namespace, name), &extensions.ConfigData{})
	if obj == nil {
		return nil, err
	}

	return obj.(*extensions.ConfigData), err
}

func (c *FakeConfigDatas) List(label labels.Selector, field fields.Selector) (*extensions.ConfigDataList, error) {
	obj, err := c.Fake.Invokes(NewListAction(configDataResourceName, c.Namespace, label, field), &extensions.ConfigDataList{})
	if obj == nil {
		return nil, err
	}

	return obj.(*extensions.ConfigDataList), err
}

func (c *FakeConfigDatas) Create(cfg *extensions.ConfigData) (*extensions.ConfigData, error) {
	obj, err := c.Fake.Invokes(NewCreateAction(configDataResourceName, c.Namespace, cfg), cfg)
	if obj == nil {
		return nil, err
	}

	return obj.(*extensions.ConfigData), err
}

func (c *FakeConfigDatas) Update(cfg *extensions.ConfigData) (*extensions.ConfigData, error) {
	obj, err := c.Fake.Invokes(NewUpdateAction(configDataResourceName, c.Namespace, cfg), cfg)
	if obj == nil {
		return nil, err
	}

	return obj.(*extensions.ConfigData), err
}

func (c *FakeConfigDatas) UpdateStatus(cfg *extensions.ConfigData) (*extensions.ConfigData, error) {
	action := CreateActionImpl{}
	action.Verb = "update"
	action.Resource = configDataResourceName
	action.Subresource = "status"
	action.Object = cfg

	obj, err := c.Fake.Invokes(action, cfg)
	if obj == nil {
		return nil, err
	}

	return obj.(*extensions.ConfigData), err
}

func (c *FakeConfigDatas) Delete(name string) error {
	_, err := c.Fake.Invokes(NewDeleteAction(configDataResourceName, c.Namespace, name), &extensions.ConfigData{})
	return err
}

func (c *FakeConfigDatas) Watch(label labels.Selector, field fields.Selector, opts api.ListOptions) (watch.Interface, error) {
	return c.Fake.InvokesWatch(NewWatchAction(configDataResourceName, c.Namespace, label, field, opts))
}
