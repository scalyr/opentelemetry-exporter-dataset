// Code generated by mockery v2.12.3. DO NOT EDIT.

package mocks

import (
	time "time"

	gosnmp "github.com/gosnmp/gosnmp"
	mock "github.com/stretchr/testify/mock"
)

// goSNMPWrapper is an autogenerated mock type for the goSNMPWrapper type
type MockGoSNMPWrapper struct {
	mock.Mock
}

// BulkWalkAll provides a mock function with given fields: rootOid
func (_m *MockGoSNMPWrapper) BulkWalkAll(rootOid string) ([]gosnmp.SnmpPDU, error) {
	ret := _m.Called(rootOid)

	var r0 []gosnmp.SnmpPDU
	if rf, ok := ret.Get(0).(func(string) []gosnmp.SnmpPDU); ok {
		r0 = rf(rootOid)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]gosnmp.SnmpPDU)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(rootOid)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Close provides a mock function with given fields:
func (_m *MockGoSNMPWrapper) Close() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Connect provides a mock function with given fields:
func (_m *MockGoSNMPWrapper) Connect() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Get provides a mock function with given fields: oids
func (_m *MockGoSNMPWrapper) Get(oids []string) (*gosnmp.SnmpPacket, error) {
	ret := _m.Called(oids)

	var r0 *gosnmp.SnmpPacket
	if rf, ok := ret.Get(0).(func([]string) *gosnmp.SnmpPacket); ok {
		r0 = rf(oids)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gosnmp.SnmpPacket)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]string) error); ok {
		r1 = rf(oids)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetCommunity provides a mock function with given fields:
func (_m *MockGoSNMPWrapper) GetCommunity() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetMaxOids provides a mock function with given fields:
func (_m *MockGoSNMPWrapper) GetMaxOids() int {
	ret := _m.Called()

	var r0 int
	if rf, ok := ret.Get(0).(func() int); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}

// GetMsgFlags provides a mock function with given fields:
func (_m *MockGoSNMPWrapper) GetMsgFlags() gosnmp.SnmpV3MsgFlags {
	ret := _m.Called()

	var r0 gosnmp.SnmpV3MsgFlags
	if rf, ok := ret.Get(0).(func() gosnmp.SnmpV3MsgFlags); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(gosnmp.SnmpV3MsgFlags)
	}

	return r0
}

// GetPort provides a mock function with given fields:
func (_m *MockGoSNMPWrapper) GetPort() uint16 {
	ret := _m.Called()

	var r0 uint16
	if rf, ok := ret.Get(0).(func() uint16); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(uint16)
	}

	return r0
}

// GetSecurityModel provides a mock function with given fields:
func (_m *MockGoSNMPWrapper) GetSecurityModel() gosnmp.SnmpV3SecurityModel {
	ret := _m.Called()

	var r0 gosnmp.SnmpV3SecurityModel
	if rf, ok := ret.Get(0).(func() gosnmp.SnmpV3SecurityModel); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(gosnmp.SnmpV3SecurityModel)
	}

	return r0
}

// GetSecurityParameters provides a mock function with given fields:
func (_m *MockGoSNMPWrapper) GetSecurityParameters() gosnmp.SnmpV3SecurityParameters {
	ret := _m.Called()

	var r0 gosnmp.SnmpV3SecurityParameters
	if rf, ok := ret.Get(0).(func() gosnmp.SnmpV3SecurityParameters); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(gosnmp.SnmpV3SecurityParameters)
		}
	}

	return r0
}

// GetTarget provides a mock function with given fields:
func (_m *MockGoSNMPWrapper) GetTarget() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetTimeout provides a mock function with given fields:
func (_m *MockGoSNMPWrapper) GetTimeout() time.Duration {
	ret := _m.Called()

	var r0 time.Duration
	if rf, ok := ret.Get(0).(func() time.Duration); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(time.Duration)
	}

	return r0
}

// GetTransport provides a mock function with given fields:
func (_m *MockGoSNMPWrapper) GetTransport() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetVersion provides a mock function with given fields:
func (_m *MockGoSNMPWrapper) GetVersion() gosnmp.SnmpVersion {
	ret := _m.Called()

	var r0 gosnmp.SnmpVersion
	if rf, ok := ret.Get(0).(func() gosnmp.SnmpVersion); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(gosnmp.SnmpVersion)
	}

	return r0
}

// SetCommunity provides a mock function with given fields: community
func (_m *MockGoSNMPWrapper) SetCommunity(community string) {
	_m.Called(community)
}

// SetMaxOids provides a mock function with given fields: maxOids
func (_m *MockGoSNMPWrapper) SetMaxOids(maxOids int) {
	_m.Called(maxOids)
}

// SetMsgFlags provides a mock function with given fields: msgFlags
func (_m *MockGoSNMPWrapper) SetMsgFlags(msgFlags gosnmp.SnmpV3MsgFlags) {
	_m.Called(msgFlags)
}

// SetPort provides a mock function with given fields: port
func (_m *MockGoSNMPWrapper) SetPort(port uint16) {
	_m.Called(port)
}

// SetSecurityModel provides a mock function with given fields: securityModel
func (_m *MockGoSNMPWrapper) SetSecurityModel(securityModel gosnmp.SnmpV3SecurityModel) {
	_m.Called(securityModel)
}

// SetSecurityParameters provides a mock function with given fields: securityParameters
func (_m *MockGoSNMPWrapper) SetSecurityParameters(securityParameters gosnmp.SnmpV3SecurityParameters) {
	_m.Called(securityParameters)
}

// SetTarget provides a mock function with given fields: target
func (_m *MockGoSNMPWrapper) SetTarget(target string) {
	_m.Called(target)
}

// SetTimeout provides a mock function with given fields: timeout
func (_m *MockGoSNMPWrapper) SetTimeout(timeout time.Duration) {
	_m.Called(timeout)
}

// SetTransport provides a mock function with given fields: transport
func (_m *MockGoSNMPWrapper) SetTransport(transport string) {
	_m.Called(transport)
}

// SetVersion provides a mock function with given fields: version
func (_m *MockGoSNMPWrapper) SetVersion(version gosnmp.SnmpVersion) {
	_m.Called(version)
}

// WalkAll provides a mock function with given fields: rootOid
func (_m *MockGoSNMPWrapper) WalkAll(rootOid string) ([]gosnmp.SnmpPDU, error) {
	ret := _m.Called(rootOid)

	var r0 []gosnmp.SnmpPDU
	if rf, ok := ret.Get(0).(func(string) []gosnmp.SnmpPDU); ok {
		r0 = rf(rootOid)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]gosnmp.SnmpPDU)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(rootOid)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockNewGoSNMPWrapperT interface {
	mock.TestingT
	Cleanup(func())
}

// newMockGoSNMPWrapper creates a new instance of goSNMPWrapper. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func newMockGoSNMPWrapper(t mockNewGoSNMPWrapperT) *MockGoSNMPWrapper {
	mock := &MockGoSNMPWrapper{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
