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

package clusterstate_test

import (
	"strings"
	"testing"

	"github.com/Kitio-Tek/athos-kubernetes/internal/clusterstate"
)

func TestEvaluate_DesiredState(t *testing.T) {
	d := clusterstate.Evaluate(clusterstate.Observed{
		DesiredReplicas: 3, CurrentReplicas: 3, ReadyReplicas: 3,
		PrimaryHealthy: true, HasPersistedData: true,
	})
	if d.Action != clusterstate.ActionNothing {
		t.Errorf("got %s, want Nothing", d.Action)
	}
	if !d.IsTerminal() {
		t.Errorf("Nothing should be terminal")
	}
}

func TestEvaluate_ScaleUp(t *testing.T) {
	d := clusterstate.Evaluate(clusterstate.Observed{
		DesiredReplicas: 3, CurrentReplicas: 1, ReadyReplicas: 1,
		PrimaryHealthy: true, HasPersistedData: true,
	})
	if d.Action != clusterstate.ActionScaleUp {
		t.Errorf("got %s, want ScaleUp", d.Action)
	}
}

func TestEvaluate_ScaleDown(t *testing.T) {
	d := clusterstate.Evaluate(clusterstate.Observed{
		DesiredReplicas: 1, CurrentReplicas: 3, ReadyReplicas: 3,
		PrimaryHealthy: true, HasPersistedData: true,
	})
	if d.Action != clusterstate.ActionScaleDown {
		t.Errorf("got %s, want ScaleDown", d.Action)
	}
}

func TestEvaluate_AwaitReady(t *testing.T) {
	d := clusterstate.Evaluate(clusterstate.Observed{
		DesiredReplicas: 3, CurrentReplicas: 3, ReadyReplicas: 1,
		PrimaryHealthy: true, HasPersistedData: true,
	})
	if d.Action != clusterstate.ActionAwaitReady {
		t.Errorf("got %s, want AwaitReady", d.Action)
	}
	if !d.IsTerminal() {
		t.Errorf("AwaitReady should be terminal")
	}
}

func TestEvaluate_UpdateImage(t *testing.T) {
	d := clusterstate.Evaluate(clusterstate.Observed{
		DesiredReplicas: 3, CurrentReplicas: 3, ReadyReplicas: 3,
		DesiredImage: "pg:17", CurrentImage: "pg:16",
		PrimaryHealthy: true, HasPersistedData: true,
	})
	if d.Action != clusterstate.ActionUpdateImage {
		t.Errorf("got %s", d.Action)
	}
	if !strings.Contains(d.Reason, "pg:17") {
		t.Errorf("reason should mention new image: %q", d.Reason)
	}
}

func TestEvaluate_Failover(t *testing.T) {
	d := clusterstate.Evaluate(clusterstate.Observed{
		DesiredReplicas: 3, CurrentReplicas: 3, ReadyReplicas: 2,
		PrimaryHealthy: false, HasPersistedData: true,
	})
	if d.Action != clusterstate.ActionFailover {
		t.Errorf("got %s, want Failover", d.Action)
	}
}

func TestEvaluate_RecoverWhenNoData(t *testing.T) {
	d := clusterstate.Evaluate(clusterstate.Observed{
		DesiredReplicas: 1, CurrentReplicas: 1,
		PrimaryHealthy:   false,
		HasPersistedData: false,
	})
	if d.Action != clusterstate.ActionRecover {
		t.Errorf("got %s, want Recover", d.Action)
	}
}

func TestDecision_String(t *testing.T) {
	d := clusterstate.Decision{Action: clusterstate.ActionNothing, Reason: "ok"}
	if got := d.String(); got != "Nothing: ok" {
		t.Errorf("String = %q", got)
	}
}

func TestDecision_IsTerminal(t *testing.T) {
	cases := map[clusterstate.Action]bool{
		clusterstate.ActionNothing:     true,
		clusterstate.ActionAwaitReady:  true,
		clusterstate.ActionScaleUp:     false,
		clusterstate.ActionScaleDown:   false,
		clusterstate.ActionUpdateImage: false,
		clusterstate.ActionFailover:    false,
		clusterstate.ActionRecover:     false,
	}
	for a, want := range cases {
		got := (clusterstate.Decision{Action: a}).IsTerminal()
		if got != want {
			t.Errorf("IsTerminal(%s) = %v, want %v", a, got, want)
		}
	}
}
