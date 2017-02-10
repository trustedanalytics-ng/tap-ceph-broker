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
	"strings"

	"github.com/gocraft/web"
	"github.com/trustedanalytics-ng/tap-ceph-broker/model"
	commonHttp "github.com/trustedanalytics-ng/tap-go-common/http"
)

const rbdPath = "/usr/bin/rbd"

var errNotFound = errors.New("not found")

func validateSize(size uint64) error {
	if size == 0 {
		return errors.New("rbd size cannot be null")
	}
	return nil
}

func validateImageName(name string) error {
	if len(name) == 0 {
		return errors.New("rbd image name is empty")
	}
	return nil
}

func validateFileSystem(input string) error {
	allowedFS := []string{model.EXT4, model.XFS}
	for _, fs := range allowedFS {
		if fs == input {
			return nil
		}
	}
	return fmt.Errorf("file system %q is not allowed", input)
}

func validateRBD(rbd model.RBD) error {
	if err := validateSize(rbd.Size); err != nil {
		return err
	}
	if err := validateFileSystem(rbd.FileSystem); err != nil {
		return err
	}
	return validateImageName(rbd.ImageName)
}

func rbdNotFound(message string) bool {
	const notFound = "NO SUCH FILE"
	return strings.Contains(strings.ToUpper(message), notFound)
}

func (c *Context) rbdCreate(name string, size uint64) error {
	_, err := c.OS.ExecuteCommand(rbdPath, "create", name, fmt.Sprintf("--size=%d", size), "--image-feature=layering")
	return err
}

func (c *Context) rbdMap(name string) (string, error) {
	out, err := c.OS.ExecuteCommand(rbdPath, "map", name)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func (c *Context) rbdUnmap(name string) error {
	_, err := c.OS.ExecuteCommand(rbdPath, "unmap", name)
	return err
}

func (c *Context) rbdRemove(name string) error {
	if output, err := c.OS.ExecuteCommandCombinedOutput(rbdPath, "remove", name); err != nil {
		if rbdNotFound(string(output)) {
			return errNotFound
		}
		return err
	}
	return nil
}

func (c *Context) formatDevice(device string, fs string) error {
	_, err := c.OS.ExecuteCommand("/sbin/mkfs."+fs, device)
	return err
}

func (c *Context) createAndFormatRBD(input model.RBD) (model.RBD, error) {
	if err := c.rbdCreate(input.ImageName, input.Size); err != nil {
		return model.RBD{}, fmt.Errorf("cannot create RBD image with name %q and size %d: %v", input.ImageName, input.Size, err)
	}
	device, err := c.rbdMap(input.ImageName)
	if err != nil {
		return model.RBD{}, fmt.Errorf("cannot map RBD image %q: %v", input.ImageName, err)
	}
	if err = c.formatDevice(device, input.FileSystem); err != nil {
		return model.RBD{}, fmt.Errorf("cannot format device %q: %v", device, err)
	}
	if err = c.rbdUnmap(input.ImageName); err != nil {
		return model.RBD{}, fmt.Errorf("cannot unmap RBD image %q: %v", input.ImageName, err)
	}

	return input, nil
}

// CreateRBD creates and formats RBD
func (c *Context) CreateRBD(rw web.ResponseWriter, req *web.Request) {
	input := model.RBD{}
	err := commonHttp.ReadJson(req, &input)
	if err != nil {
		commonHttp.Respond400(rw, err)
		return
	}
	if err = validateRBD(input); err != nil {
		commonHttp.Respond400(rw, err)
		return
	}

	rbd, err := c.createAndFormatRBD(input)
	if err != nil {
		commonHttp.Respond500(rw, err)
		return
	}

	if err = commonHttp.WriteJson(rw, rbd, http.StatusOK); err != nil {
		err = fmt.Errorf("cannot parse response: %v", err)
		commonHttp.Respond500(rw, err)
		return
	}
}

// DeleteRBD deletes RBD
func (c *Context) DeleteRBD(rw web.ResponseWriter, req *web.Request) {
	name := req.PathParams["imageName"]

	if err := validateImageName(name); err != nil {
		commonHttp.Respond400(rw, err)
		return
	}

	if err := c.rbdRemove(name); err != nil {
		errNew := fmt.Errorf("cannot delete RBD: %v", err)
		if err == errNotFound {
			commonHttp.Respond404(rw, errNew)
			return
		}
		commonHttp.Respond500(rw, errNew)
		return
	}

	rw.WriteHeader(http.StatusNoContent)
}
