// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"
import rest "k8s.io/client-go/rest"
import v1alpha3 "github.com/gojek/merlin/istio/client-go/pkg/clientset/versioned/typed/networking/v1alpha3"

// NetworkingV1alpha3Interface is an autogenerated mock type for the NetworkingV1alpha3Interface type
type NetworkingV1alpha3Interface struct {
	mock.Mock
}

// RESTClient provides a mock function with given fields:
func (_m *NetworkingV1alpha3Interface) RESTClient() rest.Interface {
	ret := _m.Called()

	var r0 rest.Interface
	if rf, ok := ret.Get(0).(func() rest.Interface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(rest.Interface)
		}
	}

	return r0
}

// VirtualServices provides a mock function with given fields: namespace
func (_m *NetworkingV1alpha3Interface) VirtualServices(namespace string) v1alpha3.VirtualServiceInterface {
	ret := _m.Called(namespace)

	var r0 v1alpha3.VirtualServiceInterface
	if rf, ok := ret.Get(0).(func(string) v1alpha3.VirtualServiceInterface); ok {
		r0 = rf(namespace)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(v1alpha3.VirtualServiceInterface)
		}
	}

	return r0
}