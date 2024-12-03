// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/adettelle/go-keeper/internal/server/api (interfaces: ICardRepo)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	repo "github.com/adettelle/go-keeper/internal/repo"
	gomock "github.com/golang/mock/gomock"
)

// MockICardRepo is a mock of ICardRepo interface.
type MockICardRepo struct {
	ctrl     *gomock.Controller
	recorder *MockICardRepoMockRecorder
}

// MockICardRepoMockRecorder is the mock recorder for MockICardRepo.
type MockICardRepoMockRecorder struct {
	mock *MockICardRepo
}

// NewMockICardRepo creates a new mock instance.
func NewMockICardRepo(ctrl *gomock.Controller) *MockICardRepo {
	mock := &MockICardRepo{ctrl: ctrl}
	mock.recorder = &MockICardRepoMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockICardRepo) EXPECT() *MockICardRepoMockRecorder {
	return m.recorder
}

// AddCard mocks base method.
func (m *MockICardRepo) AddCard(arg0 context.Context, arg1, arg2, arg3, arg4, arg5, arg6 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddCard", arg0, arg1, arg2, arg3, arg4, arg5, arg6)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddCard indicates an expected call of AddCard.
func (mr *MockICardRepoMockRecorder) AddCard(arg0, arg1, arg2, arg3, arg4, arg5, arg6 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddCard", reflect.TypeOf((*MockICardRepo)(nil).AddCard), arg0, arg1, arg2, arg3, arg4, arg5, arg6)
}

// DeleteCardByTitle mocks base method.
func (m *MockICardRepo) DeleteCardByTitle(arg0 context.Context, arg1, arg2 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteCardByTitle", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteCardByTitle indicates an expected call of DeleteCardByTitle.
func (mr *MockICardRepoMockRecorder) DeleteCardByTitle(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteCardByTitle", reflect.TypeOf((*MockICardRepo)(nil).DeleteCardByTitle), arg0, arg1, arg2)
}

// GetAllCards mocks base method.
func (m *MockICardRepo) GetAllCards(arg0 context.Context, arg1 string) ([]repo.CardToGet, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAllCards", arg0, arg1)
	ret0, _ := ret[0].([]repo.CardToGet)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAllCards indicates an expected call of GetAllCards.
func (mr *MockICardRepoMockRecorder) GetAllCards(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllCards", reflect.TypeOf((*MockICardRepo)(nil).GetAllCards), arg0, arg1)
}

// GetCardByTitle mocks base method.
func (m *MockICardRepo) GetCardByTitle(arg0 context.Context, arg1, arg2 string) (repo.CardGetByTitle, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCardByTitle", arg0, arg1, arg2)
	ret0, _ := ret[0].(repo.CardGetByTitle)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCardByTitle indicates an expected call of GetCardByTitle.
func (mr *MockICardRepoMockRecorder) GetCardByTitle(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCardByTitle", reflect.TypeOf((*MockICardRepo)(nil).GetCardByTitle), arg0, arg1, arg2)
}

// UpdateCard mocks base method.
func (m *MockICardRepo) UpdateCard(arg0 context.Context, arg1 string, arg2, arg3, arg4, arg5 *string, arg6 int) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateCard", arg0, arg1, arg2, arg3, arg4, arg5, arg6)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateCard indicates an expected call of UpdateCard.
func (mr *MockICardRepoMockRecorder) UpdateCard(arg0, arg1, arg2, arg3, arg4, arg5, arg6 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateCard", reflect.TypeOf((*MockICardRepo)(nil).UpdateCard), arg0, arg1, arg2, arg3, arg4, arg5, arg6)
}