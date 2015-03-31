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
	"testing"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/api/rest"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/runtime"
)

func TestCheckGeneratedNameError(t *testing.T) {
	expect := errors.NewNotFound("foo", "bar")
	if err := rest.CheckGeneratedNameError(Strategy, expect, &api.Pod{}); err != expect {
		t.Errorf("NotFoundError should be ignored: %v", err)
	}

	expect = errors.NewAlreadyExists("foo", "bar")
	if err := rest.CheckGeneratedNameError(Strategy, expect, &api.Pod{}); err != expect {
		t.Errorf("AlreadyExists should be returned when no GenerateName field: %v", err)
	}

	expect = errors.NewAlreadyExists("foo", "bar")
	if err := rest.CheckGeneratedNameError(Strategy, expect, &api.Pod{ObjectMeta: api.ObjectMeta{GenerateName: "foo"}}); err == nil || !errors.IsServerTimeout(err) {
		t.Errorf("expected try again later error: %v", err)
	}
}

func TestConfigDataStrategy(t *testing.T) {
	ctx := api.NewDefaultContext()
	if !Strategy.NamespaceScoped() {
		t.Errorf("ConfigData must be namespace scoped")
	}
	if Strategy.AllowCreateOnUpdate() {
		t.Errorf("ConfigData should not allow create on update")
	}

	cfg := &extensions.ConfigData{
		ObjectMeta: api.ObjectMeta{
			Name:      "validConfigData",
			Namespace: api.NamespaceDefault,
		},
	}

	Strategy.PrepareForCreate(cfg)

	errs := Strategy.Validate(ctx, cfg)
	if len(errs) != 0 {
		t.Errorf("unexpected error validating %v", errs)
	}

	newCfg := &extensions.ConfigData{
		ObjectMeta: api.ObjectMeta{
			Name:            "validConfigData2",
			Namespace:       api.NamespaceDefault,
			ResourceVersion: "2",
		},
		Data: map[string]string{
			"key": "updatedValue",
		},
	}

	Strategy.PrepareForUpdate(newCfg, cfg)

	errs = Strategy.ValidateUpdate(ctx, newCfg, cfg)
	if len(errs) == 0 {
		t.Errorf("Expected a validation error")
	}
}

func newConfigData() extensions.ConfigData {
	return extensions.ConfigData{
		ObjectMeta: api.ObjectMeta{
			Name:            "valid",
			Namespace:       "default",
			Labels:          map[string]string{},
			Annotations:     map[string]string{},
			ResourceVersion: "1",
		},
		Data: map[string]string{
			"validKey": "validValue",
		},
	}
}

func TestBeforeUpdate(t *testing.T) {
	testCases := []struct {
		Name   string
		Update func(oldCfg, newCfg *extensions.ConfigData)
		Err    bool
	}{
		{
			Name: "no change",
			Update: func(_, _ *extensions.ConfigData) {
			},
			Err: false,
		},
		{
			Name: "bad namespace",
			Update: func(oldCfg, newCfg *extensions.ConfigData) {
				newCfg.Namespace = "#$%%invalid"
			},
			Err: true,
		},
		{
			Name: "update data",
			Update: func(oldCfg, newCfg *extensions.ConfigData) {
				newCfg.Data["validKey2"] = "validValue2"
			},
			Err: true,
		},
	}

	for _, tc := range testCases {
		oldCfg, newCfg := newConfigData(), newConfigData()

		tc.Update(&oldCfg, &newCfg)

		ctx := api.NewDefaultContext()
		err := rest.BeforeUpdate(Strategy, ctx, runtime.Object(&oldCfg), runtime.Object(&newCfg))
		if tc.Err && err == nil {
			t.Errorf("expected error for %q, got %v", tc.Name, err)
		}
		if !tc.Err && err != nil {
			t.Errorf("unexpected error for %q: got %v", tc.Name, err)
		}
	}
}
