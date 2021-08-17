// Code generated by protoc-gen-ext. DO NOT EDIT.
// source: github.com/solo-io/gloo-mesh/api/common/v1/status.proto

package v1

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"

	"github.com/golang/protobuf/proto"
	equality "github.com/solo-io/protoc-gen-ext/pkg/equality"
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
)

// Equal function
func (m *AppliedIngressGateway) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*AppliedIngressGateway)
	if !ok {
		that2, ok := that.(AppliedIngressGateway)
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

	if h, ok := interface{}(m.GetDestinationRef()).(equality.Equalizer); ok {
		if !h.Equal(target.GetDestinationRef()) {
			return false
		}
	} else {
		if !proto.Equal(m.GetDestinationRef(), target.GetDestinationRef()) {
			return false
		}
	}

	if len(m.GetExternalAddresses()) != len(target.GetExternalAddresses()) {
		return false
	}
	for idx, v := range m.GetExternalAddresses() {

		if strings.Compare(v, target.GetExternalAddresses()[idx]) != 0 {
			return false
		}

	}

	if m.GetPort() != target.GetPort() {
		return false
	}

	if m.GetExternalPort() != target.GetExternalPort() {
		return false
	}

	return true
}