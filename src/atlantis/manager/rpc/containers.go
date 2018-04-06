/* Copyright 2014 Ooyala, Inc. All rights reserved.
 *
 * This file is licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
 * except in compliance with the License. You may obtain a copy of the License at
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License is
 * distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and limitations under the License.
 */

package rpc

import (
	. "atlantis/common"
	"atlantis/manager/datamodel"
	. "atlantis/manager/rpc/types"
	"atlantis/manager/supervisor"
	. "atlantis/supervisor/rpc/types"
	"errors"
	"fmt"
	"sort"
)

type GetContainerExecutor struct {
	arg   ManagerGetContainerArg
	reply *ManagerGetContainerReply
}

func (e *GetContainerExecutor) Request() interface{} {
	return e.arg
}

func (e *GetContainerExecutor) Result() interface{} {
	return e.reply
}

func (e *GetContainerExecutor) Description() string {
	return "[" + e.arg.ManagerAuthArg.User + "] " + e.arg.ContainerID
}

func (e *GetContainerExecutor) Execute(t *Task) (err error) {
	if e.arg.ContainerID == "" {
		return errors.New("Container ID is empty")
	}
	instance, err := datamodel.GetInstance(e.arg.ContainerID)
	if err != nil {
		e.reply.Status = StatusError
		return err
	}
	var ihReply *SupervisorGetReply
	ihReply, err = supervisor.Get(instance.Host, e.arg.ContainerID)
	if err != nil {
		e.reply.Status = StatusError
		return err
	}
	e.reply.Status = ihReply.Status
	ihReply.Container.Host = instance.Host
	e.reply.Container = ihReply.Container
	e.reply.Status = StatusOk
	return err
}

func (e *GetContainerExecutor) Authorize() error {
	return SimpleAuthorize(&e.arg.ManagerAuthArg)
}

func (m *ManagerRPC) GetContainer(arg ManagerGetContainerArg, reply *ManagerGetContainerReply) error {
	return NewTask("GetContainer", &GetContainerExecutor{arg, reply}).Run()
}

type ListContainersExecutor struct {
	arg   ManagerListContainersArg
	reply *ManagerListContainersReply
}

func (e *ListContainersExecutor) Request() interface{} {
	return e.arg
}

func (e *ListContainersExecutor) Result() interface{} {
	return e.reply
}

func (e *ListContainersExecutor) Description() string {
	return fmt.Sprintf("["+e.arg.ManagerAuthArg.User+"] %s @ %s in %s", e.arg.App, e.arg.Sha, e.arg.Env)
}

func (e *ListContainersExecutor) Execute(t *Task) error {
	var err error
	if e.arg.App == "" && e.arg.Sha == "" && e.arg.Env == "" {
		// try to list all instances
		allContainerIDs, err := datamodel.ListAllInstances()
		if err != nil {
			e.reply.Status = StatusError
			return err
		} else {
			e.reply.Status = StatusOk
		}
		// filter by allowed app
		if err := AuthorizeSuperUser(&e.arg.ManagerAuthArg); err == nil {
			// if superuser, show everything
			e.reply.ContainerIDs = allContainerIDs
		} else {
			// else only show what is allowed
			allowedApps := GetAllowedApps(&e.arg.ManagerAuthArg, e.arg.ManagerAuthArg.User)
			e.reply.ContainerIDs = []string{}
			for _, cid := range allContainerIDs {
				if inst, err := datamodel.GetInstance(cid); err == nil && allowedApps[inst.App] {
					e.reply.ContainerIDs = append(e.reply.ContainerIDs, cid)
				}
			}
		}
		sort.Strings(e.reply.ContainerIDs)
		return nil
	}
	if e.arg.App == "" {
		return errors.New("App is empty")
	}
	if e.arg.Sha == "" {
		return errors.New("Sha is empty")
	}
	if e.arg.Env == "" {
		return errors.New("Environment is empty")
	}
	e.reply.ContainerIDs, err = datamodel.ListInstances(e.arg.App, e.arg.Sha, e.arg.Env)
	if err != nil {
		e.reply.Status = StatusError
	} else {
		sort.Strings(e.reply.ContainerIDs)
		e.reply.Status = StatusOk
	}
	return err
}

func (e *ListContainersExecutor) Authorize() error {
	if e.arg.App == "" && e.arg.Sha == "" && e.arg.Env == "" {
		// execute will filter based on allowed apps
		return SimpleAuthorize(&e.arg.ManagerAuthArg)
	}
	return AuthorizeApp(&e.arg.ManagerAuthArg, e.arg.App)
}

