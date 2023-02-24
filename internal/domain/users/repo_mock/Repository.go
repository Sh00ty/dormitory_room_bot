// Code generated by mockery v2.12.1. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
	entities "gitlab.com/Sh00ty/dormitory_room_bot/internal/entities"

	testing "testing"

	value_objects "gitlab.com/Sh00ty/dormitory_room_bot/internal/value_objects"
)

// Repository is an autogenerated mock type for the Repository type
type Repository struct {
	mock.Mock
}

// Create provides a mock function with given fields: ctx, usr
func (_m *Repository) Create(ctx context.Context, usr entities.User) error {
	ret := _m.Called(ctx, usr)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, entities.User) error); ok {
		r0 = rf(ctx, usr)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Delete provides a mock function with given fields: ctx, userID
func (_m *Repository) Delete(ctx context.Context, userID value_objects.UserID) error {
	ret := _m.Called(ctx, userID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, value_objects.UserID) error); ok {
		r0 = rf(ctx, userID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetBatch provides a mock function with given fields: ctx, userIDList
func (_m *Repository) GetBatch(ctx context.Context, userIDList []value_objects.UserID) ([]entities.User, error) {
	ret := _m.Called(ctx, userIDList)

	var r0 []entities.User
	if rf, ok := ret.Get(0).(func(context.Context, []value_objects.UserID) []entities.User); ok {
		r0 = rf(ctx, userIDList)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]entities.User)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, []value_objects.UserID) error); ok {
		r1 = rf(ctx, userIDList)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetBatchFromUsernames provides a mock function with given fields: ctx, usernameList
func (_m *Repository) GetBatchFromUsernames(ctx context.Context, usernameList []string) ([]entities.User, error) {
	ret := _m.Called(ctx, usernameList)

	var r0 []entities.User
	if rf, ok := ret.Get(0).(func(context.Context, []string) []entities.User); ok {
		r0 = rf(ctx, usernameList)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]entities.User)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, []string) error); ok {
		r1 = rf(ctx, usernameList)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewRepository creates a new instance of Repository. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewRepository(t testing.TB) *Repository {
	mock := &Repository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}