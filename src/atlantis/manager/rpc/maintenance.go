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
	"errors"
	"fmt"
)

// Container Maintenance Management
// NOTE[jigish]: this is a simple pass-through to supervisor with auth

type ContainerMaintenanceExecutor struct {
	arg   ManagerContainerMaintenanceArg
	reply *ManagerContainerMaintenanceReply
}

func (e *ContainerMaintenanceExecutor) Request() interface{} {
	return e.arg
}

func (e *ContainerMaintenanceExecutor) Result() interface{} {
	return e.reply
}

func (e *ContainerMaintenanceExecutor) Description() string {
	return fmt.Sprintf("["+e.arg.ManagerAuthArg.User+"] %s : %t", e.arg.ContainerID, e.arg.Maintenance)
}

func (e *ContainerMaintenanceExecutor) Execute(t *Task) error {
	if e.arg.ContainerID == "" {
		return errors.New("Please specify a container id.")
	}
	instance, err := datamodel.GetInstance(e.arg.ContainerID)
	if err != nil {
		return err
	}
	ihReply, err := supervisor.ContainerMaintenance(instance.Host, e.arg.ContainerID, e.arg.Maintenance)
	e.reply.Status = ihReply.Status
	return err
}

func (e *ContainerMaintenanceExecutor) Authorize() error {
	instance, err := datamodel.GetInstance(e.arg.ContainerID)
	if err != nil {
		return err
	}
	return AuthorizeApp(&e.arg.ManagerAuthArg, instance.App)
}

func (m *ManagerRPC) ContainerMaintenance(arg ManagerContainerMaintenanceArg,
	reply *ManagerContainerMaintenanceReply) error {
	return NewTask("ContainerMaintenance", &ContainerMaintenanceExecutor{arg, reply}).Run()
}

// Manager Idle Check

type IdleExecutor struct {
	arg   ManagerIdleArg
	reply *ManagerIdleReply
}

func (e *IdleExecutor) Request() interface{} {
	return e.arg
}

func (e *IdleExecutor) Result() interface{} {
	return e.reply
}

func (e *IdleExecutor) Description() string {
	return "Idle?"
}

func (e *IdleExecutor) Execute(t *Task) error {
	e.reply.Idle = Tracker.Idle(t)
	e.reply.Status = StatusOk
	return nil
}

func (e *IdleExecutor) Authorize() error {
	return nil // let anybody ask if we're idle. i dont care.
}

func (e *IdleExecutor) AllowDuringMaintenance() bool {
	return true // allow running thus during maintenance
}

func (m *ManagerRPC) Idle(arg ManagerIdleArg, reply *ManagerIdleReply) error {
	return NewTask("Idle", &IdleExecutor{arg, reply}).Run()
}
