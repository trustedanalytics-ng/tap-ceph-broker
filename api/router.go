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
	"net/http"

	"github.com/gocraft/web"

	commonHttp "github.com/trustedanalytics/tap-go-common/http"
)

func SetupRouter(context *Context) *web.Router {
	router := web.New(*context)
	router.Middleware(web.LoggerMiddleware)

	router.Get("/healthz", context.GetHealthz)

	apiRouter := router.Subrouter(*context, "/api/v1")
	route(apiRouter, context)
	v1AliasRouter := router.Subrouter(*context, "/api/v1.0")
	route(v1AliasRouter, context)

	router.Get("/", context.Index)
	router.Error(context.Error)

	return router
}

func route(router *web.Router, context *Context) {
	router.Middleware(context.BasicAuthorizeMiddleware)

	router.Post("/rbd", (*context).CreateRBD)
	router.Delete("/rbd/:imageName", (*context).DeleteRBD)

	router.Get("/lock", (*context).ListLocks)
	router.Delete("/lock/:imageName/:lockName/:locker", (*context).DeleteLock)
}

func (c *Context) Index(rw web.ResponseWriter, req *web.Request) {
	commonHttp.WriteJson(rw, "I'm OK", http.StatusOK)
}

func (c *Context) Error(rw web.ResponseWriter, r *web.Request, err interface{}) {
	logger.Errorf("Respond500: reason: %v", err)
	rw.WriteHeader(http.StatusInternalServerError)
}
