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
	"atlantis/manager/crypto"
	crypt "atlantis/crypto"
	"atlantis/manager/helper"
	"atlantis/manager/rpc/types"
	"encoding/json"
	"errors"
	"fmt"
	"log"
)

type App struct {
	Name string		`db:"name"`
	NonAtlantis bool	`db:"nonatlantis"`
	Internal bool		`db:"internal"`
	Email string		`db:"email"`
	Repo string		`db:"repo"`
	Root string		`db:"root"`
}

type EnvDepData struct {
	
	ID int64	`db:"id"`
	Enviroment string `db:"env"`
	SecGroup string	`db:"secgroup"`
	DataMap	string	`db:"datamap"`
	EncryptedData string	`db:"encryptedData"`
	App string `db:"app"`
	Instance string `db:"instance"`
}

type AppDepData struct {
	ID int64 `db:"id"`
	DepApp string	`db:"depApp"`
	EnvDepDataId int64 `db:"envdepdataId"`
	App string `db:"app"`    
}

type ZkApp types.App

func GetApp(name string) (za *ZkApp, err error) {
	za = &ZkApp{}
	err = getJson(helper.GetBaseAppPath(name), za)
	if za.DependerEnvData == nil {
		za.DependerEnvData = map[string]*types.DependerEnvData{}
		za.Save()
	}
	if za.DependerAppData == nil {
		za.DependerAppData = map[string]*types.DependerAppData{}
		za.Save()
	}
	
	////////////////////////// SQL ////////////////////////////////
	obj, err := DbMap.Get(App{}, name)
	if err != nil {
		fmt.Printf("\n%v\n", err)	
	} else {
		if obj == nil {
			fmt.Println("App doesn't exists \n")
			err = errors.New("No app with name: " + name)
			za = nil	
		} else { 
			app := obj.(*App)
			err = nil
			za = &ZkApp{ app.NonAtlantis, app.Internal, app.Name, app.Email, app.Repo, app.Root, 
				map[string]*types.DependerEnvData{}, map[string]*types.DependerAppData{}}
		}
	}
	///////////////////////////////////////////////////////////////

	return
}

func CreateOrUpdateApp(nonAtlantis, internal bool, name, repo, root, email string) (*ZkApp, error) {
	za, err := GetApp(name)
	if err != nil {
		za = &ZkApp{
			NonAtlantis:     nonAtlantis,
			Internal:        internal,
			Name:            name,
			Repo:            repo,
			Root:            root,
			Email:           email,
			DependerEnvData: map[string]*types.DependerEnvData{},
			DependerAppData: map[string]*types.DependerAppData{},
		}
		if err := za.Save(); err != nil {
			return za, err
		}
	} else {
		za.Name = name
		za.Repo = repo
		za.Root = root
		za.Email = email
		if za.Internal != internal {
			return za, errors.New("apps may not change from internal to external (and visa versa). please unregister and reregister.")
		}
		if za.NonAtlantis != nonAtlantis {
			return za, errors.New("apps may not change from non-atlantis to atlantis (and visa versa). please unregister and reregister.")
		}
		if err := za.Save(); err != nil {
			return za, err
		}
	}
	return za, nil
}

func (za *ZkApp) Delete() error {
	// this just deletes the registration. no need to clean up already deployed instances
	if !za.NonAtlantis {
		if err := ReclaimRouterPortsForApp(za.Internal, za.Name); err != nil {
			return err
		}
	}
	
	////////////////////// SQL /////////////////////////////////
	app := App{}
	app.Name = za.Name
	DbMap.Delete(app)
	//if not
	//_, err := DbMap.Exec("delete from apps where name=?", app.Name)
	//////////////////////////////////////////////////////////
	
	
	return Zk.RecursiveDelete(za.path())
}

func (za *ZkApp) path() string {
	return helper.GetBaseAppPath(za.Name)
}

