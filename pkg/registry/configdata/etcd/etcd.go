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
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/registry/configdata"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/storage"

	etcdgeneric "k8s.io/kubernetes/pkg/registry/generic/etcd"
)

// REST implements a RESTStorage for ConfigData against etcd
type REST struct {
	*etcdgeneric.Etcd
}

// NewREST returns a RESTStorage object that will work with ConfigData objects.
func NewREST(storageInterface storage.Interface) *REST {
	cfgPrefix := "/configDatas"

	store := &etcdgeneric.Etcd{
		NewFunc: func() runtime.Object {
			return &extensions.ConfigData{}
		},

		// NewListFunc returns an object to store results of an etcd list.
		NewListFunc: func() runtime.Object {
			return &extensions.ConfigDataList{}
		},

		// Produces a path that etcd understands, to the root of the resource
		// by combining the namespace in the context with the given prefix.
		KeyRootFunc: func(ctx extensions.Context) string {
			return etcdgeneric.NamespaceKeyRootFunc(ctx, cfgPrefix)
		},

		// Produces a path that etcd understands, to the resource by combining
		// the namespace in the context with the given prefix
		KeyFunc: func(ctx extensions.Context, name string) (string, error) {
			return etcdgeneric.NamespaceKeyFunc(ctx, cfgPrefix, name)
		},

		// Retrieves the name field of a ConfigData object.
		ObjectNameFunc: func(obj runtime.Object) (string, error) {
			return obj.(*extensions.ConfigData).Name, nil
		},

		// Matches objects based on labels/fields for list and watch
		PredicateFunc: configdata.Match,

		EndpointName: "configDatas",

		CreateStrategy: configdata.Strategy,
		UpdateStrategy: configdata.Strategy,

		Storage: storageInterface,
	}
	return &REST{store}
}
