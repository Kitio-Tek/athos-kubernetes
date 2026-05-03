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

package secretkey_test

import (
	"strings"
	"testing"

	"github.com/Kitio-Tek/athos-kubernetes/internal/secretkey"
)

func newData() map[string][]byte {
	return map[string][]byte{
		secretkey.Username: []byte("postgres"),
		secretkey.Password: []byte("hunter2"),
		secretkey.Host:     []byte("pg-rw"),
		secretkey.Port:     []byte("5432"),
		secretkey.Database: []byte("postgres"),
		secretkey.URI:      []byte("postgresql://postgres:hunter2@pg-rw:5432/postgres"),
	}
}

func TestHas_AllRequired(t *testing.T) {
	if !secretkey.Has(newData()) {
		t.Error("expected complete data to pass Has check")
	}
}

func TestHas_MissingValue(t *testing.T) {
	d := newData()
	delete(d, secretkey.Database)
	if secretkey.Has(d) {
		t.Error("missing key should fail Has check")
	}
}

func TestHas_EmptyValueCountsAsMissing(t *testing.T) {
	d := newData()
	d[secretkey.Password] = []byte{}
	if secretkey.Has(d) {
		t.Error("empty password should fail Has check")
	}
}

func TestMissingKeys(t *testing.T) {
	d := newData()
	delete(d, secretkey.Host)
	delete(d, secretkey.Port)
	missing := secretkey.MissingKeys(d)
	if len(missing) != 2 {
		t.Fatalf("expected 2 missing, got %v", missing)
	}
}

func TestString(t *testing.T) {
	d := newData()
	if v, ok := secretkey.String(d, secretkey.Username); !ok || v != "postgres" {
		t.Errorf("String returned %q,%v", v, ok)
	}
	if _, ok := secretkey.String(d, "nope"); ok {
		t.Error("expected ok=false for missing key")
	}
}

func TestMustString_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()
	secretkey.MustString(newData(), "absent-key")
}

func TestMustString_OK(t *testing.T) {
	if got := secretkey.MustString(newData(), secretkey.Username); got != "postgres" {
		t.Errorf("MustString = %q", got)
	}
}

func TestLibpqURI_WithPassword(t *testing.T) {
	got := secretkey.LibpqURI("h", "5432", "db", "u", "p")
	if !strings.HasPrefix(got, "postgresql://u:p@h:5432/db") {
		t.Errorf("URI = %q", got)
	}
}

func TestLibpqURI_NoPassword(t *testing.T) {
	got := secretkey.LibpqURI("h", "5432", "db", "u", "")
	if got != "postgresql://u@h:5432/db" {
		t.Errorf("URI = %q", got)
	}
}

func TestJDBCFromLibpq_WithPassword(t *testing.T) {
	got := secretkey.JDBCFromLibpq("h", "5432", "db", "u", "p")
	if !strings.Contains(got, "jdbc:postgresql://h:5432/db?user=u&password=p") {
		t.Errorf("JDBC URL = %q", got)
	}
}

func TestJDBCFromLibpq_NoPassword(t *testing.T) {
	got := secretkey.JDBCFromLibpq("h", "5432", "db", "u", "")
	if !strings.HasSuffix(got, "?user=u") {
		t.Errorf("JDBC URL = %q", got)
	}
}
