/**
 * Copyright (c) 2017 Intel Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
// Automatically generated by MockGen. DO NOT EDIT!
// Source: os/os.go

package api

import (
	gomock "github.com/golang/mock/gomock"
)

// Mock of OS interface
type MockOS struct {
	ctrl     *gomock.Controller
	recorder *_MockOSRecorder
}

// Recorder for MockOS (not exported)
type _MockOSRecorder struct {
	mock *MockOS
}

func NewMockOS(ctrl *gomock.Controller) *MockOS {
	mock := &MockOS{ctrl: ctrl}
	mock.recorder = &_MockOSRecorder{mock}
	return mock
}

func (_m *MockOS) EXPECT() *_MockOSRecorder {
	return _m.recorder
}

func (_m *MockOS) Command(name string, arg ...string) (string, error) {
	_s := []interface{}{name}
	for _, _x := range arg {
		_s = append(_s, _x)
	}
	ret := _m.ctrl.Call(_m, "Command", _s...)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockOSRecorder) Command(arg0 interface{}, arg1 ...interface{}) *gomock.Call {
	_s := append([]interface{}{arg0}, arg1...)
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Command", _s...)
}
