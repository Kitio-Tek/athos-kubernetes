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

	"github.com/Kitio-Tek/athos-kubernetes/internal/sqlescape"
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
// subsequent calls will emit ALTER USER on the same name. All identifier
// and string-literal interpolation goes through internal/sqlescape so
// embedded quote characters cannot break out of the SQL grammar.
func Plan(u User, exists bool) ([]string, error) {
	if u.Name == "" {
		return nil, fmt.Errorf("userstate: user name is required")
	}
	if !sqlescape.IsValidIdentifier(u.Name) {
		return nil, fmt.Errorf("userstate: invalid user name %q", u.Name)
	}
	if err := sqlescape.AssertSafePassword(u.Password); err != nil {
		return nil, fmt.Errorf("userstate: %w", err)
	}

	name := sqlescape.Identifier(u.Name)
	stmts := []string{}
	header := "CREATE USER"
	if exists {
		header = "ALTER USER"
	}
	stmts = append(stmts, fmt.Sprintf("%s %s WITH LOGIN", header, name))
	if u.Password != "" {
		stmts = append(stmts, fmt.Sprintf("ALTER USER %s WITH PASSWORD %s",
			name, sqlescape.StringLiteral(u.Password)))
	}
	if u.Superuser {
		stmts = append(stmts, fmt.Sprintf("ALTER USER %s SUPERUSER", name))
	} else {
		stmts = append(stmts, fmt.Sprintf("ALTER USER %s NOSUPERUSER", name))
	}
	if u.ConnectionLimit != 0 {
		stmts = append(stmts, fmt.Sprintf("ALTER USER %s CONNECTION LIMIT %d", name, u.ConnectionLimit))
	}
	for _, role := range sorted(u.Roles) {
		if !sqlescape.IsValidIdentifier(role) {
			return nil, fmt.Errorf("userstate: invalid role name %q", role)
		}
		stmts = append(stmts, fmt.Sprintf("GRANT %s TO %s", sqlescape.Identifier(role), name))
	}
	for _, db := range sortedKeys(u.GrantsByDatabase) {
		if !sqlescape.IsValidIdentifier(db) {
			return nil, fmt.Errorf("userstate: invalid database name %q", db)
		}
		dbIdent := sqlescape.Identifier(db)
		grants := u.GrantsByDatabase[db]
		for _, g := range sorted(grants) {
			// Privilege names are restricted to a small enumerated set, but
			// validate as identifiers regardless to keep the same defence
			// in depth as user/database names.
			if !sqlescape.IsValidIdentifier(g) {
				return nil, fmt.Errorf("userstate: invalid privilege %q", g)
			}
			stmts = append(stmts, fmt.Sprintf("GRANT %s ON DATABASE %s TO %s", g, dbIdent, name))
		}
	}
	return stmts, nil
}

// PlanRevoke returns the SQL needed to revoke privileges before a
// PostgresUser is removed. Identifiers are escaped via internal/sqlescape
// so a user-supplied name with embedded quote characters cannot break out
// of the statement.
func PlanRevoke(u User) []string {
	if !sqlescape.IsValidIdentifier(u.Name) {
		return nil
	}
	name := sqlescape.Identifier(u.Name)
	stmts := []string{}
	for _, db := range sortedKeys(u.GrantsByDatabase) {
		if !sqlescape.IsValidIdentifier(db) {
			continue
		}
		stmts = append(stmts, fmt.Sprintf("REVOKE ALL ON DATABASE %s FROM %s",
			sqlescape.Identifier(db), name))
	}
	stmts = append(stmts, fmt.Sprintf("REASSIGN OWNED BY %s TO postgres", name))
	stmts = append(stmts, fmt.Sprintf("DROP OWNED BY %s", name))
	stmts = append(stmts, fmt.Sprintf("DROP USER IF EXISTS %s", name))
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
