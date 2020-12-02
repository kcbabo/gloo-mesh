// Code generated by skv2. DO NOT EDIT.

//go:generate mockgen -source ./sets.go -destination mocks/sets.go

package v1alpha1sets

import (
	xds_enterprise_agent_mesh_gloo_solo_io_v1alpha1 "github.com/solo-io/gloo-mesh/pkg/api/xds.enterprise.agent.mesh.gloo.solo.io/v1alpha1"

	"github.com/rotisserie/eris"
	sksets "github.com/solo-io/skv2/contrib/pkg/sets"
	"github.com/solo-io/skv2/pkg/ezkube"
	"k8s.io/apimachinery/pkg/util/sets"
)

type XdsConfigSet interface {
	// Get the set stored keys
	Keys() sets.String
	// List of resources stored in the set. Pass an optional filter function to filter on the list.
	List(filterResource ...func(*xds_enterprise_agent_mesh_gloo_solo_io_v1alpha1.XdsConfig) bool) []*xds_enterprise_agent_mesh_gloo_solo_io_v1alpha1.XdsConfig
	// Return the Set as a map of key to resource.
	Map() map[string]*xds_enterprise_agent_mesh_gloo_solo_io_v1alpha1.XdsConfig
	// Insert a resource into the set.
	Insert(xdsConfig ...*xds_enterprise_agent_mesh_gloo_solo_io_v1alpha1.XdsConfig)
	// Compare the equality of the keys in two sets (not the resources themselves)
	Equal(xdsConfigSet XdsConfigSet) bool
	// Check if the set contains a key matching the resource (not the resource itself)
	Has(xdsConfig ezkube.ResourceId) bool
	// Delete the key matching the resource
	Delete(xdsConfig ezkube.ResourceId)
	// Return the union with the provided set
	Union(set XdsConfigSet) XdsConfigSet
	// Return the difference with the provided set
	Difference(set XdsConfigSet) XdsConfigSet
	// Return the intersection with the provided set
	Intersection(set XdsConfigSet) XdsConfigSet
	// Find the resource with the given ID
	Find(id ezkube.ResourceId) (*xds_enterprise_agent_mesh_gloo_solo_io_v1alpha1.XdsConfig, error)
	// Get the length of the set
	Length() int
}

func makeGenericXdsConfigSet(xdsConfigList []*xds_enterprise_agent_mesh_gloo_solo_io_v1alpha1.XdsConfig) sksets.ResourceSet {
	var genericResources []ezkube.ResourceId
	for _, obj := range xdsConfigList {
		genericResources = append(genericResources, obj)
	}
	return sksets.NewResourceSet(genericResources...)
}

type xdsConfigSet struct {
	set sksets.ResourceSet
}

func NewXdsConfigSet(xdsConfigList ...*xds_enterprise_agent_mesh_gloo_solo_io_v1alpha1.XdsConfig) XdsConfigSet {
	return &xdsConfigSet{set: makeGenericXdsConfigSet(xdsConfigList)}
}

func NewXdsConfigSetFromList(xdsConfigList *xds_enterprise_agent_mesh_gloo_solo_io_v1alpha1.XdsConfigList) XdsConfigSet {
	list := make([]*xds_enterprise_agent_mesh_gloo_solo_io_v1alpha1.XdsConfig, 0, len(xdsConfigList.Items))
	for idx := range xdsConfigList.Items {
		list = append(list, &xdsConfigList.Items[idx])
	}
	return &xdsConfigSet{set: makeGenericXdsConfigSet(list)}
}

func (s *xdsConfigSet) Keys() sets.String {
	if s == nil {
		return sets.String{}
	}
	return s.set.Keys()
}

