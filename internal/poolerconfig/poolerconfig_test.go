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

package poolerconfig_test

import (
	"strings"
	"testing"

	"github.com/Kitio-Tek/athos-kubernetes/internal/poolerconfig"
)

func TestRender_Defaults(t *testing.T) {
	out := poolerconfig.Render(poolerconfig.Spec{})
	for _, want := range []string{
		"listen_port = 6432",
		"pool_mode = transaction",
		"auth_type = scram-sha-256",
		"max_client_conn = 100",
		"default_pool_size = 20",
		"ignore_startup_parameters = extra_float_digits",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in output", want)
		}
	}
}

func TestRender_DatabasesSorted(t *testing.T) {
	out := poolerconfig.Render(poolerconfig.Spec{
		Databases: map[string]string{
			"b": "host=h dbname=b",
			"a": "host=h dbname=a",
		},
	})
	idxA := strings.Index(out, "a = ")
	idxB := strings.Index(out, "b = ")
	if idxA < 0 || idxB < 0 || idxA > idxB {
		t.Errorf("databases not sorted alphabetically: %q", out)
	}
}

func TestRender_AuthQueryAndUser(t *testing.T) {
	out := poolerconfig.Render(poolerconfig.Spec{
		AuthUser:  "athos_auth",
		AuthQuery: "SELECT user, pwd FROM athos.auth_users WHERE user=$1",
	})
	if !strings.Contains(out, "auth_user = athos_auth") {
		t.Error("missing auth_user")
	}
	if !strings.Contains(out, "auth_query = SELECT user") {
		t.Error("missing auth_query")
	}
}

func TestRender_OverridePort(t *testing.T) {
	out := poolerconfig.Render(poolerconfig.Spec{ListenPort: 7000})
	if !strings.Contains(out, "listen_port = 7000") {
		t.Errorf("port override lost: %q", out)
	}
}

func TestRenderUserList_Sorted(t *testing.T) {
	out := poolerconfig.RenderUserList(map[string]string{
		"bob":   "md5xyz",
		"alice": "md5abc",
	})
	if !strings.HasPrefix(out, "\"alice\"") {
		t.Errorf("not sorted: %q", out)
	}
}

func TestRenderUserList_Empty(t *testing.T) {
	if got := poolerconfig.RenderUserList(nil); got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestIsValidPoolMode(t *testing.T) {
	cases := map[poolerconfig.PoolMode]bool{
		poolerconfig.PoolModeSession:     true,
		poolerconfig.PoolModeTransaction: true,
		poolerconfig.PoolModeStatement:   true,
		"none":                           false,
	}
	for in, want := range cases {
		if got := poolerconfig.IsValidPoolMode(in); got != want {
			t.Errorf("IsValidPoolMode(%q) = %v, want %v", in, got, want)
		}
	}
}
