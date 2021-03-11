// Code generated by protoc-gen-ext. DO NOT EDIT.
// source: github.com/solo-io/gloo-mesh/api/enterprise/networking/v1beta1/virtual_destination.proto

package v1beta1

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"

	"github.com/golang/protobuf/proto"
	equality "github.com/solo-io/protoc-gen-ext/pkg/equality"

	v1 "github.com/solo-io/gloo-mesh/pkg/api/common.mesh.gloo.solo.io/v1"
)

// ensure the imports are used
var (
	_ = errors.New("")
	_ = fmt.Print
	_ = binary.LittleEndian
	_ = bytes.Compare
	_ = strings.Compare
	_ = equality.Equalizer(nil)
	_ = proto.Message(nil)

	_ = v1.ApprovalState(0)
)

// Equal function
func (m *VirtualDestinationSpec) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*VirtualDestinationSpec)
	if !ok {
		that2, ok := that.(VirtualDestinationSpec)
		if ok {
			target = &that2
		} else {
			return false
		}
	}
	if target == nil {
		return m == nil
	} else if m == nil {
		return false
	}

	if strings.Compare(m.GetHostname(), target.GetHostname()) != 0 {
		return false
	}

	if h, ok := interface{}(m.GetPort()).(equality.Equalizer); ok {
		if !h.Equal(target.GetPort()) {
			return false
		}
	} else {
		if !proto.Equal(m.GetPort(), target.GetPort()) {
			return false
		}
	}

	switch m.ExportTo.(type) {

	case *VirtualDestinationSpec_VirtualMesh:
		if _, ok := target.ExportTo.(*VirtualDestinationSpec_VirtualMesh); !ok {
			return false
		}

		if h, ok := interface{}(m.GetVirtualMesh()).(equality.Equalizer); ok {
			if !h.Equal(target.GetVirtualMesh()) {
				return false
			}
		} else {
			if !proto.Equal(m.GetVirtualMesh(), target.GetVirtualMesh()) {
				return false
			}
		}

	case *VirtualDestinationSpec_MeshList_:
		if _, ok := target.ExportTo.(*VirtualDestinationSpec_MeshList_); !ok {
			return false
		}

		if h, ok := interface{}(m.GetMeshList()).(equality.Equalizer); ok {
			if !h.Equal(target.GetMeshList()) {
				return false
			}
		} else {
			if !proto.Equal(m.GetMeshList(), target.GetMeshList()) {
				return false
			}
		}

	default:
		// m is nil but target is not nil
		if m.ExportTo != target.ExportTo {
			return false
		}
	}

	switch m.FailoverConfig.(type) {

	case *VirtualDestinationSpec_Static:
		if _, ok := target.FailoverConfig.(*VirtualDestinationSpec_Static); !ok {
			return false
		}

		if h, ok := interface{}(m.GetStatic()).(equality.Equalizer); ok {
			if !h.Equal(target.GetStatic()) {
				return false
			}
		} else {
			if !proto.Equal(m.GetStatic(), target.GetStatic()) {
				return false
			}
		}

	case *VirtualDestinationSpec_Localized:
		if _, ok := target.FailoverConfig.(*VirtualDestinationSpec_Localized); !ok {
			return false
		}

		if h, ok := interface{}(m.GetLocalized()).(equality.Equalizer); ok {
			if !h.Equal(target.GetLocalized()) {
				return false
			}
		} else {
			if !proto.Equal(m.GetLocalized(), target.GetLocalized()) {
				return false
			}
		}

	default:
		// m is nil but target is not nil
		if m.FailoverConfig != target.FailoverConfig {
			return false
		}
	}

	return true
}

// Equal function
func (m *VirtualDestinationBackingDestination) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*VirtualDestinationBackingDestination)
	if !ok {
		that2, ok := that.(VirtualDestinationBackingDestination)
		if ok {
			target = &that2
		} else {
			return false
		}
	}
	if target == nil {
		return m == nil
	} else if m == nil {
		return false
	}

	switch m.Type.(type) {

	case *VirtualDestinationBackingDestination_KubeService:
		if _, ok := target.Type.(*VirtualDestinationBackingDestination_KubeService); !ok {
			return false
		}

		if h, ok := interface{}(m.GetKubeService()).(equality.Equalizer); ok {
			if !h.Equal(target.GetKubeService()) {
				return false
			}
		} else {
			if !proto.Equal(m.GetKubeService(), target.GetKubeService()) {
				return false
			}
		}

	default:
		// m is nil but target is not nil
		if m.Type != target.Type {
			return false
		}
	}

	return true
}

