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

// Package exec wraps PSQL and pg_ctl invocations with helpers that build the
// command lines, streaming-friendly output, and structured error mapping.
//
// The package does NOT execute commands directly; it returns Command values
// that callers (typically the e2e/test helpers) can run via os/exec or via
// kubectl exec depending on context.
package exec

import (
	"fmt"
	"strings"
)

// Command represents a single command line and the environment it expects.
type Command struct {
	// Bin is the executable name. It is typically resolved against $PATH.
	Bin string
	// Args are the positional arguments passed to Bin.
	Args []string
	// Env is the slice of "KEY=VALUE" environment variables exported before
	// the command is invoked.
	Env []string
	// Stdin, if non-empty, is piped to the command's standard input.
	Stdin string
}

// String renders the command as a shell-escaped one-liner. The output is
// suitable for logging but is NOT safe to feed back into a shell.
func (c Command) String() string {
	parts := []string{c.Bin}
	for _, a := range c.Args {
		parts = append(parts, shellQuote(a))
	}
	return strings.Join(parts, " ")
}

func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	if !strings.ContainsAny(s, " '\"$\\\n\t") {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

// PSQLOptions controls how a psql Command is built.
type PSQLOptions struct {
	Host          string
	Port          int
	User          string
	Database      string
	Password      string
	SQL           string
	NoPsqlrc      bool
	TupleOnly     bool
	NonInteractive bool
	Variables     map[string]string
}

// PSQL returns a Command that runs psql with the requested options. The
// PGPASSWORD environment variable is set when Password is non-empty so the
// secret never appears on the command line.
func PSQL(opts PSQLOptions) Command {
	args := []string{}
	if opts.Host != "" {
		args = append(args, "-h", opts.Host)
	}
	if opts.Port > 0 {
		args = append(args, "-p", fmt.Sprintf("%d", opts.Port))
	}
	if opts.User != "" {
		args = append(args, "-U", opts.User)
	}
	if opts.Database != "" {
		args = append(args, "-d", opts.Database)
	}
	if opts.NoPsqlrc {
		args = append(args, "-X")
	}
	if opts.TupleOnly {
		args = append(args, "-t")
	}
	if opts.NonInteractive {
		args = append(args, "-q")
	}
	for k, v := range opts.Variables {
		args = append(args, "-v", fmt.Sprintf("%s=%s", k, v))
	}
	if opts.SQL != "" {
		args = append(args, "-c", opts.SQL)
	}
	cmd := Command{Bin: "psql", Args: args}
	if opts.Password != "" {
		cmd.Env = append(cmd.Env, "PGPASSWORD="+opts.Password)
	}
	return cmd
}

// PGCtlOptions controls how a pg_ctl Command is built.
type PGCtlOptions struct {
	DataDir string
	Action  string
	Mode    string
	Wait    bool
}

// PGCtl returns a Command that runs pg_ctl <action> with the given data
// directory and optional shutdown mode.
func PGCtl(opts PGCtlOptions) (Command, error) {
	if opts.DataDir == "" {
		return Command{}, fmt.Errorf("exec: pg_ctl requires DataDir")
	}
	if opts.Action == "" {
		return Command{}, fmt.Errorf("exec: pg_ctl requires Action")
	}
	args := []string{opts.Action, "-D", opts.DataDir}
	if opts.Mode != "" {
		args = append(args, "-m", opts.Mode)
	}
	if opts.Wait {
		args = append(args, "-w")
	}
	return Command{Bin: "pg_ctl", Args: args}, nil
}

// PGBaseBackup returns a Command that runs pg_basebackup against the given
// connection.
func PGBaseBackup(host string, port int, user, target string) Command {
	return Command{
		Bin: "pg_basebackup",
		Args: []string{
			"-h", host,
			"-p", fmt.Sprintf("%d", port),
			"-U", user,
			"-D", target,
			"-Fp",
			"-Xs",
			"-P",
			"-R",
		},
	}
}

// PGDump returns a Command that runs pg_dump.
func PGDump(host string, port int, user, db, target string) Command {
	return Command{
		Bin: "pg_dump",
		Args: []string{
			"-h", host,
			"-p", fmt.Sprintf("%d", port),
			"-U", user,
			"-d", db,
			"-Fc",
			"-f", target,
		},
	}
}

// SQLEscape returns s wrapped in single quotes with embedded single quotes
// doubled. It is suitable for embedding string literals in dynamically
// constructed SQL — but parameterised queries should be preferred.
func SQLEscape(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

// IdentEscape returns ident wrapped in double quotes with embedded double
// quotes doubled. Use for table or column names that originate from user
// input or CRD fields.
func IdentEscape(ident string) string {
	return "\"" + strings.ReplaceAll(ident, "\"", "\"\"") + "\""
}
