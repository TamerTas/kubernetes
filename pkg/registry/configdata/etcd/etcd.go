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

package etcd

import (
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/registry/configdata"
	"k8s.io/kubernetes/pkg/registry/generic"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/storage"

	api "k8s.io/kubernetes/pkg/apis/experimental"
	etcdgeneric "k8s.io/kubernetes/pkg/registry/generic/etcd"
)

type REST struct {
	*etcdgeneric.Etcd
}

// NewREST returns a RESTStorage object that will work with ConfigData objects.
func NewREST(s storage.Interface) *REST {
	cfgPrefix := "/configdatas"

	store := &etcdgeneric.Etcd{
		Storage:        s,
		EndpointName:   "configdatas",
		CreateStrategy: configdata.Strategy,
		UpdateStrategy: configdata.Strategy,

		NewFunc: func() runtime.Object {
			return &api.ConfigData{}
		},
		NewListFunc: func() runtime.Object {
			return &api.ConfigDataList{}
		},
		KeyRootFunc: func(ctx api.Context) string {
			return etcdgeneric.NamespaceKeyRootFunc(ctx, cfgPrefix)
		},
		KeyFunc: func(ctx api.Context, id string) (string, error) {
			return etcdgeneric.NamespaceKeyFunc(ctx, cfgPrefix, id)
		},
		ObjectNameFunc: func(obj runtime.Object) (string, error) {
			return obj.(*api.ConfigData).Name, nil
		},
		PredicateFunc: func(label labels.Selector, field fields.Selector) generic.Matcher {
			return configdata.Matcher(label, field)
		},
	}
	return &REST{store}
}
