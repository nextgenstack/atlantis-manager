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
	aldap "atlantis/manager/ldap"
	. "atlantis/manager/rpc/types"
	"errors"
	"fmt"
	"github.com/mavricknz/ldap"
	"log"
	"regexp"
	"sort"
	"strings"
)

const maxTeamAdmins int = 2

// ----------------------------------------------------------------------------------------------------------
// Teams
// ----------------------------------------------------------------------------------------------------------

type CreateTeamExecutor struct {
	arg   ManagerTeamArg
	reply *ManagerTeamReply
}

func (e *CreateTeamExecutor) Request() interface{} {
	return e.arg
}

func (e *CreateTeamExecutor) Result() interface{} {
	return e.reply
}

func (e *CreateTeamExecutor) Description() string {
	return fmt.Sprintf("["+e.arg.ManagerAuthArg.User+"] %s : %s", e.arg.User, e.arg.Team)
}

func (e *CreateTeamExecutor) Execute(t *Task) error {
        return errors.New("Team creation is no longer supported")
}

func (e *CreateTeamExecutor) Authorize() error {
	if err := checkRole("permissions", "write"); err != nil {
		return err
	}
	return SimpleAuthorize(&e.arg.ManagerAuthArg)
}

type DeleteTeamExecutor struct {
	arg   ManagerTeamArg
	reply *ManagerTeamReply
}

func (e *DeleteTeamExecutor) Request() interface{} {
	return e.arg
}

func (e *DeleteTeamExecutor) Result() interface{} {
	return e.reply
}

func (e *DeleteTeamExecutor) Description() string {
	return fmt.Sprintf("["+e.arg.ManagerAuthArg.User+"] %s", e.arg.Team)
}

func (e *DeleteTeamExecutor) Execute(t *Task) error {
	return errors.New("Team Delete is no longer supported")
}

func (e *DeleteTeamExecutor) Authorize() error {
	if err := checkRole("permissions", "write"); err != nil {
		return err
	}
	return AuthorizeTeamAdmin(&e.arg.ManagerAuthArg, e.arg.Team)
}

func (m *ManagerRPC) CreateTeam(arg ManagerTeamArg, reply *ManagerTeamReply) error {
	return NewTask("CreateTeam", &CreateTeamExecutor{arg, reply}).Run()
}

func (m *ManagerRPC) DeleteTeam(arg ManagerTeamArg, reply *ManagerTeamReply) error {
	return NewTask("DeleteTeam", &DeleteTeamExecutor{arg, reply}).Run()
}

type AddTeamEmailExecutor struct {
	arg   ManagerEmailArg
	reply *ManagerEmailReply
}

func (e *AddTeamEmailExecutor) Request() interface{} {
	return e.arg
}

func (e *AddTeamEmailExecutor) Result() interface{} {
	return e.reply
}

func (e *AddTeamEmailExecutor) Description() string {
	return fmt.Sprintf("["+e.arg.ManagerAuthArg.User+"] %s : %s", e.arg.Email, e.arg.Team)
}

func (e *AddTeamEmailExecutor) Execute(t *Task) error {
	return ModifyTeamEmail(ldap.ModAdd, e.arg, e.reply)
}

func (e *AddTeamEmailExecutor) Authorize() error {
	if err := checkRole("permissions", "write"); err != nil {
		return err
	}
	return AuthorizeTeamAdmin(&e.arg.ManagerAuthArg, e.arg.Team)
}

type RemoveTeamEmailExecutor struct {
	arg   ManagerEmailArg
	reply *ManagerEmailReply
}

func (e *RemoveTeamEmailExecutor) Request() interface{} {
	return e.arg
}

func (e *RemoveTeamEmailExecutor) Result() interface{} {
	return e.reply
}

func (e *RemoveTeamEmailExecutor) Description() string {
	return fmt.Sprintf("["+e.arg.ManagerAuthArg.User+"] %s : %s", e.arg.Email, e.arg.Team)
}

func (e *RemoveTeamEmailExecutor) Execute(t *Task) error {
	return ModifyTeamEmail(ldap.ModDelete, e.arg, e.reply)
}