func (m *ManagerRPC) ListContainers(arg ManagerListContainersArg, reply *ManagerListContainersReply) error {
	return NewTask("ListContainers", &ListContainersExecutor{arg, reply}).Run()
}

type ListEnvsExecutor struct {
	arg   ManagerListEnvsArg
	reply *ManagerListEnvsReply
}

func (e *ListEnvsExecutor) Request() interface{} {
	return e.arg
}

func (e *ListEnvsExecutor) Result() interface{} {
	return e.reply
}

func (e *ListEnvsExecutor) Description() string {
	if e.arg.App == "" || e.arg.Sha == "" {
		return fmt.Sprintf("["+e.arg.ManagerAuthArg.User+"] %s @ %s", e.arg.App, e.arg.Sha)
	}
	return "[" + e.arg.ManagerAuthArg.User + "] All"
}

func (e *ListEnvsExecutor) Execute(t *Task) error {
	var err error
	if e.arg.App == "" || e.arg.Sha == "" {
		e.reply.Envs, err = datamodel.ListEnvs()
	} else {
		e.reply.Envs, err = datamodel.ListAppEnvs(e.arg.App, e.arg.Sha)
	}
	if err != nil {
		e.reply.Status = StatusError
	} else {
		sort.Strings(e.reply.Envs)
		e.reply.Status = StatusOk
	}
	return err
}

func (e *ListEnvsExecutor) Authorize() error {
	if e.arg.App == "" {
		return SimpleAuthorize(&e.arg.ManagerAuthArg)
	}
	return AuthorizeApp(&e.arg.ManagerAuthArg, e.arg.App)
}

func (m *ManagerRPC) ListEnvs(arg ManagerListEnvsArg, reply *ManagerListEnvsReply) error {
	return NewTask("ListEnvs", &ListEnvsExecutor{arg, reply}).Run()
}

type ListShasExecutor struct {
	arg   ManagerListShasArg
	reply *ManagerListShasReply
}

func (e *ListShasExecutor) Request() interface{} {
	return e.arg
}

func (e *ListShasExecutor) Result() interface{} {
	return e.reply
}

func (e *ListShasExecutor) Description() string {
	return "[" + e.arg.ManagerAuthArg.User + "] " + e.arg.App
}

func (e *ListShasExecutor) Execute(t *Task) error {
	var err error
	if e.arg.App == "" {
		return errors.New("App is empty")
	}
	e.reply.Shas, err = datamodel.ListShas(e.arg.App)
	if err != nil {
		e.reply.Status = StatusError
	} else {
		sort.Strings(e.reply.Shas)
		e.reply.Status = StatusOk
	}
	return err
}
func (e *ListShasExecutor) Authorize() error {
	return AuthorizeApp(&e.arg.ManagerAuthArg, e.arg.App)
}

func (m *ManagerRPC) ListShas(arg ManagerListShasArg, reply *ManagerListShasReply) error {
	return NewTask("ListShas", &ListShasExecutor{arg, reply}).Run()
}

type ListAppsExecutor struct {
	arg   ManagerListAppsArg
	reply *ManagerListAppsReply
}

func (e *ListAppsExecutor) Request() interface{} {
	return e.arg
}

func (e *ListAppsExecutor) Result() interface{} {
	return e.reply
}

func (e *ListAppsExecutor) Description() string {
	return "[" + e.arg.ManagerAuthArg.User + "] ListApps"
}

func (e *ListAppsExecutor) Execute(t *Task) error {
	var err error
	apps, err := datamodel.ListApps()
	err = AuthorizeSuperUser(&e.arg.ManagerAuthArg)
	if err == nil {
		e.reply.Apps = apps
	} else {
		allowedApps := GetAllowedApps(&e.arg.ManagerAuthArg, e.arg.ManagerAuthArg.User)
		appsCount := len(allowedApps)
		totalAppsCount := len(apps)
		e.reply.Apps = make([]string, 0, appsCount)
		for i := 0; i < totalAppsCount; i++ {
			if allowedApps[apps[i]] {
				e.reply.Apps = append(e.reply.Apps, apps[i])
			}
		}
		err = nil
	}
	if err != nil {
		e.reply.Status = StatusError
	} else {
		sort.Strings(e.reply.Apps)
		e.reply.Status = StatusOk
	}
	return err
}

func (e *ListAppsExecutor) Authorize() error {
	return SimpleAuthorize(&e.arg.ManagerAuthArg)
}

func (m *ManagerRPC) ListApps(arg ManagerListAppsArg, reply *ManagerListAppsReply) error {
	return NewTask("ListApps", &ListAppsExecutor{arg, reply}).Run()
}