func (za *ZkApp) Save() error {
	if err := setJson(za.path(), za); err != nil {
		return err
	}


	///////////////////////// SQL ///////////////////////////////////////
	fmt.Println("YAY SQL SAVE APP STUFF")
	app := App{za.Name, za.NonAtlantis, za.Internal, za.Email, za.Repo, za.Root}
	obj, err := DbMap.Get(App{}, za.Name)	
	if err != nil {
		fmt.Printf("\n Failed trying to check if app exists \n")
	}
	if za.Name == "" {
		return errors.New("Trying to save app with no name")
	}
	//app doesnt exist, insert
	if obj == nil {

		fmt.Printf("\n App %s does not exists, insert \n", za.Name)
		err = DbMap.Insert(&app)
		if err != nil {
			fmt.Printf("\n %v \n", err)
		}
	} else {
		fmt.Printf("\n App %s DOES exist, update \n ", za.Name)
		//update
		_, err = DbMap.Update(&app)
		if err != nil {
			fmt.Printf("\n %v \n", err)
		}
	}	
	/////////////////////////////////////////////////////////////////////

	return nil
}

func (za *ZkApp) AddDependerEnvData(data *types.DependerEnvData) error {
	if za.DependerEnvData == nil {
		za.DependerEnvData = map[string]*types.DependerEnvData{}
	}
	if _, err := GetEnv(data.Name); err != nil {
		return err
	}
	crypto.EncryptDependerEnvData(data)
	za.DependerEnvData[data.Name] = data

	//////////////////////////// SQL ///////////////////////////////////
	secGroupStr, err := json.Marshal(data.SecurityGroup)
	if err != nil {

	}
	mapData, err := json.Marshal(data.DataMap)
	if err != nil {
	}	
	encDataStr := string(crypt.Encrypt(mapData)) 			
	dep := EnvDepData{Enviroment: data.Name, SecGroup: string(secGroupStr), DataMap: string(mapData), 
				EncryptedData: encDataStr, App: za.Name }	
	err = DbMap.Insert(&dep)
	if err != nil {
	
	}	
	///////////////////////////////////////////////////////////////////

	return za.Save()
}

func (za *ZkApp) RemoveDependerEnvData(env string) error {
	if za.DependerEnvData == nil {
		za.DependerEnvData = map[string]*types.DependerEnvData{}
	}
	delete(za.DependerEnvData, env)

	///////////////////////////// SQL ///////////////////////////////
	_, err := DbMap.Exec("delete from deps where env=? and app=?", env, za.Name)
	if err != nil {
	}
	////////////////////////////////////////////////////////////////

	return za.Save()
}

func (za *ZkApp) GetDependerEnvData(env string, decrypt bool) *types.DependerEnvData {
	if za.DependerEnvData == nil {
		za.DependerEnvData = map[string]*types.DependerEnvData{}
		za.Save()
	}
	ded := za.DependerEnvData[env]
	if ded == nil {
		return nil
	}
	if decrypt {
		crypto.DecryptDependerEnvData(ded)
	}

	/////////////////////// SQL /////////////////////////////////
	var dep EnvDepData 
	err := DbMap.SelectOne(&dep, "select * from deps where env=? and app=?", env, za.Name)
	if err != nil {
	}
	var secG map[string][]uint16
	var dMap map[string]interface{}
	err = json.Unmarshal([]byte(dep.SecGroup), secG)
	if err != nil {
	}
	err = json.Unmarshal([]byte(dep.DataMap), dMap)
	if err != nil{
	}
	//ded := DependerEnvData{dep.Enviroment, secG, dep.EncryptedData, dep.DataMap}
	////////////////////////////////////////////////////////////


	return ded
}