func (e *RemoveTeamEmailExecutor) Authorize() error {
	if err := checkRole("permissions", "write"); err != nil {
		return err
	}
	return AuthorizeTeamAdmin(&e.arg.ManagerAuthArg, e.arg.Team)
}

func ModifyTeamEmail(action int, arg ManagerEmailArg, reply *ManagerEmailReply) error {
	return errors.New("Team email modification is no longer supported")
}

func (m *ManagerRPC) AddTeamEmail(arg ManagerEmailArg, reply *ManagerEmailReply) error {
	return NewTask("AddTeamEmail", &AddTeamEmailExecutor{arg, reply}).Run()
}

func (m *ManagerRPC) RemoveTeamEmail(arg ManagerEmailArg, reply *ManagerEmailReply) error {
	return NewTask("RemoveTeamEmail", &RemoveTeamEmailExecutor{arg, reply}).Run()
}

type AddTeamAdminExecutor struct {
	arg   ManagerModifyTeamAdminArg
	reply *ManagerModifyTeamAdminReply
}

func (e *AddTeamAdminExecutor) Request() interface{} {
	return e.arg
}

func (e *AddTeamAdminExecutor) Result() interface{} {
	return e.reply
}

func (e *AddTeamAdminExecutor) Description() string {
	return fmt.Sprintf("["+e.arg.ManagerAuthArg.User+"] %s : %s", e.arg.User, e.arg.Team)
}

func (e *AddTeamAdminExecutor) Execute(t *Task) error {
	return errors.New("Modify Team admin is no longer supported")
}

func (e *AddTeamAdminExecutor) Authorize() error {
	if err := checkRole("permissions", "write"); err != nil {
		return err
	}
	req := ManagerTeamAdminArg{e.arg.ManagerAuthArg, e.arg.Team}
	var res ManagerTeamAdminReply
	err := NewTask("AddTeamAdmin-IsTeamAdmin", &IsTeamAdminExecutor{req, &res}).Run()
	if err != nil || !res.IsAdmin {
		return errors.New("Permission denied")
	}
	reql := ManagerListTeamAdminsArg{e.arg.ManagerAuthArg, e.arg.Team}
	var resl ManagerListTeamAdminsReply
	err = NewTask("AddTeamAdmin-ListTeamAdmins", &ListTeamAdminsExecutor{reql, &resl}).Run()
	if err != nil {
		return err
	}
	if len(resl.TeamAdmins) >= maxTeamAdmins {
		return fmt.Errorf("Maximum of %d admins are allowed per team", maxTeamAdmins)
	}
	return nil
}

type RemoveTeamAdminExecutor struct {
	arg   ManagerModifyTeamAdminArg
	reply *ManagerModifyTeamAdminReply
}

func (e *RemoveTeamAdminExecutor) Request() interface{} {
	return e.arg
}

func (e *RemoveTeamAdminExecutor) Result() interface{} {
	return e.reply
}

func (e *RemoveTeamAdminExecutor) Description() string {
	return fmt.Sprintf("["+e.arg.ManagerAuthArg.User+"] %s : %s", e.arg.User, e.arg.Team)
}

func (e *RemoveTeamAdminExecutor) Execute(t *Task) error {
	return errors.New("Modify Team admin is no longer supported")
}

func (e *RemoveTeamAdminExecutor) Authorize() error {
	if err := checkRole("permissions", "write"); err != nil {
		return err
	}
	req := ManagerTeamAdminArg{e.arg.ManagerAuthArg, e.arg.Team}
	var res ManagerTeamAdminReply
	err := NewTask("RemoveTeamAdmin-IsTeamAdmin", &IsTeamAdminExecutor{req, &res}).Run()
	if err != nil || !res.IsAdmin {
		return errors.New("Permission denied")
	}
	return nil
}

func ModifyTeamAdmin(action int, arg ManagerModifyTeamAdminArg, reply *ManagerModifyTeamAdminReply) error {
	return errors.New("Modify Team admin is no longer supported")
}

func (m *ManagerRPC) AddTeamAdmin(arg ManagerModifyTeamAdminArg, reply *ManagerModifyTeamAdminReply) error {
	return NewTask("AddTeamAdmin", &AddTeamAdminExecutor{arg, reply}).Run()
}

