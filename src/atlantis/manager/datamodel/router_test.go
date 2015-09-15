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

package datamodel

import (
	. "atlantis/manager/constant"
	"atlantis/manager/helper"
	"atlantis/router/config"
	routerzk "atlantis/router/zk"
	. "github.com/adjust/gocheck"
	"sort"
)

func (s *DatamodelSuite) TestRouterPorts(c *C) {
	// test internal
	Zk.RecursiveDelete(helper.GetBaseRouterPortsPath(true))
	Zk.RecursiveDelete(helper.GetBaseRouterPortsPath(false))
	Zk.RecursiveDelete(helper.GetBaseLockPath())
	CreateRouterPortsPaths()
	CreateLockPaths()

	MinRouterPort = uint16(65533)
	MaxRouterPort = uint16(65535)

	c.Assert(HasRouterPortForAppEnv(true, "app", "env"), Equals, false)
	c.Assert(HasRouterPortForAppEnv(false, "app", "env"), Equals, false)

	appPort, err := reserveRouterPort(true, "app", "env")
	c.Assert(err, IsNil)
	appPort2, err := reserveRouterPort(true, "app", "env")
	c.Assert(err, IsNil)
	c.Assert(appPort, Equals, appPort2)
	c.Assert(HasRouterPortForAppEnv(true, "app", "env"), Equals, true)
	c.Assert(HasRouterPortForAppEnv(false, "app", "env"), Equals, false)

	app2Port, err := reserveRouterPort(true, "app2", "env")
	c.Assert(err, IsNil)
	c.Assert(appPort, Not(Equals), app2Port)
	c.Assert(HasRouterPortForAppEnv(true, "app2", "env"), Equals, true)
	c.Assert(HasRouterPortForAppEnv(false, "app2", "env"), Equals, false)

	_, err = reserveRouterPort(true, "app3", "env2")
	c.Assert(err, IsNil)
	c.Assert(HasRouterPortForAppEnv(true, "app3", "env2"), Equals, true)
	c.Assert(HasRouterPortForAppEnv(false, "app3", "env2"), Equals, false)

	_, err = reserveRouterPort(true, "app4", "env3")
	c.Assert(err, Not(IsNil))

	err = ReclaimRouterPortsForEnv(true, "env")
	c.Assert(err, IsNil)
	c.Assert(HasRouterPortForAppEnv(true, "app", "env"), Equals, false)
	c.Assert(HasRouterPortForAppEnv(true, "app2", "env"), Equals, false)
	c.Assert(HasRouterPortForAppEnv(true, "app3", "env2"), Equals, true)
	err = ReclaimRouterPortsForApp(true, "app3")
	c.Assert(err, IsNil)
	c.Assert(HasRouterPortForAppEnv(true, "app", "env"), Equals, false)
	c.Assert(HasRouterPortForAppEnv(true, "app2", "env"), Equals, false)
	c.Assert(HasRouterPortForAppEnv(true, "app3", "env2"), Equals, false)
	_, err = reserveRouterPort(true, "app4", "env2")
	c.Assert(err, IsNil)
	c.Assert(HasRouterPortForAppEnv(true, "app4", "env2"), Equals, true)
	c.Assert(HasRouterPortForAppEnv(false, "app4", "env2"), Equals, false)

	// test external
	Zk.RecursiveDelete(helper.GetBaseRouterPortsPath(true))
	Zk.RecursiveDelete(helper.GetBaseRouterPortsPath(false))
	Zk.RecursiveDelete(helper.GetBaseLockPath())
	CreateRouterPortsPaths()
	CreateLockPaths()

	MinRouterPort = uint16(65533)
	MaxRouterPort = uint16(65535)

	c.Assert(HasRouterPortForAppEnv(true, "app", "env"), Equals, false)
	c.Assert(HasRouterPortForAppEnv(false, "app", "env"), Equals, false)

	appPort, err = reserveRouterPort(false, "app", "env")
	c.Assert(err, IsNil)
	appPort2, err = reserveRouterPort(false, "app", "env")
	c.Assert(err, IsNil)
	c.Assert(appPort, Equals, appPort2)
	c.Assert(HasRouterPortForAppEnv(false, "app", "env"), Equals, true)
	c.Assert(HasRouterPortForAppEnv(true, "app", "env"), Equals, false)

	app2Port, err = reserveRouterPort(false, "app2", "env")
	c.Assert(err, IsNil)
	c.Assert(appPort, Not(Equals), app2Port)
	c.Assert(HasRouterPortForAppEnv(false, "app2", "env"), Equals, true)
	c.Assert(HasRouterPortForAppEnv(true, "app2", "env"), Equals, false)

	_, err = reserveRouterPort(false, "app3", "env2")
	c.Assert(err, IsNil)
	c.Assert(HasRouterPortForAppEnv(false, "app3", "env2"), Equals, true)
	c.Assert(HasRouterPortForAppEnv(true, "app3", "env2"), Equals, false)

	_, err = reserveRouterPort(false, "app4", "env3")
	c.Assert(err, Not(IsNil))

	err = ReclaimRouterPortsForEnv(false, "env")
	c.Assert(err, IsNil)
	c.Assert(HasRouterPortForAppEnv(false, "app", "env"), Equals, false)
	c.Assert(HasRouterPortForAppEnv(false, "app2", "env"), Equals, false)
	c.Assert(HasRouterPortForAppEnv(false, "app3", "env2"), Equals, true)
	err = ReclaimRouterPortsForApp(false, "app3")
	c.Assert(err, IsNil)
	c.Assert(HasRouterPortForAppEnv(false, "app", "env"), Equals, false)
	c.Assert(HasRouterPortForAppEnv(false, "app2", "env"), Equals, false)
	c.Assert(HasRouterPortForAppEnv(false, "app3", "env2"), Equals, false)
	_, err = reserveRouterPort(false, "app4", "env2")
	c.Assert(err, IsNil)
	c.Assert(HasRouterPortForAppEnv(false, "app4", "env2"), Equals, true)
	c.Assert(HasRouterPortForAppEnv(true, "app4", "env2"), Equals, false)
}

