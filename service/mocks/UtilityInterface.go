// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// UtilityInterface is an autogenerated mock type for the UtilityInterface type
type UtilityInterface struct {
	mock.Mock
}

// EncodeInput provides a mock function with given fields: input
func (_m *UtilityInterface) EncodeInput(input string) string {
	ret := _m.Called(input)

	var r0 string
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(input)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

type mockConstructorTestingTNewUtilityInterface interface {
	mock.TestingT
	Cleanup(func())
}

// NewUtilityInterface creates a new instance of UtilityInterface. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewUtilityInterface(t mockConstructorTestingTNewUtilityInterface) *UtilityInterface {
	mock := &UtilityInterface{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