func (m *ManagerRPC) RemoveTeamAdmin(arg ManagerModifyTeamAdminArg, reply *ManagerModifyTeamAdminReply) error {
	return NewTask("RemoveTeamAdmin", &RemoveTeamAdminExecutor{arg, reply}).Run()
}

type AddTeamMemberExecutor struct {
	arg   ManagerTeamMemberArg
	reply *ManagerTeamMemberReply
}

func (e *AddTeamMemberExecutor) Request() interface{} {
	return e.arg
}

func (e *AddTeamMemberExecutor) Result() interface{} {
	return e.reply
}

func (e *AddTeamMemberExecutor) Description() string {
	return fmt.Sprintf("["+e.arg.ManagerAuthArg.User+"] %s : %s", e.arg.User, e.arg.Team)
}

func (e *AddTeamMemberExecutor) Execute(t *Task) error {
	return errors.New("Modify Team memeber is no longer supported")
}

func (e *AddTeamMemberExecutor) Authorize() error {
	if err := checkRole("permissions", "write"); err != nil {
		return err
	}
	return AuthorizeTeamAdmin(&e.arg.ManagerAuthArg, e.arg.Team)
}

type RemoveTeamMemberExecutor struct {
	arg   ManagerTeamMemberArg
	reply *ManagerTeamMemberReply
}

func (e *RemoveTeamMemberExecutor) Request() interface{} {
	return e.arg
}

func (e *RemoveTeamMemberExecutor) Result() interface{} {
	return e.reply
}

func (e *RemoveTeamMemberExecutor) Description() string {
	return fmt.Sprintf("["+e.arg.ManagerAuthArg.User+"] %s : %s", e.arg.User, e.arg.Team)
}

func (e *RemoveTeamMemberExecutor) Execute(t *Task) error {
	return errors.New("Modify Team memeber is no longer supporte")
}

func (e *RemoveTeamMemberExecutor) Authorize() error {
	if err := checkRole("permissions", "write"); err != nil {
		return err
	}
	return AuthorizeTeamAdmin(&e.arg.ManagerAuthArg, e.arg.Team)
}

func ModifyTeamMember(action int, arg ManagerTeamMemberArg, reply *ManagerTeamMemberReply) error {
	return nil
}

func (m *ManagerRPC) AddTeamMember(arg ManagerTeamMemberArg, reply *ManagerTeamMemberReply) error {
	return NewTask("AddTeamMember", &AddTeamMemberExecutor{arg, reply}).Run()
}

func (m *ManagerRPC) RemoveTeamMember(arg ManagerTeamMemberArg, reply *ManagerTeamMemberReply) error {
	return NewTask("RemoveTeamMember", &RemoveTeamMemberExecutor{arg, reply}).Run()
}

// ----------------------------------------------------------------------------------------------------------
// Listers
// ----------------------------------------------------------------------------------------------------------

type ListTeamsExecutor struct {
	arg   ManagerListTeamsArg
	reply *ManagerListTeamsReply
}

func (e *ListTeamsExecutor) Request() interface{} {
	return e.arg
}

func (e *ListTeamsExecutor) Result() interface{} {
	return e.reply
}

func (e *ListTeamsExecutor) Description() string {
	return "ListTeams"
}

func (e *ListTeamsExecutor) Execute(t *Task) (err error) {
	e.reply.Teams, err = ListTeams(&e.arg.ManagerAuthArg)
	if err == nil {
		sort.Strings(e.reply.Teams)
	}
	return err
}

func (e *ListTeamsExecutor) Authorize() error {
	return SimpleAuthorize(&e.arg.ManagerAuthArg)
}

func (m *ManagerRPC) ListTeams(arg ManagerListTeamsArg, reply *ManagerListTeamsReply) error {
	return NewTask("ListTeams", &ListTeamsExecutor{arg, reply}).Run()
}

type ListTeamEmailsExecutor struct {
	arg   ManagerListTeamEmailsArg
	reply *ManagerListTeamEmailsReply
}

func (e *ListTeamEmailsExecutor) Request() interface{} {
	return e.arg
}

func (e *ListTeamEmailsExecutor) Result() interface{} {
	return e.reply
}

func (e *ListTeamEmailsExecutor) Description() string {
	return "ListTeamEmails"
}

