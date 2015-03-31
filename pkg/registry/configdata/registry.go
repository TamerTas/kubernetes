package configdata

import (
	"fmt"

	"k8s.io/kubernetes/pkg/api/rest"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/watch"

	api "k8s.io/kubernetes/pkg/apis/experimental"
)

// Registry is an interface for things that know how to store ConfigDatas.
type Registry interface {
	// ListConfigDatas obtains a list of ConfigDatas having labels and fields which match selector.
	ListConfigDatas(ctx api.Context, label labels.Selector, field fields.Selector) (*api.ConfigDataList, error)
	// WatchConfigDatas watch for new/changed/deleted ConfigDatas.
	WatchConfigDatas(ctx api.Context, label labels.Selector, field fields.Selector, resourceVersion string) (watch.Interface, error)
	// GetConfigDatas gets a specific ConfigData.
	GetConfigData(ctx api.Context, name string) (*api.ConfigData, error)
	// CreateConfigData creates a ConfigData based on a specification.
	CreateConfigData(ctx api.Context, cfg *api.ConfigData) (*api.ConfigData, error)
	// UpdateConfigData updates an existing ConfigData.
	UpdateConfigData(ctx api.Context, cfg *api.ConfigData) (*api.ConfigData, error)
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

func (s *storage) ListConfigDatas(ctx api.Context, label labels.Selector, field fields.Selector) (*api.ConfigDataList, error) {
	if !field.Empty() {
		return nil, fmt.Errorf("field selector not supported yet")
	}

	obj, err := s.List(ctx, label, field)
	if err != nil {
		return nil, err
	}

	return obj.(*api.ConfigDataList), err
}

func (s *storage) WatchConfigDatas(ctx api.Context, label labels.Selector, field fields.Selector, resourceVersion string) (watch.Interface, error) {
	return s.Watch(ctx, label, field, resourceVersion)
}

func (s *storage) GetConfigData(ctx api.Context, name string) (*api.ConfigData, error) {
	obj, err := s.Get(ctx, name)
	if err != nil {
		return nil, err
	}

	return obj.(*api.ConfigData), nil
}

func (s *storage) CreateConfigData(ctx api.Context, cfg *api.ConfigData) (*api.ConfigData, error) {
	obj, err := s.Create(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return obj.(*api.ConfigData), nil
}

func (s *storage) UpdateConfigData(ctx api.Context, cfg *api.ConfigData) (*api.ConfigData, error) {
	obj, _, err := s.Update(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return obj.(*api.ConfigData), nil
}

func (s *storage) DeleteConfigData(ctx api.Context, name string) error {
	_, err := s.Delete(ctx, name, nil)

	return err
}