func (s *xdsConfigSet) List(filterResource ...func(*xds_enterprise_agent_mesh_gloo_solo_io_v1alpha1.XdsConfig) bool) []*xds_enterprise_agent_mesh_gloo_solo_io_v1alpha1.XdsConfig {
	if s == nil {
		return nil
	}
	var genericFilters []func(ezkube.ResourceId) bool
	for _, filter := range filterResource {
		genericFilters = append(genericFilters, func(obj ezkube.ResourceId) bool {
			return filter(obj.(*xds_enterprise_agent_mesh_gloo_solo_io_v1alpha1.XdsConfig))
		})
	}

	var xdsConfigList []*xds_enterprise_agent_mesh_gloo_solo_io_v1alpha1.XdsConfig
	for _, obj := range s.set.List(genericFilters...) {
		xdsConfigList = append(xdsConfigList, obj.(*xds_enterprise_agent_mesh_gloo_solo_io_v1alpha1.XdsConfig))
	}
	return xdsConfigList
}

func (s *xdsConfigSet) Map() map[string]*xds_enterprise_agent_mesh_gloo_solo_io_v1alpha1.XdsConfig {
	if s == nil {
		return nil
	}

	newMap := map[string]*xds_enterprise_agent_mesh_gloo_solo_io_v1alpha1.XdsConfig{}
	for k, v := range s.set.Map() {
		newMap[k] = v.(*xds_enterprise_agent_mesh_gloo_solo_io_v1alpha1.XdsConfig)
	}
	return newMap
}

func (s *xdsConfigSet) Insert(
	xdsConfigList ...*xds_enterprise_agent_mesh_gloo_solo_io_v1alpha1.XdsConfig,
) {
	if s == nil {
		panic("cannot insert into nil set")
	}

	for _, obj := range xdsConfigList {
		s.set.Insert(obj)
	}
}

func (s *xdsConfigSet) Has(xdsConfig ezkube.ResourceId) bool {
	if s == nil {
		return false
	}
	return s.set.Has(xdsConfig)
}

func (s *xdsConfigSet) Equal(
	xdsConfigSet XdsConfigSet,
) bool {
	if s == nil {
		return xdsConfigSet == nil
	}
	return s.set.Equal(makeGenericXdsConfigSet(xdsConfigSet.List()))
}

func (s *xdsConfigSet) Delete(XdsConfig ezkube.ResourceId) {
	if s == nil {
		return
	}
	s.set.Delete(XdsConfig)
}

func (s *xdsConfigSet) Union(set XdsConfigSet) XdsConfigSet {
	if s == nil {
		return set
	}
	return NewXdsConfigSet(append(s.List(), set.List()...)...)
}

func (s *xdsConfigSet) Difference(set XdsConfigSet) XdsConfigSet {
	if s == nil {
		return set
	}
	newSet := s.set.Difference(makeGenericXdsConfigSet(set.List()))
	return &xdsConfigSet{set: newSet}
}

func (s *xdsConfigSet) Intersection(set XdsConfigSet) XdsConfigSet {
	if s == nil {
		return nil
	}
	newSet := s.set.Intersection(makeGenericXdsConfigSet(set.List()))
	var xdsConfigList []*xds_enterprise_agent_mesh_gloo_solo_io_v1alpha1.XdsConfig
	for _, obj := range newSet.List() {
		xdsConfigList = append(xdsConfigList, obj.(*xds_enterprise_agent_mesh_gloo_solo_io_v1alpha1.XdsConfig))
	}
	return NewXdsConfigSet(xdsConfigList...)
}

func (s *xdsConfigSet) Find(id ezkube.ResourceId) (*xds_enterprise_agent_mesh_gloo_solo_io_v1alpha1.XdsConfig, error) {
	if s == nil {
		return nil, eris.Errorf("empty set, cannot find XdsConfig %v", sksets.Key(id))
	}
	obj, err := s.set.Find(&xds_enterprise_agent_mesh_gloo_solo_io_v1alpha1.XdsConfig{}, id)
	if err != nil {
		return nil, err
	}

	return obj.(*xds_enterprise_agent_mesh_gloo_solo_io_v1alpha1.XdsConfig), nil
}

func (s *xdsConfigSet) Length() int {
	if s == nil {
		return 0
	}
	return s.set.Length()
}