func (e *ListTeamEmailsExecutor) Execute(t *Task) (err error) {
	e.reply.TeamEmails, err = ListTeamEmails(e.arg.Team, &e.arg.ManagerAuthArg)
	if err == nil {
		sort.Strings(e.reply.TeamEmails)
	}
	return err
}

func (e *ListTeamEmailsExecutor) Authorize() error {
	return SimpleAuthorize(&e.arg.ManagerAuthArg)
}

func (m *ManagerRPC) ListTeamEmails(arg ManagerListTeamEmailsArg, reply *ManagerListTeamEmailsReply) error {
	return NewTask("ListTeamEmails", &ListTeamEmailsExecutor{arg, reply}).Run()
}

type ListTeamAdminsExecutor struct {
	arg   ManagerListTeamAdminsArg
	reply *ManagerListTeamAdminsReply
}

func (e *ListTeamAdminsExecutor) Request() interface{} {
	return e.arg
}

func (e *ListTeamAdminsExecutor) Result() interface{} {
	return e.reply
}

func (e *ListTeamAdminsExecutor) Description() string {
	return "ListTeamAdmins"
}

func (e *ListTeamAdminsExecutor) Execute(t *Task) (err error) {
	e.reply.TeamAdmins = []string{}
	return nil
}

func (e *ListTeamAdminsExecutor) Authorize() error {
	return SimpleAuthorize(&e.arg.ManagerAuthArg)
}

func (m *ManagerRPC) ListTeamAdmins(arg ManagerListTeamAdminsArg, reply *ManagerListTeamAdminsReply) error {
	return NewTask("ListTeamAdmins", &ListTeamAdminsExecutor{arg, reply}).Run()
}

type ListTeamMembersExecutor struct {
	arg   ManagerListTeamMembersArg
	reply *ManagerListTeamMembersReply
}

func (e *ListTeamMembersExecutor) Request() interface{} {
	return e.arg
}

func (e *ListTeamMembersExecutor) Result() interface{} {
	return e.reply
}

func (e *ListTeamMembersExecutor) Description() string {
	return "ListTeamMembers"
}

func (e *ListTeamMembersExecutor) Execute(t *Task) (err error) {
	e.reply.TeamMembers, err = ListTeamMembers(e.arg.Team, &e.arg.ManagerAuthArg)
	if err == nil {
		sort.Strings(e.reply.TeamMembers)
	}
	return err
}

func (e *ListTeamMembersExecutor) Authorize() error {
	return SimpleAuthorize(&e.arg.ManagerAuthArg)
}

func (m *ManagerRPC) ListTeamMembers(arg ManagerListTeamMembersArg, reply *ManagerListTeamMembersReply) error {
	return NewTask("ListTeamMembers", &ListTeamMembersExecutor{arg, reply}).Run()
}

type ListTeamAppsExecutor struct {
	arg   ManagerListTeamAppsArg
	reply *ManagerListTeamAppsReply
}

func (e *ListTeamAppsExecutor) Request() interface{} {
	return e.arg
}

func (e *ListTeamAppsExecutor) Result() interface{} {
	return e.reply
}

func (e *ListTeamAppsExecutor) Description() string {
	return "ListTeamApps"
}

func (e *ListTeamAppsExecutor) Execute(t *Task) (err error) {
	e.reply.TeamApps, err = ListTeamApps(e.arg.Team)
	if err == nil {
		sort.Strings(e.reply.TeamApps)
	}
	return err
}

func (e *ListTeamAppsExecutor) Authorize() error {
	return SimpleAuthorize(&e.arg.ManagerAuthArg)
}

func (m *ManagerRPC) ListTeamApps(arg ManagerListTeamAppsArg, reply *ManagerListTeamAppsReply) error {
	return NewTask("ListTeamApps", &ListTeamAppsExecutor{arg, reply}).Run()
}

// ----------------------------------------------------------------------------------------------------------
// Apps
// ----------------------------------------------------------------------------------------------------------

type AllowAppExecutor struct {
	arg   ManagerAppArg
	reply *ManagerAppReply
}

func (e *AllowAppExecutor) Request() interface{} {
	return e.arg
}

func (e *AllowAppExecutor) Result() interface{} {
	return e.reply
}

