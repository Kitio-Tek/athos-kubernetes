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

package healthcheck_test

import (
	"strings"
	"testing"

	"github.com/Kitio-Tek/vigil-kubernetes/internal/healthcheck"
)

func passing(name string) healthcheck.Probe {
	return healthcheck.FuncProbe{ProbeName: name, Fn: func() healthcheck.Check {
		return healthcheck.SimpleCheck(name, "ok", true, false)
	}}
}

func failing(name string, critical bool) healthcheck.Probe {
	return healthcheck.FuncProbe{ProbeName: name, Fn: func() healthcheck.Check {
		return healthcheck.SimpleCheck(name, "down", false, critical)
	}}
}

func unknown(name string) healthcheck.Probe {
	return healthcheck.FuncProbe{ProbeName: name, Fn: func() healthcheck.Check {
		return healthcheck.Check{Name: name, Status: healthcheck.StatusUnknown}
	}}
}

func TestReport_StatusEmpty(t *testing.T) {
	r := healthcheck.Report{}
	if got := r.Status(); got != healthcheck.StatusUnknown {
		t.Errorf("empty report status = %q, want Unknown", got)
	}
}

func TestReport_StatusAllPassing(t *testing.T) {
	r := healthcheck.Run("pod-0", passing("a"), passing("b"))
	if got := r.Status(); got != healthcheck.StatusPassing {
		t.Errorf("status = %q, want Passing", got)
	}
}

func TestReport_StatusUnknownWhenSomeUnknown(t *testing.T) {
	r := healthcheck.Run("pod-0", passing("a"), unknown("b"))
	if got := r.Status(); got != healthcheck.StatusUnknown {
		t.Errorf("status = %q, want Unknown", got)
	}
}

func TestReport_StatusFailingNonCritical(t *testing.T) {
	r := healthcheck.Run("pod-0", passing("a"), failing("b", false))
	if got := r.Status(); got != healthcheck.StatusFailing {
		t.Errorf("status = %q, want Failing", got)
	}
}

func TestReport_StatusFailingCritical(t *testing.T) {
	r := healthcheck.Run("pod-0", failing("a", true), passing("b"))
	if got := r.Status(); got != healthcheck.StatusFailing {
		t.Errorf("status = %q, want Failing", got)
	}
}

func TestReport_SortedChecks(t *testing.T) {
	r := healthcheck.Run("pod-0", passing("c"), passing("a"), passing("b"))
	sorted := r.SortedChecks()
	if got := []string{sorted[0].Name, sorted[1].Name, sorted[2].Name}; got[0] != "a" || got[1] != "b" || got[2] != "c" {
		t.Errorf("SortedChecks order = %v", got)
	}
}

func TestReport_Summary(t *testing.T) {
	r := healthcheck.Run("pod-0", passing("ok"), failing("bad", true))
	s := r.Summary()
	if !strings.Contains(s, "instance=pod-0") {
		t.Error("missing instance in summary")
	}
	if !strings.Contains(s, "bad=Failing") {
		t.Error("failing check should appear in summary")
	}
	if strings.Contains(s, "ok=Passing") {
		t.Error("passing checks should not appear in summary")
	}
}

func TestReport_FailingChecks(t *testing.T) {
	r := healthcheck.Run("pod-0", passing("ok"), failing("bad", false))
	failing := r.FailingChecks()
	if len(failing) != 1 || failing[0].Name != "bad" {
		t.Errorf("FailingChecks = %+v", failing)
	}
}

func TestRun_NilProbeIgnored(t *testing.T) {
	r := healthcheck.Run("pod-0", nil, passing("a"))
	if len(r.Checks) != 1 {
		t.Errorf("expected 1 check after nil probe, got %d", len(r.Checks))
	}
}

func TestRun_NilFuncProbe(t *testing.T) {
	r := healthcheck.Run("pod-0", healthcheck.FuncProbe{ProbeName: "broken"})
	if r.Checks[0].Status != healthcheck.StatusUnknown {
		t.Errorf("expected Unknown for nil-fn probe, got %q", r.Checks[0].Status)
	}
}

func TestCheck_Equal(t *testing.T) {
	a := healthcheck.SimpleCheck("a", "msg", true, true)
	b := healthcheck.SimpleCheck("a", "msg", true, true)
	if !a.Equal(b) {
		t.Error("identical checks should be equal")
	}
	c := healthcheck.SimpleCheck("a", "different", true, true)
	if a.Equal(c) {
		t.Error("checks with different message should not be equal")
	}
}

func TestCombineReports(t *testing.T) {
	r1 := healthcheck.Run("pod-0", passing("a"))
	r2 := healthcheck.Run("pod-1", failing("b", false))
	combined := healthcheck.CombineReports("cluster", r1, r2)
	if combined.Instance != "cluster" {
		t.Errorf("combined instance = %q", combined.Instance)
	}
	if len(combined.Checks) != 2 {
		t.Errorf("expected 2 combined checks, got %d", len(combined.Checks))
	}
}

func TestReport_Filter(t *testing.T) {
	r := healthcheck.Run("pod", passing("a"), failing("b", false), passing("c"))
	filtered := r.Filter(func(c healthcheck.Check) bool { return c.Status == healthcheck.StatusPassing })
	if len(filtered.Checks) != 2 {
		t.Errorf("expected 2 passing, got %d", len(filtered.Checks))
	}
}

func TestSimpleCheck_FailingIsFailing(t *testing.T) {
	c := healthcheck.SimpleCheck("x", "down", false, true)
	if c.Status != healthcheck.StatusFailing {
		t.Errorf("expected Failing, got %q", c.Status)
	}
}

func TestSimpleCheck_PassingIsPassing(t *testing.T) {
	c := healthcheck.SimpleCheck("x", "ok", true, false)
	if c.Status != healthcheck.StatusPassing {
		t.Errorf("expected Passing, got %q", c.Status)
	}
}

func TestRun_ObservedAtPopulated(t *testing.T) {
	r := healthcheck.Run("pod", passing("a"))
	if r.Checks[0].ObservedAt.IsZero() {
		t.Error("ObservedAt should be populated")
	}
}

func TestRun_DurationPopulated(t *testing.T) {
	r := healthcheck.Run("pod", passing("a"))
	if r.Checks[0].Duration < 0 {
		t.Error("Duration should be non-negative")
	}
}

func TestFuncProbe_Name(t *testing.T) {
	p := healthcheck.FuncProbe{ProbeName: "foo", Fn: func() healthcheck.Check { return healthcheck.Check{} }}
	if p.Name() != "foo" {
		t.Errorf("FuncProbe.Name = %q", p.Name())
	}
}
