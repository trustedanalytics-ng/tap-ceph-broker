/**
 * Copyright (c) 2016 Intel Corporation
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

package api

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/trustedanalytics/tap-ceph-broker/model"
)

func TestCreateRBD(t *testing.T) {
	Convey("Testing CreateRBD", t, func() {
		mockCtrl, _, mock, client := prepareMocksAndClient(t)
		sampleName := "sampleRBD"
		var sampleSize uint64 = 1000
		sampleFS := model.XFS
		device := model.RBD{ImageName: sampleName, Size: sampleSize, FileSystem: sampleFS}
		sampleDevice := "/dev/rbd1"

		Convey("When os commands are executed correctly", func() {
			gomock.InOrder(
				mock.osMock.EXPECT().ExecuteCommand(rbdPath, "create", sampleName, fmt.Sprintf("--size=%d", sampleSize), "--image-feature=layering").Return("", nil),
				mock.osMock.EXPECT().ExecuteCommand(rbdPath, "map", sampleName).Return(sampleDevice, nil),
				mock.osMock.EXPECT().ExecuteCommand("/sbin/mkfs."+sampleFS, sampleDevice).Return("", nil),
				mock.osMock.EXPECT().ExecuteCommand(rbdPath, "unmap", sampleName).Return("", nil),
			)

			status, err := client.CreateRBD(device)

			So(status, ShouldEqual, http.StatusOK)
			So(err, ShouldBeNil)
		})

		Convey("When format command goes wrong", func() {
			gomock.InOrder(
				mock.osMock.EXPECT().ExecuteCommand(rbdPath, "create", sampleName, fmt.Sprintf("--size=%d", sampleSize), "--image-feature=layering").Return("", nil),
				mock.osMock.EXPECT().ExecuteCommand(rbdPath, "map", sampleName).Return(sampleDevice, nil),
				mock.osMock.EXPECT().ExecuteCommand("/sbin/mkfs."+sampleFS, sampleDevice).Return("", fmt.Errorf("some error!")),
			)

			status, err := client.CreateRBD(device)

			So(status, ShouldEqual, http.StatusInternalServerError)
			So(err, ShouldNotBeNil)
		})

		Reset(func() {
			mockCtrl.Finish()
		})
	})
}

func TestDeleteRBD(t *testing.T) {
	Convey("Testing DeleteRBD", t, func() {
		mockCtrl, _, mock, client := prepareMocksAndClient(t)
		sampleName := "sampleRBD"

		Convey("When os commands are executed correctly", func() {
			mock.osMock.EXPECT().ExecuteCommandCombinedOutput(rbdPath, "remove", sampleName).Return("", nil)

			status, err := client.DeleteRBD(sampleName)

			So(status, ShouldEqual, http.StatusNoContent)
			So(err, ShouldBeNil)
		})

		Convey("When format command goes wrong", func() {
			mock.osMock.EXPECT().ExecuteCommandCombinedOutput(rbdPath, "remove", sampleName).Return("", fmt.Errorf("some error!"))

			status, err := client.DeleteRBD(sampleName)

			So(status, ShouldEqual, http.StatusInternalServerError)
			So(err, ShouldNotBeNil)
		})

		Convey("When empty name is passed", func() {
			status, err := client.DeleteRBD("")

			So(status, ShouldEqual, http.StatusNotFound)
			So(err, ShouldNotBeNil)
		})

		Reset(func() {
			mockCtrl.Finish()
		})
	})
}

func TestValidateRBD(t *testing.T) {
	testCases := []struct {
		rbd     model.RBD
		isError bool
	}{
		{model.RBD{ImageName: "", Size: 100, FileSystem: model.XFS}, true},
		{model.RBD{ImageName: "", Size: 0, FileSystem: model.XFS}, true},
		{model.RBD{ImageName: "someimage", Size: 0, FileSystem: model.EXT4}, true},
		{model.RBD{ImageName: "someimage", Size: 200, FileSystem: "wrongFS"}, true},
		{model.RBD{ImageName: "some image", Size: 100, FileSystem: model.EXT4}, false},
		{model.RBD{ImageName: "some image", Size: 1024 * 1024, FileSystem: model.XFS}, false},
		{model.RBD{ImageName: "some image_123", Size: 1024 * 1024 * 1000 * 9, FileSystem: model.XFS}, false},
	}

	for _, tc := range testCases {
		err := validateRBD(tc.rbd)
		if (err == nil && tc.isError) || (err != nil && !tc.isError) {
			t.Errorf("validateRBD(%v) returned error: %v; error expected: %v", tc.rbd, err != nil, tc.isError)
		}
	}
}

func TestRbdNotFound(t *testing.T) {
	testCases := []struct {
		message string
		output  bool
	}{
		{"rbd: delete error: (2) No such file or directory", true},
		{"No such file or directory", true},
		{"No such file", true},
		{"no such file", true},
		{"NO SUCH FILE", true},
		{"there is such file", false},
		{"OK", false},
		{"file removed", false},
		{"", false},
	}

	for _, tc := range testCases {
		output := rbdNotFound(tc.message)
		if output != tc.output {
			t.Errorf("rbdNotFound(%s) = %v; want %v", tc.message, output, tc.output)
		}
	}
}