func (za *ZkApp) AddDependerAppData(data *types.DependerAppData) error {
	if za.DependerAppData == nil {
		za.DependerAppData = map[string]*types.DependerAppData{}
	}
	if _, err := GetApp(data.Name); err != nil {
		return err
	}
	for _, ded := range data.DependerEnvData {
		if _, err := GetEnv(ded.Name); err != nil {
			return err
		}
		crypto.EncryptDependerEnvData(ded)
	}
	za.DependerAppData[data.Name] = data
	
	////////////////////////////// SQL //////////////////////////////////////////
	secGroupStr, err := json.Marshal(data.DependerEnvData[data.Name].SecurityGroup)
	if err != nil {

	}
	mapData, err := json.Marshal(data.DependerEnvData[data.Name].DataMap)
	if err != nil {
	}	
	encDataStr := string(crypt.Encrypt(mapData)) 			
	dep := EnvDepData{Enviroment: data.DependerEnvData[data.Name].Name, SecGroup: string(secGroupStr), DataMap: string(mapData), 
				EncryptedData: encDataStr}
	//should populate dep with ID
	err = DbMap.Insert(&dep)
	appDep := AppDepData{ DepApp: data.Name, EnvDepDataId: dep.ID, App: za.Name}  	
	err = DbMap.Insert(&appDep)
	if err != nil {
	
	}	
	////////////////////////////////////////////////////////////////////////////



	return za.Save()
}

func (za *ZkApp) RemoveDependerAppData(app string) error {
	if za.DependerAppData == nil {
		za.DependerAppData = map[string]*types.DependerAppData{}
	}
	delete(za.DependerAppData, app)


	///////////////////////////// SQL ///////////////////////////////////////
	//get envdepdata id to delete that after deleting appdepdata
	var envDepDataId int64
	_, err := DbMap.Select(&envDepDataId, "select envdepdataId from appdepdata where depApp=? and app=?", app, za.Name)
	if err != nil {
	}
	_, err = DbMap.Exec("delete from appdepdata where depApp=? and app=?", app, za.Name)
	if err != nil {
	}
	_, err = DbMap.Exec("delete from envdepdata where id=?", envDepDataId)
	if err != nil{
	}
	////////////////////////////////////////////////////////////////////////	

	return za.Save()
}

func (za *ZkApp) GetDependerAppData(app string, decrypt bool) *types.DependerAppData {
	if za.DependerAppData == nil {
		za.DependerAppData = map[string]*types.DependerAppData{}
		za.Save()
	}
	dad := za.DependerAppData[app]
	if dad == nil {
		return nil
	}
	if decrypt {
		for _, ded := range dad.DependerEnvData {
			crypto.DecryptDependerEnvData(ded)
		}
	}
	//////////////////////////// SQL ///////////////////////////////////////
	//TODO: more dependecy shit 
	///////////////////////////////////////////////////////////////////////

	return dad
}

func (za *ZkApp) AddDependerEnvDataForDependerApp(app string, data *types.DependerEnvData) error {
	if _, err := GetApp(app); err != nil {
		return err
	}
	if _, err := GetEnv(data.Name); err != nil {
		return err
	}
	dad := za.GetDependerAppData(app, false)
	if dad == nil {
		dad = &types.DependerAppData{Name: app, DependerEnvData: map[string]*types.DependerEnvData{}}
	}
	crypto.EncryptDependerEnvData(data)
	dad.DependerEnvData[data.Name] = data
	za.DependerAppData[app] = dad
	return za.Save()
}

func (za *ZkApp) RemoveDependerEnvDataForDependerApp(app, env string) error {
	dad := za.GetDependerAppData(app, false)
	if dad == nil {
		return nil
	}
	delete(dad.DependerEnvData, env)
	za.DependerAppData[app] = dad
	return za.Save()
}

func (za *ZkApp) GetDependerEnvDataForDependerApp(app, env string, decrypt bool) *types.DependerEnvData {
	dad := za.GetDependerAppData(app, false)
	if dad == nil {
		return nil
	}
	ded := dad.DependerEnvData[env]
	if ded == nil {
		return nil
	}
	if decrypt {
		crypto.DecryptDependerEnvData(ded)
	}
	return ded
}

func ListRegisteredApps() (apps []string, err error) {
	apps, _, err = Zk.VisibleChildren(helper.GetBaseAppPath())
	if err != nil {
		log.Printf("Error getting list of registered apps. Error: %s.", err.Error())
	}
	if apps == nil {
		log.Println("No registered apps found. Returning empty list.")
		apps = []string{}
	}
	
	/////////////////////////// SQL ///////////////////////////////////
	var appSql []App
	_, err = DbMap.Select(&appSql, "select * from apps")
	if err != nil {

	}
	////////////////////////////////////////////////////////////////////
	
	return
}
