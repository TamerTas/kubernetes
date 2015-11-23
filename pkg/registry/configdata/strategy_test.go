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
			Name:      "valid-config-data",
			Namespace: api.NamespaceDefault,
		},
		Data: map[string]string{
			"foo": "bar",
		},
	}

	Strategy.PrepareForCreate(cfg)

	errs := Strategy.Validate(ctx, cfg)
	if len(errs) != 0 {
		t.Errorf("unexpected error validating %v", errs)
	}

	newCfg := &extensions.ConfigData{
		ObjectMeta: api.ObjectMeta{
			Name:            "valid-config-data-2",
			Namespace:       api.NamespaceDefault,
			ResourceVersion: "4",
		},
		Data: map[string]string{
			"invalidKey": "updatedValue",
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
			"valid-key": "validValue",
		},
	}
}

func TestBeforeUpdate(t *testing.T) {
	cases := []struct {
		name   string
		update func(_, _ *extensions.ConfigData)
		err    bool
	}{
		{
			name:   "no change",
			update: func(_, _ *extensions.ConfigData) {},
			err:    false,
		},
		{
			name: "bad namespace",
			update: func(_, newCfg *extensions.ConfigData) {
				newCfg.Namespace = "#$%%invalid"
			},
			err: true,
		},
		{
			name: "bad update to data",
			update: func(_, newCfg *extensions.ConfigData) {
				newCfg.Data["%%#@$invalidKey"] = "validValue2"
			},
			err: true,
		},
	}

	for _, tc := range cases {
		var (
			oldCfg = newConfigData()
			newCfg = newConfigData()
		)

		tc.update(&oldCfg, &newCfg)

		ctx := api.NewDefaultContext()
		err := rest.BeforeUpdate(Strategy, ctx, runtime.Object(&newCfg), runtime.Object(&oldCfg))
		if tc.err && err == nil {
			t.Errorf("expected error for %q, got %v", tc.name, err)
		}
		if !tc.err && err != nil {
			t.Errorf("unexpected error for %q: got %v", tc.name, err)
		}
	}
}
