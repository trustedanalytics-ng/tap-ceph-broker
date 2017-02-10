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
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gocraft/web"
	"github.com/golang/mock/gomock"

	"github.com/trustedanalytics-ng/tap-ceph-broker/client"
)

type MockPack struct {
	osMock *MockOS
}

func prepareMocksAndClient(t *testing.T) (mockCtrl *gomock.Controller, c Context, mocks MockPack, client client.CephBroker) {
	mockCtrl = gomock.NewController(t)
	mocks = MockPack{
		osMock: NewMockOS(mockCtrl),
	}
	c = Context{
		OS: mocks.osMock,
	}
	router := SetupRouter(&c)
	client = getCatalogClient(router, t)
	return
}

func getCatalogClient(router *web.Router, t *testing.T) client.CephBroker {
	const user = "user"
	const password = "password"

	os.Setenv("CEPH_BROKER_USER", user)
	os.Setenv("CEPH_BROKER_PASS", password)

	testServer := httptest.NewServer(router)
	catalogClient, err := client.NewCephBrokerBasicAuth(testServer.URL, user, password)
	if err != nil {
		t.Fatal("Catalog client error: ", err)
	}
	return catalogClient
}
