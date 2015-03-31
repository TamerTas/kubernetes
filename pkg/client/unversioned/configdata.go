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

package unversioned

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

type ConfigDatasNamespacer interface {
	ConfigDatas(namespace string) ConfigDatasInterface
}

type ConfigDatasInterface interface {
	Get(string) (*extensions.ConfigData, error)
	List(labels.Selector, fields.Selector) (*extensions.ConfigDataList, error)
	Create(*extensions.ConfigData) (*extensions.ConfigData, error)
	Delete(string) error
	Update(*extensions.ConfigData) (*extensions.ConfigData, error)
	UpdateStatus(*extensions.ConfigData) (*extensions.ConfigData, error)
	Watch(labels.Selector, fields.Selector, api.ListOptions) (watch.Interface, error)
}

type configDatas struct {
	client    *Client
	namespace string
}

// configDatas should implement ConfigDatasInterface
var _ ConfigDatasInterface = &configDatas{}

func newConfigDatas(c *Client, ns string) *configDatas {
	return &configDatas{
		client:    c,
		namespace: ns,
	}
}

func (c *configDatas) Get(name string) (*extensions.ConfigData, error) {
	result := &extensions.ConfigData{}
	err := c.client.Get().
		Namespace(c.namespace).
		Resource(configDataResourceName).
		Name(name).
		Do().
		Into(result)

	return result, err
}

func (c *configDatas) List(label labels.Selector, field fields.Selector) (*extensions.ConfigDataList, error) {
	result := &extensions.ConfigDataList{}
	err := c.client.Get().
		Namespace(c.namespace).
		Resource(configDataResourceName).
		LabelsSelectorParam(label).
		FieldsSelectorParam(field).
		Do().
		Into(result)

	return result, err
}

func (c *configDatas) Create(cfg *extensions.ConfigData) (*extensions.ConfigData, error) {
	result := &extensions.ConfigData{}
	err := c.client.Post().
		Namespace(c.namespace).
		Resource(configDataResourceName).
		Body(cfg).
		Do().
		Into(result)

	return result, err
}

func (c *configDatas) Delete(name string) error {
	return c.client.Delete().
		Namespace(c.namespace).
		Resource(configDataResourceName).
		Name(name).
		Do().
		Error()
}

func (c *configDatas) Update(cfg *extensions.ConfigData) (*extensions.ConfigData, error) {
	result := &extensions.ConfigData{}

	err := c.client.Put().
		Namespace(c.namespace).
		Resource(configDataResourceName).
		Name(cfg.Name).
		Body(cfg).
		Do().
		Into(result)

	return result, err
}

func (c *configDatas) UpdateStatus(cfg *extensions.ConfigData) (*extensions.ConfigData, error) {
	result := &extensions.ConfigData{}

	err := c.client.Put().
		Namespace(c.namespace).
		Resource(configDataResourceName).
		Name(cfg.Name).
		SubResource("status").
		Body(cfg).
		Do().
		Into(result)

	return result, err
}

func (c *configDatas) Watch(label labels.Selector, field fields.Selector, opts api.ListOptions) (watch.Interface, error) {
	return c.client.Get().
		Prefix("watch").
		Namespace(c.namespace).
		Resource(configDataResourceName).
		VersionedParams(&opts, api.Scheme).
		LabelsSelectorParam(label).
		FieldsSelectorParam(field).
		Watch()
}