func (s *DatamodelSuite) TestReserveRouterPortAndUpdateTrie(c *C) {
	Zk.RecursiveDelete(helper.GetBaseRouterPortsPath(true))
	Zk.RecursiveDelete(helper.GetBaseRouterPortsPath(false))
	Zk.RecursiveDelete(helper.GetBaseLockPath())
	Zk.RecursiveDelete("/atlantis/router")
	CreateRouterPaths()
	CreateRouterPortsPaths()
	CreateLockPaths()

	MinRouterPort = uint16(65533)
	MaxRouterPort = uint16(65535)

	helper.SetRouterRoot(true)
	port, created, err := ReserveRouterPortAndUpdateTrie(true, "app", "sha", "env")
	c.Assert(err, IsNil)
	c.Assert(created, Equals, true)
	c.Assert(port, Equals, "65533")
	trie, err := routerzk.GetTrie(Zk.Conn, helper.GetAppEnvTrieName("app", "env"))
	c.Assert(err, IsNil)
	c.Assert(len(trie.Rules), Equals, 1)
	port, created, err = ReserveRouterPortAndUpdateTrie(true, "app", "sha2", "env")
	c.Assert(err, IsNil)
	c.Assert(created, Equals, false)
	c.Assert(port, Equals, "65533")
	trie, err = routerzk.GetTrie(Zk.Conn, helper.GetAppEnvTrieName("app", "env"))
	c.Assert(err, IsNil)
	c.Assert(len(trie.Rules), Equals, 2)
}

func (s *DatamodelSuite) TestRouterModel(c *C) {
	Zk.RecursiveDelete(helper.GetBaseRouterPath(true))
	Zk.RecursiveDelete(helper.GetBaseRouterPath(false))
	CreateRouterPaths()
	routers, err := ListRouters(true)
	c.Assert(err, IsNil)
	for _, routersInZone := range routers {
		c.Assert(len(routersInZone), Equals, 0)
	}
	zkRouter := Router(true, Region, "host", "2.2.2.2")
	err = zkRouter.Save()
	c.Assert(err, IsNil)
	fetchedRouter, err := GetRouter(true, Region, "host")
	c.Assert(err, IsNil)
	c.Assert(zkRouter, DeepEquals, fetchedRouter)
	zkRouter.CName = "mycname"
	zkRouter.RecordIDs = []string{"rid1", "rid2"}
	zkRouter.Save()
	fetchedRouter, err = GetRouter(true, Region, "host")
	c.Assert(err, IsNil)
	c.Assert(zkRouter, DeepEquals, fetchedRouter)
	routers, err = ListRouters(true)
	c.Assert(len(routers[Region]), Equals, 1)
	err = zkRouter.Delete()
	c.Assert(err, IsNil)
	routers, err = ListRouters(true)
	c.Assert(len(routers[Region]), Equals, 0)
}

