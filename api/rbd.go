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
	"errors"
	"fmt"
	"net/http"
	"os/exec"
	"strings"

	"github.com/gocraft/web"
	"github.com/trustedanalytics/tap-ceph-broker/model"
	"github.com/trustedanalytics/tap-go-common/util"
)

const rbdPath = "/usr/bin/rbd"

var errNotFound = errors.New("not found")

func validateImageName(name string) error {
	if len(name) == 0 {
		return errors.New("rbd image name is empty")
	}
	return nil
}

func validateRBD(rbd model.RBD) error {
	if rbd.Size == 0 {
		return errors.New("rbd size cannot be null")
	}
	return validateImageName(rbd.ImageName)
}

func rbdNotFound(message string) bool {
	const notFound = "NO SUCH FILE"
	return strings.Contains(strings.ToUpper(message), notFound)
}

func rbdCreate(name string, size uint64) error {
	_, err := exec.Command(rbdPath, "create", name, fmt.Sprintf("--size=%d", size), "--image-feature=layering").Output()
	return err
}

func rbdMap(name string) (string, error) {
	out, err := exec.Command(rbdPath, "map", name).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func rbdUnmap(name string) error {
	_, err := exec.Command(rbdPath, "unmap", name).Output()
	return err
}

func rbdRemove(name string) error {
	if output, err := exec.Command(rbdPath, "remove", name).CombinedOutput(); err != nil {
		if rbdNotFound(string(output)) {
			return errNotFound
		}
		return err
	}
	return nil
}

func formatDevice(device string) error {
	_, err := exec.Command("mkfs.xfs", device).Output()
	return err
}

func createAndFormatRBD(input model.RBD) (model.RBD, error) {
	if err := rbdCreate(input.ImageName, input.Size); err != nil {
		return model.RBD{}, fmt.Errorf("cannot create RBD image with name %q and size %d: %v", input.ImageName, input.Size, err)
	}
	device, err := rbdMap(input.ImageName)
	if err != nil {
		return model.RBD{}, fmt.Errorf("cannot map RBD image %q: %v", input.ImageName, err)
	}
	if err = formatDevice(device); err != nil {
		return model.RBD{}, fmt.Errorf("cannot format device %q: %v", device, err)
	}
	if err = rbdUnmap(input.ImageName); err != nil {
		return model.RBD{}, fmt.Errorf("cannot unmap RBD image %q: %v", input.ImageName, err)
	}

	return input, nil
}

// CreateRBD creates and formats RBD
func (c *Context) CreateRBD(rw web.ResponseWriter, req *web.Request) {
	input := model.RBD{}
	err := util.ReadJson(req, &input)
	if err != nil {
		util.Respond400(rw, err)
		return
	}
	if err = validateRBD(input); err != nil {
		util.Respond400(rw, err)
		return
	}

	rbd, err := createAndFormatRBD(input)
	if err != nil {
		util.Respond500(rw, err)
		return
	}

	if err = util.WriteJson(rw, rbd, http.StatusOK); err != nil {
		err = fmt.Errorf("cannot parse response: %v", err)
		util.Respond500(rw, err)
		return
	}
}

// DeleteRBD deletes RBD
func (c *Context) DeleteRBD(rw web.ResponseWriter, req *web.Request) {
	name := req.PathParams["imageName"]

	if err := validateImageName(name); err != nil {
		util.Respond400(rw, err)
		return
	}

	if err := rbdRemove(name); err != nil {
		errNew := fmt.Errorf("cannot delete RBD: %v", err)
		if err == errNotFound {
			util.Respond404(rw, errNew)
			return
		}
		util.Respond500(rw, errNew)
		return
	}

	rw.WriteHeader(http.StatusNoContent)
}
