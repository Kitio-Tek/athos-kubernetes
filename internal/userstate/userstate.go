/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package userstate computes the SQL statements required to drive a
// PostgresUser to its desired state. The output is a deterministic ordered
// list of statements; the caller is responsible for executing them inside a
// transaction against the cluster's primary instance.
package userstate

import (
	"fmt"
	"sort"
)

// User describes the desired state captured by the PostgresUser CRD.
type User struct {
	Name             string
	Password         string
	Superuser        bool
	ConnectionLimit  int32
	Roles            []string
	GrantsByDatabase map[string][]string
}

// Plan emits the SQL statements needed to create or refresh a user.
// The first call against a fresh cluster will emit a CREATE USER ... and
// subsequent calls will emit ALTER USER on the same name.
func Plan(u User, exists bool) ([]string, error) {
	if u.Name == "" {
		return nil, fmt.Errorf("userstate: user name is required")
	}
	stmts := []string{}
	header := "CREATE USER"
	if exists {
		header = "ALTER USER"
	}
	stmts = append(stmts, fmt.Sprintf("%s %q WITH LOGIN", header, u.Name))
	if u.Password != "" {
		stmts = append(stmts, fmt.Sprintf("ALTER USER %q WITH PASSWORD '%s'", u.Name, u.Password))
	}
	if u.Superuser {
		stmts = append(stmts, fmt.Sprintf("ALTER USER %q SUPERUSER", u.Name))
	} else {
		stmts = append(stmts, fmt.Sprintf("ALTER USER %q NOSUPERUSER", u.Name))
	}
	if u.ConnectionLimit != 0 {
		stmts = append(stmts, fmt.Sprintf("ALTER USER %q CONNECTION LIMIT %d", u.Name, u.ConnectionLimit))
	}
	for _, role := range sorted(u.Roles) {
		stmts = append(stmts, fmt.Sprintf("GRANT %q TO %q", role, u.Name))
	}
	for _, db := range sortedKeys(u.GrantsByDatabase) {
		grants := u.GrantsByDatabase[db]
		for _, g := range sorted(grants) {
			stmts = append(stmts, fmt.Sprintf("GRANT %s ON DATABASE %q TO %q", g, db, u.Name))
		}
	}
	return stmts, nil
}

// PlanRevoke returns the SQL needed to revoke privileges before a
// PostgresUser is removed.
func PlanRevoke(u User) []string {
	stmts := []string{}
	for _, db := range sortedKeys(u.GrantsByDatabase) {
		stmts = append(stmts, fmt.Sprintf("REVOKE ALL ON DATABASE %q FROM %q", db, u.Name))
	}
	stmts = append(stmts, fmt.Sprintf("REASSIGN OWNED BY %q TO postgres", u.Name))
	stmts = append(stmts, fmt.Sprintf("DROP OWNED BY %q", u.Name))
	stmts = append(stmts, fmt.Sprintf("DROP USER IF EXISTS %q", u.Name))
	return stmts
}

func sorted(items []string) []string {
	out := append([]string(nil), items...)
	sort.Strings(out)
	return out
}

func sortedKeys(m map[string][]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