func (e *AllowAppExecutor) Description() string {
	return fmt.Sprintf("["+e.arg.ManagerAuthArg.User+"] %s : %s", e.arg.App, e.arg.Team)
}

func (e *AllowAppExecutor) Execute(t *Task) error {
	var suReply ManagerSuperUserReply
	if err := NewTask("IsAppAllowed-IsSuperUser",
		&IsSuperUserExecutor{ManagerSuperUserArg{e.arg.ManagerAuthArg}, &suReply}).Run(); err != nil {
			return err
	}

	if !TeamExists(e.arg.Team, &e.arg.ManagerAuthArg) && !suReply.IsSuperUser {
		return errors.New("You do not have permission to allow apps for team " + e.arg.Team)
	}
	teamApps := datamodel.GetTeamapps(e.arg.Team)
	
	//
	//TODO: check app exist before add it to team
	//

        return teamApps.AddApp(e.arg.App)
}

func (e *AllowAppExecutor) Authorize() error {
	//TODO need to check user in the team first
	return nil
}

type DisallowAppExecutor struct {
	arg   ManagerAppArg
	reply *ManagerAppReply
}

func (e *DisallowAppExecutor) Request() interface{} {
	return e.arg
}

func (e *DisallowAppExecutor) Result() interface{} {
	return e.reply
}

func (e *DisallowAppExecutor) Description() string {
	return fmt.Sprintf("["+e.arg.ManagerAuthArg.User+"] %s : %s", e.arg.App, e.arg.Team)
}

func (e *DisallowAppExecutor) Execute(t *Task) error {
	var suReply ManagerSuperUserReply
	if err := NewTask("IsAppAllowed-IsSuperUser",
		&IsSuperUserExecutor{ManagerSuperUserArg{e.arg.ManagerAuthArg}, &suReply}).Run(); err != nil {
			return err
	}

	if !TeamExists(e.arg.Team, &e.arg.ManagerAuthArg) && !suReply.IsSuperUser {
		return errors.New("You do not have permission to disallow apps for team " + e.arg.Team)
	}

	teamApps := datamodel.GetTeamapps(e.arg.Team)
	return teamApps.DeleteApp(e.arg.App)
}

func (e *DisallowAppExecutor) Authorize() error {
	//return AuthorizeTeamAdmin(&e.arg.ManagerAuthArg, e.arg.Team)
	//TODO: need check user is a member of the tea
	return nil
}

func (m *ManagerRPC) AllowApp(arg ManagerAppArg, reply *ManagerAppReply) error {
	return NewTask("AllowApp", &AllowAppExecutor{arg, reply}).Run()
}

func (m *ManagerRPC) DisallowApp(arg ManagerAppArg, reply *ManagerAppReply) error {
	return NewTask("DisallowApp", &DisallowAppExecutor{arg, reply}).Run()
}

type IsAppAllowedExecutor struct {
	arg   ManagerIsAppAllowedArg
	reply *ManagerIsAppAllowedReply
}

func (e *IsAppAllowedExecutor) Request() interface{} {
	return e.arg
}

func (e *IsAppAllowedExecutor) Result() interface{} {
	return e.reply
}

func (e *IsAppAllowedExecutor) Description() string {
	return fmt.Sprintf("["+e.arg.ManagerAuthArg.User+"] %s : %s", e.arg.App, e.arg.User)
}

func (e *IsAppAllowedExecutor) Execute(t *Task) error {
	if aldap.SkipAuthorization {
		e.reply.IsAllowed = true
		return nil
	}

	var suReply ManagerSuperUserReply
	if err := NewTask("IsAppAllowed-IsSuperUser",
		&IsSuperUserExecutor{ManagerSuperUserArg{e.arg.ManagerAuthArg}, &suReply}).Run(); err != nil {
		return err
	}
	user := e.arg.ManagerAuthArg.User
	if suReply.IsSuperUser && e.arg.User != "" {
		user = e.arg.User
	}

	e.reply.IsAllowed = IsAppAllowed(&e.arg.ManagerAuthArg, user, e.arg.App)
	return nil
}

func (e *IsAppAllowedExecutor) Authorize() error {
	return nil
}

type ListAllowedAppsExecutor struct {
	arg   ManagerListAllowedAppsArg
	reply *ManagerListAllowedAppsReply
}

