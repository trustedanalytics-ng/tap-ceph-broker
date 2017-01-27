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

package api

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/trustedanalytics/tap-ceph-broker/model"
)

const (
	sampleImage1   = "sampleImage1"
	sampleLocker1  = "client.4175"
	sampleID1      = "kubelet_lock_magic_compute-worker-2.instance"
	sampleAddress1 = "10.0.2.153:0/3412117426"

	sampleImage2   = "sampleImage2"
	sampleLocker2  = "client.4275"
	sampleID2      = "kubelet_lock_magic_compute-worker-2.instance"
	sampleAddress2 = "10.0.2.154:0/3412117427"
)

func TestListLocks(t *testing.T) {
	tests := []struct {
		testDescription string
		images          []string
		imageLocks      []string
		result          []model.Lock
	}{
		{
			testDescription: "no images",
			images:          []string{},
			imageLocks:      []string{},
			result:          []model.Lock{},
		},
		{
			testDescription: "wrong list lock command output",
			images:          []string{"sampleImage1"},
			imageLocks:      []string{"some wrong output"},
			result:          []model.Lock{},
		},
		{
			testDescription: "one image has one lock",
			images:          []string{sampleImage1},
			imageLocks:      []string{createLockList(createLockRow(sampleLocker1, sampleID1, sampleAddress1))},
			result: []model.Lock{
				model.Lock{
					ImageName: sampleImage1,
					LockName:  sampleID1,
					Locker:    sampleLocker1,
					Address:   sampleAddress1,
				},
			},
		},
		{
			testDescription: "one image has two locks",
			images:          []string{sampleImage1},
			imageLocks: []string{
				createLockList(
					createLockRow(sampleLocker1, sampleID1, sampleAddress1),
					createLockRow(sampleLocker2, sampleID2, sampleAddress2)),
			},
			result: []model.Lock{
				model.Lock{
					ImageName: sampleImage1,
					LockName:  sampleID1,
					Locker:    sampleLocker1,
					Address:   sampleAddress1,
				},
				model.Lock{
					ImageName: sampleImage1,
					LockName:  sampleID2,
					Locker:    sampleLocker2,
					Address:   sampleAddress2,
				},
			},
		},
		{
			testDescription: "two images have one lock for each one",
			images:          []string{sampleImage1, sampleImage2},
			imageLocks: []string{
				createLockList(createLockRow(sampleLocker1, sampleID1, sampleAddress1)),
				createLockList(createLockRow(sampleLocker2, sampleID2, sampleAddress2)),
			},
			result: []model.Lock{
				model.Lock{
					ImageName: sampleImage1,
					LockName:  sampleID1,
					Locker:    sampleLocker1,
					Address:   sampleAddress1,
				},
				model.Lock{
					ImageName: sampleImage2,
					LockName:  sampleID2,
					Locker:    sampleLocker2,
					Address:   sampleAddress2,
				},
			},
		},
	}

	Convey("Testing ListLocks", t, func() {
		mockCtrl, _, mock, client := prepareMocksAndClient(t)

		for i, test := range tests {
			Convey(fmt.Sprintf("For test case %d", i), func() {
				images := strings.Join(test.images, "\n")
				images = images + "\n"
				asserts := []*gomock.Call{mock.osMock.EXPECT().ExecuteCommand(rbdPath, "list").Return(images, nil)}
				for i := 0; i < len(test.images); i++ {
					asserts = append(asserts, mock.osMock.EXPECT().ExecuteCommand(rbdPath, "lock", "list", test.images[i]).Return(test.imageLocks[i], nil))
				}
				gomock.InOrder(
					asserts...,
				)

				locks, status, err := client.ListLocks()

				So(status, ShouldEqual, http.StatusOK)
				So(err, ShouldBeNil)
				So(locks, ShouldResemble, test.result)
			})
		}

		Convey("When list command goes wrong", func() {
			mock.osMock.EXPECT().ExecuteCommand(rbdPath, "list").Return("", fmt.Errorf("some error"))

			_, status, err := client.ListLocks()

			So(status, ShouldEqual, http.StatusInternalServerError)
			So(err, ShouldNotBeNil)
		})

		Reset(func() {
			mockCtrl.Finish()
		})
	})
}

func createLockList(rows ...string) string {
	result := []string{
		fmt.Sprintf("Locker\tID\tAddress"),
		fmt.Sprintf("There is %d exclusive lock on this image.", len(rows)),
	}

	for _, row := range rows {
		result = append(result, row)
	}

	return strings.Join(result, "\n")
}

func createLockRow(locker, id, address string) string {
	return fmt.Sprintf("%s %s %s", locker, id, address)
}

func TestDeleteLock(t *testing.T) {
	Convey("Testing DeleteLock", t, func() {
		mockCtrl, _, mock, client := prepareMocksAndClient(t)

		Convey("When deleting lock exists", func() {
			lock := model.Lock{ImageName: sampleImage1, LockName: sampleID1, Locker: sampleLocker1, Address: sampleAddress1}
			mock.osMock.EXPECT().ExecuteCommand(rbdPath, "lock", "remove", lock.ImageName, lock.LockName, lock.Locker).Return("", nil)

			status, err := client.DeleteLock(lock)

			So(status, ShouldEqual, http.StatusNoContent)
			So(err, ShouldBeNil)
		})

		Convey("When deleting lock return error", func() {
			lock := model.Lock{ImageName: sampleImage1, LockName: sampleID1, Locker: sampleLocker1, Address: sampleAddress1}
			mock.osMock.EXPECT().ExecuteCommand(rbdPath, "lock", "remove", lock.ImageName, lock.LockName, lock.Locker).Return("", fmt.Errorf("some error"))

			status, err := client.DeleteLock(lock)

			So(status, ShouldEqual, http.StatusInternalServerError)
			So(err, ShouldNotBeNil)
		})

		Reset(func() {
			mockCtrl.Finish()
		})
	})
}
