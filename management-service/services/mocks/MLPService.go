// Code generated by mockery (devel). DO NOT EDIT.

package mocks

import (
	client "github.com/caraml-dev/mlp/api/client"
	mock "github.com/stretchr/testify/mock"
)

// MLPService is an autogenerated mock type for the MLPService type
type MLPService struct {
	mock.Mock
}

// GetProject provides a mock function with given fields: id
func (_m *MLPService) GetProject(id int64) (*client.Project, error) {
	ret := _m.Called(id)

	var r0 *client.Project
	if rf, ok := ret.Get(0).(func(int64) *client.Project); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*client.Project)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(int64) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
