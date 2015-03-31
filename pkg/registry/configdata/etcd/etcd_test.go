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
	"testing"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/registry/registrytest"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/tools"
)

func newStorage(t *testing.T) (*REST, *tools.FakeEtcdClient) {
	etcdStorage, fakeClient := registrytest.NewEtcdStorage(t, "extensions")

	return NewREST(etcdStorage), fakeClient
}

func validNewConfigData() *extensions.ConfigData {
	return &extensions.ConfigData{
		ObjectMeta: api.ObjectMeta{
			Name:      "foo",
			Namespace: "default",
		},
		Data: map[string]string{
			"test": "data",
		},
	}
}

func TestCreate(t *testing.T) {
	storage, fakeClient := newStorage(t)
	test := registrytest.New(t, fakeClient, storage.Etcd)

	validConfigData := validNewConfigData()
	validConfigData.ObjectMeta = api.ObjectMeta{
		GenerateName: "foo-",
	}

	test.TestCreate(
		validConfigData,
		&extensions.ConfigData,
		&extensions.ConfigData{
			ObjectMeta: api.ObjectMeta{Name: "name"},
			Data: map[string]string{
				"key": "value",
			},
		},
		&extensions.ConfigData{
			ObjectMeta: api.ObjectMeta{Name: "name"},
			Data: map[string]string{
				"dotfile": "do: nothing\n",
			},
		},
	)
}

func TestUpdate(t *testing.T) {
	storage, fakeClient := newStorage(t)
	test := registrytest.New(t, fakeClient, storage.Etcd)
	test.TestUpdate(
		// valid
		validNewConfigData(),
		// updateFunc
		func(obj runtime.Object) runtime.Object {
			cfg := obj.(*extensions.ConfigData)
			cfg.Data["updateTest"] = "value"
			return cfg
		},
		// invalid updateFunc
		func(obj runtime.Object) runtime.Object {
			cfg := obj.(*extensions.ConfigData)
			return cfg
		},
	)
}

func TestDelete(t *testing.T) {
	storage, fakeClient := newStorage(t)
	test := registrytest.New(t, fakeClient, storage.Etcd)
	test.TestDelete(validNewConfigData())
}

func TestGet(t *testing.T) {
	storage, fakeClient := newStorage(t)
	test := registrytest.New(t, fakeClient, storage.Etcd)
	test.TestGet(validNewConfigData())
}

func TestList(t *testing.T) {
	storage, fakeClient := newStorage(t)
	test := registrytest.New(t, fakeClient, storage.Etcd)
	test.TestList(validNewConfigData())
}

func TestWatch(t *testing.T) {
	storage, fakeClient := newStorage(t)
	test := registrytest.New(t, fakeClient, storage.Etcd)
	test.TestWatch(
		validNewConfigData(),
		// matching labels
		[]labels.Set{},
		// not matching labels
		[]labels.Set{
			{"foo": "bar"},
		},
		// matching fields
		[]fields.Set{},
		// not matching fields
		[]fields.Set{
			{"metadata.name": "bar"},
			{"name": "foo"},
		},
	)
}
