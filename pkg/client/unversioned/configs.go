/*
Copyright 2015 Google Inc. All rights reserved.

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
	"fmt"

	"k8s.io/kubernetes/pkg/apis/experimental"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/watch"
)

type ConfigDatasNamespacer interface {
	ConfigDatas(namespace string) ConfigDatasInterface
}

type ConfigDatasInterface interface {
	Get(string) (*experimental.ConfigData, error)
	List(labels.Selector, fields.Selector) (*experimental.ConfigDataList, error)
	Create(*experimental.ConfigData) (*experimental.ConfigData, error)
	Delete(string) error
	Update(*experimental.ConfigData) (*experimental.ConfigData, error)
	UpdateStatus(*experimental.ConfigData) (*experimental.ConfigData, error)
	Watch(labels.Selector, fields.Selector, string) (watch.Interface, error)
}

type configDatas struct {
	client    *Client
	namespace string
}

func newConfigDatas(c *Client, ns string) *configDatas {
	return &configDatas{
		client:    c,
		namespace: ns,
	}
}

func (cfg *configDatas) resourceName() string {
	return "configDatas"
}

func (cfg *configDatas) Get(name string) (*experimental.ConfigData, error) {
	result := &experimental.ConfigData{}
	err := cfg.client.Get().
		Namespace(cfg.namespace).
		Resource(cfg.resourceName()).
		Name(name).
		Do().
		Into(result)

	return result, err
}

func (cfg *configDatas) List(label labels.Selector, field fields.Selector) (*experimental.ConfigDataList, error) {
	result := &experimental.ConfigDataList{}
	err := cfg.client.Get().
		Namespace(cfg.namespace).
		Resource(cfg.resourceName()).
		LabelsSelectorParam(label).
		FieldsSelectorParam(field).
		Do().
		Into(result)

	return result, err
}

func (cfg *configDatas) Create(config *experimental.ConfigData) (*experimental.ConfigData, error) {
	result := &experimental.ConfigData{}
	err := cfg.client.Post().
		Namespace(cfg.namespace).
		Resource(cfg.resourceName()).
		Body(config).
		Do().
		Into(result)

	return result, err
}

func (cfg *configDatas) Delete(name string) error {
	return cfg.client.Delete().
		Namespace(cfg.namespace).
		Resource(cfg.resourceName()).
		Name(name).
		Do().
		Error()
}

func (oldCfg *configDatas) Update(newCfg *experimental.ConfigData) (*experimental.ConfigData, error) {
	result := &experimental.ConfigData{}
	if len(newCfg.ResourceVersion) == 0 {
		err := fmt.Errorf("invalid update object, missing resource version: %v", newCfg)
		return nil, err
	}
	err := oldCfg.client.Put().
		Namespace(oldCfg.namespace).
		Resource(oldCfg.resourceName()).
		Name(newCfg.Name).
		Body(newCfg).
		Do().
		Into(result)

	return result, err
}

func (oldCfg *configDatas) UpdateStatus(newCfg *experimental.ConfigData) (*experimental.ConfigData, error) {
	result := &experimental.ConfigData{}
	if len(newCfg.ResourceVersion) == 0 {
		err := fmt.Errorf("invalid update object, missing resource version: %v", newCfg)
		return nil, err
	}
	err := oldCfg.client.Put().
		Namespace(oldCfg.namespace).
		Resource(oldCfg.resourceName()).
		Name(newCfg.Name).
		SubResource("status").
		Body(newCfg).
		Do().
		Into(result)

	return result, err
}

func (cfg *configDatas) Watch(label labels.Selector, field fields.Selector, resourceVersion string) (watch.Interface, error) {
	return cfg.client.Get().
		Prefix("watch").
		Namespace(cfg.namespace).
		Resource(cfg.resourceName()).
		Param("resourceVersion", resourceVersion).
		LabelsSelectorParam(label).
		FieldsSelectorParam(field).
		Watch()
}
