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
	"os/exec"
	"strings"

	"github.com/gocraft/web"
	"github.com/trustedanalytics/tap-ceph-broker/model"
	"github.com/trustedanalytics/tap-go-common/util"
)

func filter_nonempty_lines(lines []string) []string {
	nonempty_lines := []string{}
	for _, l := range lines {
		if len(strings.TrimSpace(l)) > 0 {
			nonempty_lines = append(nonempty_lines, l)
		}
	}
	return nonempty_lines
}

func listImages() ([]string, error) {
	logger.Debug("listImages")
y	output, err := exec.Command(rbdPath, "list").CombinedOutput()
	if err != nil {
		return []string{}, err
	}
	logger.Debug("listImages: rbd output: ", string(output))
	image_lines := filter_nonempty_lines(strings.Split(string(output), "\n"))
	return image_lines, nil
}

func lockListForImage(imageName string) ([]model.Lock, error) {
	logger.Debug("lockListForImage: getting locks for image", imageName)
	out := []model.Lock{}
	output, err := exec.Command(rbdPath, "lock", "list", imageName).CombinedOutput()
	if err != nil {
		return out, err
	}
	logger.Debug("lockListForImage: rbd output: ", string(output))
	lock_lines := filter_nonempty_lines(strings.Split(string(output), "\n"))

	for i, nonempty_lock_line := range lock_lines {
		if i < 2 {
			continue // skip header and 'There is 1 exclusive lock on this image.' line
		}

		fields := strings.Fields(nonempty_lock_line)
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

func allLocks() ([]model.Lock, error) {
	logger.Debug("allLocks")
	locks := []model.Lock{}
	images, err := listImages()
	logger.Info("allLocks: images", images)
	if err != nil {
		return locks, err
	}
	for _, image := range images {
		logger.Info("allLocks: getting locks for image", image)
		imageLocks, err := lockListForImage(image)
		if err != nil {
			return locks, err
		}
		for _, imageLock := range imageLocks {
			locks = append(locks, imageLock)
		}
	}
	return locks, nil
}

func removeLock(lock model.Lock) error {
	logger.Info("removeLock:", lock)
	output, err := exec.Command(rbdPath, "lock", "remove", lock.ImageName, lock.LockName, lock.Locker).CombinedOutput()
	if err != nil {
		logger.Info("removeLock: FAILED:", err, string(output))
		return err
	}
	logger.Info("removeLock: SUCCESS.")
	return nil
}

func (c *Context) ListLocks(rw web.ResponseWriter, req *web.Request) {
	locks, err := allLocks()
	if err != nil {
		util.Respond500(rw, err)
		return
	}

	if err = util.WriteJson(rw, locks, http.StatusOK); err != nil {
		err = fmt.Errorf("cannot parse response: %v", err)
		util.Respond500(rw, err)
		return
	}

}

func (c *Context) DeleteLock(rw web.ResponseWriter, req *web.Request) {
	imageName := req.PathParams["imageName"]
	lockName := req.PathParams["lockName"]
	locker := req.PathParams["locker"]

	lock := model.Lock{LockName: strings.Replace(lockName, "\"", "", -1), ImageName: imageName, Locker: locker}

	err := removeLock(lock)
	if err != nil {
		util.Respond500(rw, err)
		return
	}

	rw.WriteHeader(http.StatusNoContent)
}
