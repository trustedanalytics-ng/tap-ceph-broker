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

	"github.com/gocraft/web"

	"github.com/trustedanalytics/tap-ceph-broker/model"
	commonHttp "github.com/trustedanalytics/tap-go-common/http"
)

func filterNonemptyLines(input string) []string {
	lines := strings.Split(string(input), "\n")
	nonemptyLines := []string{}
	for _, l := range lines {
		// FIXME - workaround for Ceph warnings due to invalid ceph.conf
		if strings.HasPrefix(l, "warning:") {
			logger.Info("Skipping line:", l)
			continue
		}

		if len(strings.TrimSpace(l)) > 0 {
			nonemptyLines = append(nonemptyLines, l)
		}
	}
	return nonemptyLines
}

func (c *Context) listImages() ([]string, error) {
	logger.Debug("listImages")
	output, err := c.OS.ExecuteCommand(rbdPath, "list")
	if err != nil {
		logger.Errorf("listImages: FAILED: %v", err)
		return []string{}, err
	}
	logger.Debug("listImages: rbd output: ", string(output))
	imageLines := filterNonemptyLines(output)
	return imageLines, nil
}

func (c *Context) lockListForImage(imageName string) ([]model.Lock, error) {
	logger.Debug("lockListForImage: getting locks for image", imageName)
	out := []model.Lock{}
	output, err := c.OS.ExecuteCommand(rbdPath, "lock", "list", imageName)
	if err != nil {
		logger.Errorf("lockListForImage: FAILED: %v", err)
		return out, err
	}
	logger.Debug("lockListForImage: rbd output: ", string(output))
	lockLines := filterNonemptyLines(output)

	for i, nonemptyLockLine := range lockLines {
		if i < 2 {
			continue // skip header and 'There is 1 exclusive lock on this image.' line
		}

		fields := strings.Fields(nonemptyLockLine)
		if len(fields) < 3 {
			continue
		}

		lock := model.Lock{LockName: fields[1], ImageName: imageName, Locker: fields[0], Address: fields[2]}
		/*
		   There is 1 exclusive lock on this image.
		   Locker      ID                                                         Address
		   client.4239 kubelet_lock_magic_compute-worker-1.instance.cluster.local 10.0.2.190:0/3340152652
		*/

		out = append(out, lock)
	}
	logger.Info("locks: ", out)
	return out, nil
}

func (c *Context) allLocks() ([]model.Lock, error) {
	logger.Debug("allLocks")
	locks := []model.Lock{}
	images, err := c.listImages()
	logger.Info("allLocks: images", images)
	if err != nil {
		return locks, err
	}
	for _, image := range images {
		logger.Info("allLocks: getting locks for image", image)
		imageLocks, err := c.lockListForImage(image)
		if err != nil {
			return locks, err
		}
		locks = append(locks, imageLocks...)
	}
	return locks, nil
}

func (c *Context) removeLock(lock model.Lock) error {
	logger.Info("removeLock:", lock)
	output, err := c.OS.ExecuteCommandCombinedOutput(rbdPath, "lock", "remove", lock.ImageName, lock.LockName, lock.Locker)
	if err != nil {
		logger.Error("removeLock: FAILED:", err, string(output))
		return err
	}
	logger.Info("removeLock: SUCCESS.")
	return nil
}

func (c *Context) ListLocks(rw web.ResponseWriter, req *web.Request) {
	locks, err := c.allLocks()
	if err != nil {
		commonHttp.Respond500(rw, err)
		return
	}

	if err = commonHttp.WriteJson(rw, locks, http.StatusOK); err != nil {
		err = fmt.Errorf("cannot parse response: %v", err)
		commonHttp.Respond500(rw, err)
		return
	}

}

func (c *Context) DeleteLock(rw web.ResponseWriter, req *web.Request) {
	imageName := req.PathParams["imageName"]
	lockName := req.PathParams["lockName"]
	locker := req.PathParams["locker"]

	lock := model.Lock{LockName: strings.Replace(lockName, "\"", "", -1), ImageName: imageName, Locker: locker}

	err := c.removeLock(lock)
	if err != nil {
		commonHttp.Respond500(rw, err)
		return
	}

	rw.WriteHeader(http.StatusNoContent)
}
