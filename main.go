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

package main

import (
	"github.com/gocraft/web"

	"os"

	"github.com/trustedanalytics/tap-ceph-broker/api"
	httpGoCommon "github.com/trustedanalytics/tap-go-common/http"
)

func main() {
	context := api.Context{}

	router := web.New(context)
	router.Middleware(web.LoggerMiddleware)

	router.Get("/healthz", context.GetHealthz)

	apiRouter := router.Subrouter(context, "/api/v1")
	route(apiRouter, &context)
	v1AliasRouter := router.Subrouter(context, "/api/v1.0")
	route(v1AliasRouter, &context)

	if os.Getenv("CEPH_BROKER_SSL_CERT_FILE_LOCATION") != "" {
		httpGoCommon.StartServerTLS(os.Getenv("CEPH_BROKER_SSL_CERT_FILE_LOCATION"),
			os.Getenv("CEPH_BROKER_SSL_CERT_FILE_LOCATION"), router)
	} else {
		httpGoCommon.StartServer(router)
	}
}

func route(router *web.Router, context *api.Context) {
	router.Middleware(context.BasicAuthorizeMiddleware)

	router.Post("/rbd", (*context).CreateRBD)
	router.Delete("/rbd/:imageName", (*context).DeleteRBD)
}
