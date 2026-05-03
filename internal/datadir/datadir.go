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

// Package datadir centralises the on-disk filesystem layout used inside the
// PostgreSQL container. Every other package that needs to know where the
// data, WAL or socket directories live should consult these constants
// rather than hard-coding paths.
package datadir

import (
	"path/filepath"
	"strings"
)

// Mount points inside the postgres container. The values are deliberately
// chosen to match the Debian/Alpine packaging conventions so that the
// upstream postgres binaries find them without additional flags.
const (
	// Root is the parent directory holding all per-cluster state.
	Root = "/var/lib/postgresql"
	// PGData is the PGDATA directory consumed by the postmaster.
	PGData = Root + "/data"
	// WAL is the directory used when WAL is moved to a separate volume.
	WAL = Root + "/wal"
	// Socket is the Unix socket directory for local connections.
	Socket = "/var/run/postgresql"
	// Scripts holds operator-managed bootstrap scripts.
	Scripts = "/opt/athos/scripts"
	// TLS holds mounted TLS certificates and keys.
	TLS = "/opt/athos/tls"
	// Config holds the operator-managed postgresql.conf and pg_hba.conf.
	Config = "/etc/postgresql"
)

// PGDataPath joins PGData with rel. Empty rel returns PGData unchanged.
func PGDataPath(rel string) string {
	if rel == "" {
		return PGData
	}
	return filepath.Join(PGData, strings.TrimPrefix(rel, "/"))
}

// WALPath joins WAL with rel.
func WALPath(rel string) string {
	if rel == "" {
		return WAL
	}
	return filepath.Join(WAL, strings.TrimPrefix(rel, "/"))
}

// ConfigPath joins Config with rel.
func ConfigPath(rel string) string {
	if rel == "" {
		return Config
	}
	return filepath.Join(Config, strings.TrimPrefix(rel, "/"))
}

// IsInsidePGData reports whether path is within PGData.
func IsInsidePGData(path string) bool {
	return strings.HasPrefix(filepath.Clean(path), PGData)
}

// PostgresqlConf returns the absolute path to the operator-managed
// postgresql.conf file.
func PostgresqlConf() string {
	return ConfigPath("postgresql.conf")
}

// PGHBAConf returns the absolute path to the operator-managed pg_hba.conf.
func PGHBAConf() string {
	return ConfigPath("pg_hba.conf")
}

// PIDFile returns the absolute path to PostgreSQL's pidfile.
func PIDFile() string {
	return PGDataPath("postmaster.pid")
}

// IsValidPGDataLayout returns nil if every required path inside an active
// PGDATA layout is present in the supplied set, or an error naming the
// first missing one. The set is typically the result of listing the
// volume's top-level entries.
func IsValidPGDataLayout(present map[string]bool) error {
	required := []string{
		"PG_VERSION",
		"base",
		"global",
		"pg_wal",
	}
	for _, name := range required {
		if !present[name] {
			return &MissingError{Name: name}
		}
	}
	return nil
}

// MissingError reports a missing required entry in PGDATA.
type MissingError struct{ Name string }

func (m *MissingError) Error() string {
	return "datadir: required entry missing from PGDATA: " + m.Name
}
