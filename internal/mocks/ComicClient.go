// Code generated by mockery v2.43.1. DO NOT EDIT.

package mocks

import (
	context "context"
	domain "yadro-microservices/internal/core/domain"

	mock "github.com/stretchr/testify/mock"
)

// ComicClient is an autogenerated mock type for the ComicClient type
type ComicClient struct {
	mock.Mock
}

// GetComics provides a mock function with given fields: ctx, existingIDs
func (_m *ComicClient) GetComics(ctx context.Context, existingIDs map[int]bool) (domain.Comics, error) {
	ret := _m.Called(ctx, existingIDs)

	if len(ret) == 0 {
		panic("no return value specified for GetComics")
	}

	var r0 domain.Comics
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, map[int]bool) (domain.Comics, error)); ok {
		return rf(ctx, existingIDs)
	}
	if rf, ok := ret.Get(0).(func(context.Context, map[int]bool) domain.Comics); ok {
		r0 = rf(ctx, existingIDs)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(domain.Comics)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, map[int]bool) error); ok {
		r1 = rf(ctx, existingIDs)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewComicClient creates a new instance of ComicClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewComicClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *ComicClient {
	mock := &ComicClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}