func (e *ListAllowedAppsExecutor) Request() interface{} {
	return e.arg
}

func (e *ListAllowedAppsExecutor) Result() interface{} {
	return e.reply
}

func (e *ListAllowedAppsExecutor) Description() string {
	return fmt.Sprintf("["+e.arg.ManagerAuthArg.User+"] %s: %s", e.arg.ManagerAuthArg.User, e.arg.User)
}

func (e *ListAllowedAppsExecutor) Execute(t *Task) error {
	if aldap.SkipAuthorization {
		e.reply.Apps = []string{"all"}
		return nil
	}

	var suReply ManagerSuperUserReply
	if err := NewTask("ListAllowedApps-IsSuperUser",
		&IsSuperUserExecutor{ManagerSuperUserArg{e.arg.ManagerAuthArg}, &suReply}).Run(); err != nil {
		return err
	}
	user := e.arg.ManagerAuthArg.User
	if suReply.IsSuperUser && e.arg.User != "" {
		user = e.arg.User
	}
	appMap := GetAllowedApps(&e.arg.ManagerAuthArg, user)
	e.reply.Apps = []string{}
	for app, isAllowed := range appMap {
		if isAllowed {
			e.reply.Apps = append(e.reply.Apps, app)
		}
	}
	sort.Strings(e.reply.Apps)
	return nil
}

func (e *ListAllowedAppsExecutor) Authorize() error {
	return nil
}

func GetAllowedApps(auth *ManagerAuthArg, user string) map[string]bool {
	result := map[string]bool{}

	teams, err := ListTeams(auth)
	
	if err != nil {
		return result
	}
	for _, team := range teams {
		teamApps := datamodel.GetTeamapps(team)
		for _, app := range teamApps.Apps {
			result[app] = true
		}
	}
	return result
}

func IsAppAllowed(auth *ManagerAuthArg, user string, app string) bool {
	var suReply ManagerSuperUserReply
	if err := NewTask("IsAppAllowed-IsSuperUser",
		&IsSuperUserExecutor{ManagerSuperUserArg{*auth}, &suReply}).Run(); err != nil {
		return false
	}
	if suReply.IsSuperUser {
		// shortcut for superusers
		return true
	}
	return GetAllowedApps(auth, user)[app]
}

// ----------------------------------------------------------------------------------------------------------
// Permissions
// ----------------------------------------------------------------------------------------------------------

type IsTeamAdminExecutor struct {
	arg   ManagerTeamAdminArg
	reply *ManagerTeamAdminReply
}

func (e *IsTeamAdminExecutor) Request() interface{} {
	return e.arg
}

func (e *IsTeamAdminExecutor) Result() interface{} {
	return e.reply
}

func (e *IsTeamAdminExecutor) Description() string {
	return fmt.Sprintf("["+e.arg.ManagerAuthArg.User+"] %s : %s", e.arg.User, e.arg.Team)
}

func (e *IsTeamAdminExecutor) Execute(t *Task) error {
	//no more team admin
	e.reply.IsAdmin = false
	return nil
	
}


func (e *IsTeamAdminExecutor) Authorize() error {
	return nil
}

type IsSuperUserExecutor struct {
	arg   ManagerSuperUserArg
	reply *ManagerSuperUserReply
}

func (e *IsSuperUserExecutor) Request() interface{} {
	return e.arg
}

func (e *IsSuperUserExecutor) Result() interface{} {
	return e.reply
}

func (e *IsSuperUserExecutor) Description() string {
	return fmt.Sprintf("["+e.arg.ManagerAuthArg.User+"] %s", e.arg.User)
}

func (e *IsSuperUserExecutor) Execute(t *Task) error {

	if aldap.SkipAuthorization {
		e.reply.IsSuperUser = true
		log.Println("isSuperUser is true because skip auth flag is set")
		return nil
	}

	if !UserExists(e.arg.User, &e.arg.ManagerAuthArg) {
		e.reply.IsSuperUser = false
		return nil
	}


	teams, err := ListTeams(&e.arg.ManagerAuthArg)
	if err != nil {
		e.reply.IsSuperUser = false
		return nil
	}

	if contains(teams, aldap.SuperUserGroup) {
		e.reply.IsSuperUser = true
	} else {
		e.reply.IsSuperUser = false
	}

	t.Log("-> %t", e.reply.IsSuperUser)
	return nil
}

