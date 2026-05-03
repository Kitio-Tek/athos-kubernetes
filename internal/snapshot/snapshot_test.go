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

package snapshot_test

import (
	"strings"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Kitio-Tek/athos-kubernetes/internal/snapshot"
)

func TestBuild_Defaults(t *testing.T) {
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	m, err := snapshot.Build(snapshot.SnapshotSpec{
		ClusterName:       "pg",
		Namespace:         "default",
		SnapshotClassName: "csi-snap",
		SourcePVC:         "pgdata-pg-0",
	}, now)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if !strings.HasPrefix(m.Name, "pgdata-pg-0-snap-") {
		t.Errorf("Name prefix = %q", m.Name)
	}
	if m.Labels["pg.athos.io/cluster"] != "pg" {
		t.Error("cluster label missing")
	}
	if m.SnapshotClassName != "csi-snap" {
		t.Error("snapshotClassName lost")
	}
	if m.APIVersion != "snapshot.storage.k8s.io/v1" {
		t.Error("APIVersion not v1")
	}
}

func TestBuild_RequiresFields(t *testing.T) {
	if _, err := snapshot.Build(snapshot.SnapshotSpec{SourcePVC: "x"}, time.Now()); err == nil {
		t.Error("expected error without cluster name")
	}
	if _, err := snapshot.Build(snapshot.SnapshotSpec{ClusterName: "x"}, time.Now()); err == nil {
		t.Error("expected error without source PVC")
	}
}

func TestBuild_CustomSuffixUsed(t *testing.T) {
	now := time.Now()
	m, err := snapshot.Build(snapshot.SnapshotSpec{
		ClusterName: "pg",
		SourcePVC:   "pvc",
		Suffix:      "v42",
	}, now)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if !strings.HasSuffix(m.Name, "-v42") {
		t.Errorf("custom suffix not used: %q", m.Name)
	}
}

func TestRetention_MaxCount(t *testing.T) {
	now := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	snaps := makeSnaps(now, 5)
	policy := snapshot.RetentionPolicy{MaxCount: 2}
	deleted := policy.Apply(snaps, now)
	if len(deleted) != 3 {
		t.Errorf("expected 3 deleted, got %d", len(deleted))
	}
}

func TestRetention_MaxAge(t *testing.T) {
	now := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	snaps := makeSnaps(now, 5)
	// MaxAge=2 days: the 4th and 5th snapshots (3,4 days old) should be deleted.
	policy := snapshot.RetentionPolicy{MaxAge: 2 * 24 * time.Hour}
	deleted := policy.Apply(snaps, now)
	if len(deleted) != 2 {
		t.Errorf("expected 2 deleted by age, got %d", len(deleted))
	}
}

func TestRetention_NoLimits(t *testing.T) {
	now := time.Now()
	snaps := makeSnaps(now, 5)
	if d := (snapshot.RetentionPolicy{}).Apply(snaps, now); len(d) != 0 {
		t.Errorf("expected nothing deleted, got %d", len(d))
	}
}

func TestIsCompleted(t *testing.T) {
	yes := true
	no := false
	if !snapshot.IsCompleted(&yes) {
		t.Error("readyToUse=true should be completed")
	}
	if snapshot.IsCompleted(&no) {
		t.Error("readyToUse=false should not be completed")
	}
	if snapshot.IsCompleted(nil) {
		t.Error("nil readyToUse should not be completed")
	}
}

func makeSnaps(now time.Time, n int) []snapshot.Manifest {
	out := make([]snapshot.Manifest, n)
	for i := 0; i < n; i++ {
		out[i] = snapshot.Manifest{
			Name:              "s",
			CreationTimestamp: metav1.NewTime(now.Add(-time.Duration(i) * 24 * time.Hour)),
		}
	}
	return out
}
