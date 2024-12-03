// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/adettelle/go-keeper/internal/server/api (interfaces: IFileRepo)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	repo "github.com/adettelle/go-keeper/internal/repo"
	gomock "github.com/golang/mock/gomock"
)

// MockIFileRepo is a mock of IFileRepo interface.
type MockIFileRepo struct {
	ctrl     *gomock.Controller
	recorder *MockIFileRepoMockRecorder
}

// MockIFileRepoMockRecorder is the mock recorder for MockIFileRepo.
type MockIFileRepoMockRecorder struct {
	mock *MockIFileRepo
}

// NewMockIFileRepo creates a new mock instance.
func NewMockIFileRepo(ctrl *gomock.Controller) *MockIFileRepo {
	mock := &MockIFileRepo{ctrl: ctrl}
	mock.recorder = &MockIFileRepoMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIFileRepo) EXPECT() *MockIFileRepoMockRecorder {
	return m.recorder
}

// AddFile mocks base method.
func (m *MockIFileRepo) AddFile(arg0 context.Context, arg1, arg2, arg3, arg4, arg5 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddFile", arg0, arg1, arg2, arg3, arg4, arg5)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddFile indicates an expected call of AddFile.
func (mr *MockIFileRepoMockRecorder) AddFile(arg0, arg1, arg2, arg3, arg4, arg5 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddFile", reflect.TypeOf((*MockIFileRepo)(nil).AddFile), arg0, arg1, arg2, arg3, arg4, arg5)
}

// DeleteFileByTitle mocks base method.
func (m *MockIFileRepo) DeleteFileByTitle(arg0 context.Context, arg1, arg2 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteFileByTitle", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteFileByTitle indicates an expected call of DeleteFileByTitle.
func (mr *MockIFileRepoMockRecorder) DeleteFileByTitle(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteFileByTitle", reflect.TypeOf((*MockIFileRepo)(nil).DeleteFileByTitle), arg0, arg1, arg2)
}

// FileExists mocks base method.
func (m *MockIFileRepo) FileExists(arg0 context.Context, arg1, arg2 string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FileExists", arg0, arg1, arg2)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FileExists indicates an expected call of FileExists.
func (mr *MockIFileRepoMockRecorder) FileExists(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FileExists", reflect.TypeOf((*MockIFileRepo)(nil).FileExists), arg0, arg1, arg2)
}

// GetAllFiles mocks base method.
func (m *MockIFileRepo) GetAllFiles(arg0 context.Context, arg1 string) ([]repo.FileToGet, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAllFiles", arg0, arg1)
	ret0, _ := ret[0].([]repo.FileToGet)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAllFiles indicates an expected call of GetAllFiles.
func (mr *MockIFileRepoMockRecorder) GetAllFiles(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllFiles", reflect.TypeOf((*MockIFileRepo)(nil).GetAllFiles), arg0, arg1)
}

// GetFileCoudIDByTitle mocks base method.
func (m *MockIFileRepo) GetFileCoudIDByTitle(arg0 context.Context, arg1, arg2 string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetFileCoudIDByTitle", arg0, arg1, arg2)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetFileCoudIDByTitle indicates an expected call of GetFileCoudIDByTitle.
func (mr *MockIFileRepoMockRecorder) GetFileCoudIDByTitle(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetFileCoudIDByTitle", reflect.TypeOf((*MockIFileRepo)(nil).GetFileCoudIDByTitle), arg0, arg1, arg2)
}

// UpdateFile mocks base method.
func (m *MockIFileRepo) UpdateFile(arg0 context.Context, arg1 string, arg2, arg3 *string, arg4 int) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateFile", arg0, arg1, arg2, arg3, arg4)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateFile indicates an expected call of UpdateFile.
func (mr *MockIFileRepoMockRecorder) UpdateFile(arg0, arg1, arg2, arg3, arg4 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateFile", reflect.TypeOf((*MockIFileRepo)(nil).UpdateFile), arg0, arg1, arg2, arg3, arg4)
}