// Equal function
func (m *VirtualDestinationStatus) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*VirtualDestinationStatus)
	if !ok {
		that2, ok := that.(VirtualDestinationStatus)
		if ok {
			target = &that2
		} else {
			return false
		}
	}
	if target == nil {
		return m == nil
	} else if m == nil {
		return false
	}

	if m.GetObservedGeneration() != target.GetObservedGeneration() {
		return false
	}

	if m.GetState() != target.GetState() {
		return false
	}

	if len(m.GetMeshes()) != len(target.GetMeshes()) {
		return false
	}
	for k, v := range m.GetMeshes() {

		if h, ok := interface{}(v).(equality.Equalizer); ok {
			if !h.Equal(target.GetMeshes()[k]) {
				return false
			}
		} else {
			if !proto.Equal(v, target.GetMeshes()[k]) {
				return false
			}
		}

	}

	if len(m.GetSelectedDestinations()) != len(target.GetSelectedDestinations()) {
		return false
	}
	for idx, v := range m.GetSelectedDestinations() {

		if h, ok := interface{}(v).(equality.Equalizer); ok {
			if !h.Equal(target.GetSelectedDestinations()[idx]) {
				return false
			}
		} else {
			if !proto.Equal(v, target.GetSelectedDestinations()[idx]) {
				return false
			}
		}

	}

	if len(m.GetErrors()) != len(target.GetErrors()) {
		return false
	}
	for idx, v := range m.GetErrors() {

		if strings.Compare(v, target.GetErrors()[idx]) != 0 {
			return false
		}

	}

	return true
}

// Equal function
func (m *VirtualDestinationSpec_Port) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*VirtualDestinationSpec_Port)
	if !ok {
		that2, ok := that.(VirtualDestinationSpec_Port)
		if ok {
			target = &that2
		} else {
			return false
		}
	}
	if target == nil {
		return m == nil
	} else if m == nil {
		return false
	}

	if m.GetNumber() != target.GetNumber() {
		return false
	}

	if strings.Compare(m.GetProtocol(), target.GetProtocol()) != 0 {
		return false
	}

	return true
}

// Equal function
func (m *VirtualDestinationSpec_MeshList) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*VirtualDestinationSpec_MeshList)
	if !ok {
		that2, ok := that.(VirtualDestinationSpec_MeshList)
		if ok {
			target = &that2
		} else {
			return false
		}
	}
	if target == nil {
		return m == nil
	} else if m == nil {
		return false
	}

	if len(m.GetMeshes()) != len(target.GetMeshes()) {
		return false
	}
	for idx, v := range m.GetMeshes() {

		if h, ok := interface{}(v).(equality.Equalizer); ok {
			if !h.Equal(target.GetMeshes()[idx]) {
				return false
			}
		} else {
			if !proto.Equal(v, target.GetMeshes()[idx]) {
				return false
			}
		}

	}

	return true
}

// Equal function
func (m *VirtualDestinationSpec_BackingDestinationList) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*VirtualDestinationSpec_BackingDestinationList)
	if !ok {
		that2, ok := that.(VirtualDestinationSpec_BackingDestinationList)
		if ok {
			target = &that2
		} else {
			return false
		}
	}
	if target == nil {
		return m == nil
	} else if m == nil {
		return false
	}

	if len(m.GetDestinations()) != len(target.GetDestinations()) {
		return false
	}
	for idx, v := range m.GetDestinations() {

		if h, ok := interface{}(v).(equality.Equalizer); ok {
			if !h.Equal(target.GetDestinations()[idx]) {
				return false
			}
		} else {
			if !proto.Equal(v, target.GetDestinations()[idx]) {
				return false
			}
		}

	}

	return true
}

