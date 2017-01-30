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
package client

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewCephBrokerBasicAuth(t *testing.T) {
	Convey("Test NewCephBrokerBasicAuth with all arguments", t, func() {
		cephClient, err := NewCephBrokerBasicAuth("address", "username", "password")
		Convey("should return nil error", func() {
			So(err, ShouldBeNil)
		})
		Convey("Client address should be set to defined address", func() {
			So(cephClient.Address, ShouldEqual, "address")
		})
		Convey("Client username should be set to defined username", func() {
			So(cephClient.Username, ShouldEqual, "username")
		})
		Convey("Client password should be set to defined password", func() {
			So(cephClient.Password, ShouldEqual, "password")
		})
	})
}