func (s *DatamodelSuite) TestRouterPoolNaming(c *C) {
	c.Assert(helper.CreatePoolName(app, sha, env), Matches, app+"."+sha+"."+env)
}

func (s *DatamodelSuite) TestRouterInternalPool(c *C) {
	Zk.RecursiveDelete("/atlantis/router")
	Zk.RecursiveDelete("/atlantis/apps")
	Zk.RecursiveDelete(helper.GetBaseInstancePath())
	CreateRouterPaths()
	CreateAppPath()
	// fake register app
	CreateOrUpdateApp(false, true, app, "ssh://git@omg.com/app", "/", "omg@omg.com")
	CreateOrUpdateApp(false, true, "app2", "ssh://git@omg.com/app", "/", "omg@omg.com")
	// do tests
	instance, err := CreateInstance(app, sha, env, host+"-1")
	c.Assert(err, IsNil)
	instance.SetPort(uint16(1337))
	instance2, err := CreateInstance(app, sha, env, host+"-2")
	c.Assert(err, IsNil)
	instance2.SetPort(uint16(1338))
	c.Assert(AddToPool([]string{instance.ID, instance2.ID}), IsNil)
	theName := helper.CreatePoolName(app, sha, env)
	helper.SetRouterRoot(true)
	thePool, err := routerzk.GetPool(Zk.Conn, theName)
	c.Assert(err, IsNil)
	c.Assert(thePool.Name, Equals, theName)
	c.Assert(thePool.Config.HealthzEvery, Not(Equals), "")
	c.Assert(thePool.Config.HealthzTimeout, Not(Equals), "")
	c.Assert(thePool.Config.RequestTimeout, Not(Equals), "")
	c.Assert(thePool.Hosts, DeepEquals, map[string]config.Host{host + "-1:1337": config.Host{Address: host + "-1:1337"}, host + "-2:1338": config.Host{Address: host + "-2:1338"}})
	newInstance, err := CreateInstance("app2", "sha1", "env1", host+"-1")
	c.Assert(err, IsNil)
	newInstance.SetPort(uint16(1339))
	newInstance2, err := CreateInstance(app, sha, env, host+"-3")
	c.Assert(err, IsNil)
	newInstance2.SetPort(uint16(1340))
	c.Assert(DeleteFromPool([]string{instance2.ID}), IsNil)
	instance2.Delete()
	c.Assert(AddToPool([]string{newInstance.ID, newInstance2.ID}), IsNil)
	helper.SetRouterRoot(true)
	thePool, err = routerzk.GetPool(Zk.Conn, theName)
	c.Assert(err, IsNil)
	c.Assert(thePool.Name, Equals, theName)
	c.Assert(thePool.Config.HealthzEvery, Not(Equals), "")
	c.Assert(thePool.Config.HealthzTimeout, Not(Equals), "")
	c.Assert(thePool.Config.RequestTimeout, Not(Equals), "")
	c.Assert(thePool.Hosts, DeepEquals, map[string]config.Host{host + "-1:1337": config.Host{Address: host + "-1:1337"}, host + "-3:1340": config.Host{Address: host + "-3:1340"}})
	helper.SetRouterRoot(true)
	thePool2, err := routerzk.GetPool(Zk.Conn, helper.CreatePoolName("app2", "sha1", "env1"))
	c.Assert(err, IsNil)
	c.Assert(thePool2.Name, Equals, helper.CreatePoolName("app2", "sha1", "env1"))
	c.Assert(thePool2.Config.HealthzEvery, Not(Equals), "")
	c.Assert(thePool2.Config.HealthzTimeout, Not(Equals), "")
	c.Assert(thePool2.Config.RequestTimeout, Not(Equals), "")
	c.Assert(thePool2.Hosts, DeepEquals, map[string]config.Host{host + "-1:1339": config.Host{Address: host + "-1:1339"}})
	helper.SetRouterRoot(true)
	pools, err := routerzk.ListPools(Zk.Conn)
	c.Assert(err, IsNil)
	sort.Strings(pools)
	c.Assert(pools, DeepEquals, []string{thePool2.Name, thePool.Name})
	c.Assert(DeleteFromPool([]string{instance.ID, newInstance.ID, newInstance2.ID}), IsNil)
	instance.Delete()
	newInstance.Delete()
	newInstance2.Delete()
	helper.SetRouterRoot(true)
	thePool, err = routerzk.GetPool(Zk.Conn, theName)
	c.Assert(err, Not(IsNil))
}