func (e *IsSuperUserExecutor) Authorize() error {
	return nil
}

func (m *ManagerRPC) IsAppAllowed(arg ManagerIsAppAllowedArg, reply *ManagerIsAppAllowedReply) error {
	return NewTask("IsAppAllowed", &IsAppAllowedExecutor{arg, reply}).Run()
}

func (m *ManagerRPC) ListAllowedApps(arg ManagerListAllowedAppsArg, reply *ManagerListAllowedAppsReply) error {
	return NewTask("ListAllowedApps", &ListAllowedAppsExecutor{arg, reply}).Run()
}

func (m *ManagerRPC) IsTeamAdmin(arg ManagerTeamAdminArg, reply *ManagerTeamAdminReply) error {
	return NewTask("IsTeamAdmin", &IsTeamAdminExecutor{arg, reply}).Run()
}

func (m *ManagerRPC) IsSuperUser(arg ManagerSuperUserArg, reply *ManagerSuperUserReply) error {
	return NewTask("IsSuperUser", &IsSuperUserExecutor{arg, reply}).Run()
}

// ----------------------------------------------------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------------------------------------------------
func contains(arr []string, str string) bool {
   for _, a := range arr {
      if a == str {
         return true
      }
   }
   return false
}


func TeamExists(name string, auth *ManagerAuthArg) bool {
	if ret := aldap.LookupTeam(auth.User, auth.Secret); ret != nil {
		if contains(ret, name) {
			return true
		}
	}
	return false

}

func ListTeams(auth *ManagerAuthArg) ([]string, error) {

	if ret := aldap.LookupTeam(auth.User, auth.Secret); ret != nil {
		return ret, nil
	}
	return []string{}, nil 

}

func ListTeamAttributes(team, attribute string, ldapConn *ldap.LDAPConnection) ([]string, error) {
	filterStr := "(&(objectClass=" + aldap.TeamClass + ")(" + aldap.TeamCommonName + "=" + team + "))"
	
	searchReq := ldap.NewSimpleSearchRequest(aldap.BaseDomain, 2, filterStr, []string{attribute})
	sr, err := ldapConn.Search(searchReq)
	if err != nil {
		log.Println("Warning: error search team attribute for team ", team, " attribute ", attribute)
		log.Println("Warning: error search team attribute; error is ", err)
		return []string{}, err
	}

	ret := []string{}
	if err != nil || sr == nil {
		return ret, err
	}

	for _, entry := range sr.Entries {
		vals := entry.GetAttributeValues(attribute)
		if len(vals) > 0 {
			r, _ := regexp.Compile("^uid=([^,]+)")
			for _, dn := range vals {
				substrings := strings.Split(r.FindString(dn), "=")
				if len(substrings) == 2 {
					ret = append(ret, substrings[1])
                                }

			}
		}
	}
	return ret, nil
}

func ListTeamEmails(team string, auth *ManagerAuthArg) ([]string, error) {
	return []string{}, nil
}

func ListTeamAdmins(team string, auth *ManagerAuthArg) ([]string, error) {
	return []string{}, nil

}

func ListTeamMembers(team string, auth *ManagerAuthArg) ([]string, error) {

	ldapConn, err := aldap.CreateLdapConn(aldap.LdapServer, aldap.LdapPort, aldap.TlsConfig)
	if err != nil {
		log.Println("Warning: Unalbe to connect to ldap to fetch team members; ", err)
		return []string{}, nil
	}
	defer ldapConn.Close()

	err = aldap.LoginBind(aldap.SearchUserDn, aldap.SearchUserPwd, ldapConn)
	if err != nil {
		log.Println("Warning: Unalbe to bind to LDAP to fetch team members; ", err)
		return []string{}, nil
	}
	ret, err := ListTeamAttributes(team, aldap.UsernameAttr, ldapConn)

	return ret, err
}

func ListTeamApps(team string) ([]string, error) {

	teamApps := datamodel.GetTeamapps(team)
	return teamApps.Apps, nil
}

func UserExists(name string, auth *ManagerAuthArg) bool {
	//TODO check user against ldap 
	return true

}





