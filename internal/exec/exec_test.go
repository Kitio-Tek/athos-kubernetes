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

package exec_test

import (
	"strings"
	"testing"

	pgexec "github.com/Kitio-Tek/athos-kubernetes/internal/exec"
)

func TestCommand_String(t *testing.T) {
	c := pgexec.Command{Bin: "psql", Args: []string{"-c", "select 1"}}
	got := c.String()
	if !strings.Contains(got, "psql") || !strings.Contains(got, "'select 1'") {
		t.Errorf("String = %q", got)
	}
}

func TestPSQL_BuildsArgs(t *testing.T) {
	c := pgexec.PSQL(pgexec.PSQLOptions{
		Host:     "h",
		Port:     5432,
		User:     "u",
		Database: "d",
		SQL:      "select 1",
		Password: "secret",
	})
	if c.Bin != "psql" {
		t.Errorf("Bin = %q", c.Bin)
	}
	args := strings.Join(c.Args, " ")
	for _, want := range []string{"-h h", "-p 5432", "-U u", "-d d", "-c select 1"} {
		if !strings.Contains(args, want) {
			t.Errorf("missing %q in args: %q", want, args)
		}
	}
	if len(c.Env) != 1 || !strings.HasPrefix(c.Env[0], "PGPASSWORD=") {
		t.Errorf("env = %v", c.Env)
	}
}

func TestPSQL_PasswordOnlyWhenSet(t *testing.T) {
	c := pgexec.PSQL(pgexec.PSQLOptions{User: "u"})
	if len(c.Env) != 0 {
		t.Errorf("expected no env, got %v", c.Env)
	}
}

func TestPSQL_FlagsApplied(t *testing.T) {
	c := pgexec.PSQL(pgexec.PSQLOptions{
		User: "u", NoPsqlrc: true, TupleOnly: true, NonInteractive: true,
		Variables: map[string]string{"K": "V"},
	})
	args := strings.Join(c.Args, " ")
	for _, want := range []string{"-X", "-t", "-q", "-v K=V"} {
		if !strings.Contains(args, want) {
			t.Errorf("missing %q in args: %q", want, args)
		}
	}
}

func TestPGCtl_RequiresFields(t *testing.T) {
	if _, err := pgexec.PGCtl(pgexec.PGCtlOptions{Action: "start"}); err == nil {
		t.Error("expected error when DataDir missing")
	}
	if _, err := pgexec.PGCtl(pgexec.PGCtlOptions{DataDir: "/data"}); err == nil {
		t.Error("expected error when Action missing")
	}
}

func TestPGCtl_BuildsArgs(t *testing.T) {
	c, err := pgexec.PGCtl(pgexec.PGCtlOptions{
		DataDir: "/var/lib/pgdata",
		Action:  "stop",
		Mode:    "fast",
		Wait:    true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	args := strings.Join(c.Args, " ")
	for _, want := range []string{"stop", "-D /var/lib/pgdata", "-m fast", "-w"} {
		if !strings.Contains(args, want) {
			t.Errorf("missing %q in args: %q", want, args)
		}
	}
}

func TestPGBaseBackup(t *testing.T) {
	c := pgexec.PGBaseBackup("h", 5432, "u", "/target")
	if c.Bin != "pg_basebackup" {
		t.Errorf("Bin = %q", c.Bin)
	}
	args := strings.Join(c.Args, " ")
	for _, want := range []string{"-h h", "-D /target", "-Fp", "-Xs", "-R"} {
		if !strings.Contains(args, want) {
			t.Errorf("missing %q in args: %q", want, args)
		}
	}
}

func TestPGDump(t *testing.T) {
	c := pgexec.PGDump("h", 5432, "u", "d", "out.dump")
	args := strings.Join(c.Args, " ")
	for _, want := range []string{"-h h", "-d d", "-Fc", "-f out.dump"} {
		if !strings.Contains(args, want) {
			t.Errorf("missing %q in args: %q", want, args)
		}
	}
}

func TestSQLEscape(t *testing.T) {
	got := pgexec.SQLEscape("it's fine")
	want := "'it''s fine'"
	if got != want {
		t.Errorf("SQLEscape = %q, want %q", got, want)
	}
}

func TestIdentEscape(t *testing.T) {
	got := pgexec.IdentEscape(`bad"name`)
	want := `"bad""name"`
	if got != want {
		t.Errorf("IdentEscape = %q, want %q", got, want)
	}
}
