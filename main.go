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

package main

import (
	"os"

	"github.com/gocraft/web"

	"github.com/trustedanalytics/tap-ceph-broker/api"
	innerOS "github.com/trustedanalytics/tap-ceph-broker/os"
	httpGoCommon "github.com/trustedanalytics/tap-go-common/http"
	commonLogger "github.com/trustedanalytics/tap-go-common/logger"
)

const (
	sslCertLocationEnvVarName = "CEPH_BROKER_SSL_CERT_LOCATION"
	sslKeyLocationEnvVarName  = "CEPH_BROKER_SSL_KEY_LOCATION"
)

var logger, _ = commonLogger.InitLogger("main")

func main() {
	sos := innerOS.StandardOS{}
	context := api.Context{OS: sos}

	router := api.SetupRouter(&context)

	startServer(router)
}

func startServer(router *web.Router) {
	cert := os.Getenv(sslCertLocationEnvVarName)
	key := os.Getenv(sslKeyLocationEnvVarName)
	if cert == "" || key == "" {
		logger.Fatalf("Only SSL protocol is supported. You need to provide environment variables %q and %q.", sslCertLocationEnvVarName, sslKeyLocationEnvVarName)
	} else {
		httpGoCommon.StartServerTLS(cert, key, router)
	}
}