func (s *DatamodelSuite) TestRouterExternalPool(c *C) {
	Zk.RecursiveDelete("/atlantis/router")
	Zk.RecursiveDelete("/atlantis/apps")
	Zk.RecursiveDelete(helper.GetBaseInstancePath())
	CreateRouterPaths()
	CreateAppPath()
	// fake register app
	CreateOrUpdateApp(false, false, app, "ssh://git@omg.com/app", "/", "omg@omg.com")
	CreateOrUpdateApp(false, false, "app2", "ssh://git@omg.com/app", "/", "omg@omg.com")
	// do tests
	instance, err := CreateInstance(app, sha, env, host+"-1")
	c.Assert(err, IsNil)
	instance.SetPort(uint16(1337))
	instance2, err := CreateInstance(app, sha, env, host+"-2")
	c.Assert(err, IsNil)
	instance2.SetPort(uint16(1338))
	c.Assert(AddToPool([]string{instance.ID, instance2.ID}), IsNil)
	theName := helper.CreatePoolName(app, sha, env)
	helper.SetRouterRoot(false)
	thePool, err := routerzk.GetPool(Zk.Conn, theName)
	c.Assert(err, IsNil)
	c.Assert(thePool.Name, Equals, theName)
	c.Assert(thePool.Config.HealthzEvery, Not(Equals), "")
	c.Assert(thePool.Config.HealthzTimeout, Not(Equals), "")
	c.Assert(thePool.Config.RequestTimeout, Not(Equals), "")
	c.Assert(thePool.Hosts, DeepEquals, map[string]config.Host{host + "-1:1337": config.Host{Address: host + "-1:1337"}, host + "-2:1338": config.Host{Address: host + "-2:1338"}})
	newInstance, err := CreateInstance("app2", "sha1", "env1", host+"-1")
	c.Assert(err, IsNil)
	newInstance.SetPort(uint16(1339))
	newInstance2, err := CreateInstance(app, sha, env, host+"-3")
	c.Assert(err, IsNil)
	newInstance2.SetPort(uint16(1340))
	c.Assert(DeleteFromPool([]string{instance2.ID}), IsNil)
	instance2.Delete()
	c.Assert(AddToPool([]string{newInstance.ID, newInstance2.ID}), IsNil)
	helper.SetRouterRoot(false)
	thePool, err = routerzk.GetPool(Zk.Conn, theName)
	c.Assert(err, IsNil)
	c.Assert(thePool.Name, Equals, theName)
	c.Assert(thePool.Config.HealthzEvery, Not(Equals), "")
	c.Assert(thePool.Config.HealthzTimeout, Not(Equals), "")
	c.Assert(thePool.Config.RequestTimeout, Not(Equals), "")
	c.Assert(thePool.Hosts, DeepEquals, map[string]config.Host{host + "-1:1337": config.Host{Address: host + "-1:1337"}, host + "-3:1340": config.Host{Address: host + "-3:1340"}})
	helper.SetRouterRoot(false)
	thePool2, err := routerzk.GetPool(Zk.Conn, helper.CreatePoolName("app2", "sha1", "env1"))
	c.Assert(err, IsNil)
	c.Assert(thePool2.Name, Equals, helper.CreatePoolName("app2", "sha1", "env1"))
	c.Assert(thePool2.Config.HealthzEvery, Not(Equals), "")
	c.Assert(thePool2.Config.HealthzTimeout, Not(Equals), "")
	c.Assert(thePool2.Config.RequestTimeout, Not(Equals), "")
	c.Assert(thePool2.Hosts, DeepEquals, map[string]config.Host{host + "-1:1339": config.Host{Address: host + "-1:1339"}})
	helper.SetRouterRoot(false)
	pools, err := routerzk.ListPools(Zk.Conn)
	c.Assert(err, IsNil)
	sort.Strings(pools)
	c.Assert(pools, DeepEquals, []string{thePool2.Name, thePool.Name})
	c.Assert(DeleteFromPool([]string{instance.ID, newInstance.ID, newInstance2.ID}), IsNil)
	instance.Delete()
	newInstance.Delete()
	newInstance2.Delete()
	helper.SetRouterRoot(false)
	thePool, err = routerzk.GetPool(Zk.Conn, theName)
	c.Assert(err, Not(IsNil))
}
