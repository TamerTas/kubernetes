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
	"fmt"

	"k8s.io/kubernetes/pkg/api/rest"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/registry/generic"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/util/fielderrors"

	api "k8s.io/kubernetes/pkg/apis/experimental"
	validation "k8s.io/kubernetes/pkg/apis/experimental/validation"
)

// strategy implements behavior for ConfigData objects
type strategy struct {
	runtime.ObjectTyper
	api.NameGenerator
}

// Strategy is the default logic that applies when creating and updating ConfigData
// objects via the REST API.
var Strategy = strategy{api.Scheme, api.SimpleNameGenerator}

// Strategy should implement rest.RESTCreateStrategy
var _ rest.RESTCreateStrategy = Strategy

// Strategy should implement rest.RESTUpdateStrategy
var _ rest.RESTUpdateStrategy = Strategy

// Strategy should implement rest.RESTDeleteStrategy
var _ rest.RESTDeleteStrategy = Strategy

func (strategy) NamespaceScoped() bool {
	return true
}

func (strategy) PrepareForCreate(obj runtime.Object) {
	cfg := obj.(*api.ConfigData)

	cfg.Data = make(map[string]string)
}

func (strategy) Validate(ctx api.Context, obj runtime.Object) fielderrors.ValidationErrorList {
	cfg := obj.(*api.ConfigData)

	return validation.ValidateConfigData(cfg)
}

func (strategy) AllowUnconditionalUpdate() bool {
	return true
}

func (strategy) AllowCreateOnUpdate() bool {
	return false
}

func (strategy) PrepareForUpdate(newObj, oldObj runtime.Object) {
	//oldCfg, newCfg := obj.(*api.ConfigData), newObj.(*api.ConfigData)
}

func (strategy) ValidateUpdate(ctx api.Context, newObj, oldObj runtime.Object) fielderrors.ValidationErrorList {
	oldCfg, newCfg := obj.(*api.ConfigData), newObj.(*api.ConfigData)

	return validation.ValidateConfigData(oldCfg, newCfg)
}

// SelectableFields returns a field set that represents the object for matching purposes.
func SelectableFields(cfg *api.ConfigData) fields.Set {
	return fields.Set{}
}

// Matcher returns a generic matcher for a given label and field selector.
func MatchConfigData(label labels.Selector, field fields.Selector) generic.Matcher {
	return &generic.SelectionPredicate{
		Label: label,
		Field: field,
		GetAttrs: func(obj runtime.Object) (labels.Set, fields.Set, error) {
			switch obj := obj.(type) {
			case *api.ConfigData:
				return labels.Set(obj.ObjectMeta.Labels), SelectableFields(obj), nil
			default:
				return nil, nil, fmt.Errorf("Given object is not of type ConfigData")
			}
		},
	}
}
