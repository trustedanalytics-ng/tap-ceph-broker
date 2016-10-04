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
	"os"

	"github.com/gocraft/web"

	"github.com/trustedanalytics/tap-go-common/util"
)

// BasicAuthorizeMiddleware authorizes user with credentials taken from envs
func (c *Context) BasicAuthorizeMiddleware(rw web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc) {
	logger.Info("Trying to access url ", req.URL.Path, " by BasicAuthorize")
	username, password, isOK := req.BasicAuth()
	if !isOK || username != os.Getenv("CEPH_BROKER_USER") || password != os.Getenv("CEPH_BROKER_PASS") {
		logger.Info("EnforceAuthMiddleware - BasicAuth: Invalid Basic Auth credentials")
		util.RespondUnauthorized(rw)
		return
	}
	logger.Info("EnforceAuthMiddleware - BasicAuth: User authenticated as ", username)
	next(rw, req)
}
