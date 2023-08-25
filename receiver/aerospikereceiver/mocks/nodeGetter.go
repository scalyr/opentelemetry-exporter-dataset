// Code generated by mockery v2.13.1. DO NOT EDIT.

package mocks

import (
	mock "github.com/stretchr/testify/mock"

	cluster "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/aerospikereceiver/cluster"
)

// NodeGetter is an autogenerated mock type for the nodeGetter type
type NodeGetter struct {
	mock.Mock
}

// Close provides a mock function with given fields:
func (_m *NodeGetter) Close() {
	_m.Called()
}

// GetNodes provides a mock function with given fields:
func (_m *NodeGetter) GetNodes() []cluster.Node {
	ret := _m.Called()

	var r0 []cluster.Node
	if rf, ok := ret.Get(0).(func() []cluster.Node); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]cluster.Node)
		}
	}

	return r0
}

type mockConstructorTestingTNewNodeGetter interface {
	mock.TestingT
	Cleanup(func())
}

// NewNodeGetter creates a new instance of NodeGetter. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewNodeGetter(t mockConstructorTestingTNewNodeGetter) *NodeGetter {
	mock := &NodeGetter{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