// Equal function
func (m *VirtualDestinationSpec_LocalityConfig) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*VirtualDestinationSpec_LocalityConfig)
	if !ok {
		that2, ok := that.(VirtualDestinationSpec_LocalityConfig)
		if ok {
			target = &that2
		} else {
			return false
		}
	}
	if target == nil {
		return m == nil
	} else if m == nil {
		return false
	}

	if len(m.GetDestinationSelectors()) != len(target.GetDestinationSelectors()) {
		return false
	}
	for idx, v := range m.GetDestinationSelectors() {

		if h, ok := interface{}(v).(equality.Equalizer); ok {
			if !h.Equal(target.GetDestinationSelectors()[idx]) {
				return false
			}
		} else {
			if !proto.Equal(v, target.GetDestinationSelectors()[idx]) {
				return false
			}
		}

	}

	if len(m.GetFailoverDirectives()) != len(target.GetFailoverDirectives()) {
		return false
	}
	for idx, v := range m.GetFailoverDirectives() {

		if h, ok := interface{}(v).(equality.Equalizer); ok {
			if !h.Equal(target.GetFailoverDirectives()[idx]) {
				return false
			}
		} else {
			if !proto.Equal(v, target.GetFailoverDirectives()[idx]) {
				return false
			}
		}

	}

	if h, ok := interface{}(m.GetOutlierDetection()).(equality.Equalizer); ok {
		if !h.Equal(target.GetOutlierDetection()) {
			return false
		}
	} else {
		if !proto.Equal(m.GetOutlierDetection(), target.GetOutlierDetection()) {
			return false
		}
	}

	return true
}

// Equal function
func (m *VirtualDestinationSpec_LocalityConfig_LocalityFailoverDirective) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*VirtualDestinationSpec_LocalityConfig_LocalityFailoverDirective)
	if !ok {
		that2, ok := that.(VirtualDestinationSpec_LocalityConfig_LocalityFailoverDirective)
		if ok {
			target = &that2
		} else {
			return false
		}
	}
	if target == nil {
		return m == nil
	} else if m == nil {
		return false
	}

	if h, ok := interface{}(m.GetFrom()).(equality.Equalizer); ok {
		if !h.Equal(target.GetFrom()) {
			return false
		}
	} else {
		if !proto.Equal(m.GetFrom(), target.GetFrom()) {
			return false
		}
	}

	if len(m.GetTo()) != len(target.GetTo()) {
		return false
	}
	for idx, v := range m.GetTo() {

		if h, ok := interface{}(v).(equality.Equalizer); ok {
			if !h.Equal(target.GetTo()[idx]) {
				return false
			}
		} else {
			if !proto.Equal(v, target.GetTo()[idx]) {
				return false
			}
		}

	}

	return true
}

// Equal function
func (m *VirtualDestinationSpec_LocalityConfig_Locality) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*VirtualDestinationSpec_LocalityConfig_Locality)
	if !ok {
		that2, ok := that.(VirtualDestinationSpec_LocalityConfig_Locality)
		if ok {
			target = &that2
		} else {
			return false
		}
	}
	if target == nil {
		return m == nil
	} else if m == nil {
		return false
	}

	if strings.Compare(m.GetRegion(), target.GetRegion()) != 0 {
		return false
	}

	if strings.Compare(m.GetZone(), target.GetZone()) != 0 {
		return false
	}

	if strings.Compare(m.GetSubZone(), target.GetSubZone()) != 0 {
		return false
	}

	return true
}

// Equal function
func (m *VirtualDestinationStatus_SelectedDestinations) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*VirtualDestinationStatus_SelectedDestinations)
	if !ok {
		that2, ok := that.(VirtualDestinationStatus_SelectedDestinations)
		if ok {
			target = &that2
		} else {
			return false
		}
	}
	if target == nil {
		return m == nil
	} else if m == nil {
		return false
	}

	if h, ok := interface{}(m.GetRef()).(equality.Equalizer); ok {
		if !h.Equal(target.GetRef()) {
			return false
		}
	} else {
		if !proto.Equal(m.GetRef(), target.GetRef()) {
			return false
		}
	}

	if h, ok := interface{}(m.GetDestination()).(equality.Equalizer); ok {
		if !h.Equal(target.GetDestination()) {
			return false
		}
	} else {
		if !proto.Equal(m.GetDestination(), target.GetDestination()) {
			return false
		}
	}

	return true
}