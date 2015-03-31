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

package configdata

import (
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/rest"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/watch"
)

// Registry is an interface for things that know how to store ConfigDatas.
type Registry interface {
	// ListConfigDatas obtains a list of ConfigDatas matching the given api.ListOptions.
	ListConfigDatas(ctx api.Context, options *api.ListOptions) (*extensions.ConfigDataList, error)
	// WatchConfigDatas watch for new/changed/deleted ConfigDatas.
	WatchConfigDatas(ctx api.Context, options *api.ListOptions) (watch.Interface, error)
	// GetConfigDatas gets a specific ConfigData.
	GetConfigData(ctx api.Context, name string) (*extensions.ConfigData, error)
	// CreateConfigData creates a ConfigData based on a specification.
	CreateConfigData(ctx api.Context, cfg *extensions.ConfigData) (*extensions.ConfigData, error)
	// UpdateConfigData updates an existing ConfigData.
	UpdateConfigData(ctx api.Context, cfg *extensions.ConfigData) (*extensions.ConfigData, error)
	// DeleteConfigData deletes an existing ConfigData.
	DeleteConfigData(ctx api.Context, name string) error
}

// storage puts strong typing around storage calls
type storage struct {
	rest.StandardStorage
}

// NewRegistry returns a new Registry interface for the given Storage. Any mismatched
// types will panic.
func NewRegistry(s rest.StandardStorage) Registry {
	return &storage{s}
}

func (s *storage) ListConfigDatas(ctx api.Context, options *api.ListOptions) (*extensions.ConfigDataList, error) {
	obj, err := s.List(ctx, options)
	if err != nil {
		return nil, err
	}

	return obj.(*extensions.ConfigDataList), err
}

func (s *storage) WatchConfigDatas(ctx api.Context, options *api.ListOptions) (watch.Interface, error) {
	return s.Watch(ctx, options)
}

func (s *storage) GetConfigData(ctx api.Context, name string) (*extensions.ConfigData, error) {
	obj, err := s.Get(ctx, name)
	if err != nil {
		return nil, err
	}

	return obj.(*extensions.ConfigData), nil
}

func (s *storage) CreateConfigData(ctx api.Context, cfg *extensions.ConfigData) (*extensions.ConfigData, error) {
	obj, err := s.Create(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return obj.(*extensions.ConfigData), nil
}

func (s *storage) UpdateConfigData(ctx api.Context, cfg *extensions.ConfigData) (*extensions.ConfigData, error) {
	obj, _, err := s.Update(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return obj.(*extensions.ConfigData), nil
}

func (s *storage) DeleteConfigData(ctx api.Context, name string) error {
	_, err := s.Delete(ctx, name, nil)

	return err
}
