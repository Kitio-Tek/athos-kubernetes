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

package manifest_test

import (
	"strings"
	"testing"

	"github.com/Kitio-Tek/athos-kubernetes/internal/manifest"
)

const sample = `apiVersion: v1
kind: ConfigMap
metadata:
  name: a
---
apiVersion: v1
kind: Service
metadata:
  name: b
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: c
`

func TestSplit_Count(t *testing.T) {
	docs, err := manifest.Split(strings.NewReader(sample))
	if err != nil {
		t.Fatalf("Split: %v", err)
	}
	if len(docs) != 3 {
		t.Errorf("expected 3 docs, got %d", len(docs))
	}
	if docs[0].Index != 1 || docs[2].Index != 3 {
		t.Errorf("indices wrong: %+v", docs)
	}
}

func TestSplit_DropsEmpty(t *testing.T) {
	input := "---\n\n---\nkind: Foo\n---\n   \n---\n"
	docs, err := manifest.Split(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Split: %v", err)
	}
	if len(docs) != 1 {
		t.Errorf("empty docs should be dropped, got %d", len(docs))
	}
}

func TestJoin_Roundtrip(t *testing.T) {
	docs, err := manifest.Split(strings.NewReader(sample))
	if err != nil {
		t.Fatalf("Split: %v", err)
	}
	joined := manifest.Join(docs)
	if !strings.Contains(joined, "kind: Service") {
		t.Errorf("Join output missing Service: %q", joined)
	}
	if strings.Count(joined, "---\n") != 2 {
		t.Errorf("Join should produce 2 separators, got %q", joined)
	}
}

func TestJoin_Empty(t *testing.T) {
	if manifest.Join(nil) != "" {
		t.Error("Join(nil) should be empty")
	}
}

func TestFilter(t *testing.T) {
	docs, _ := manifest.Split(strings.NewReader(sample))
	deployments := manifest.Filter(docs, func(d manifest.Document) bool {
		return strings.Contains(d.Body, "kind: Deployment")
	})
	if len(deployments) != 1 {
		t.Errorf("expected 1 deployment, got %d", len(deployments))
	}
}

func TestHas(t *testing.T) {
	docs, _ := manifest.Split(strings.NewReader(sample))
	if !manifest.Has(docs, "Service") {
		t.Error("Has(Service) should be true")
	}
	if manifest.Has(docs, "Pod") {
		t.Error("Has(Pod) should be false")
	}
}

func TestCountByKind(t *testing.T) {
	docs, _ := manifest.Split(strings.NewReader(sample))
	counts := manifest.CountByKind(docs)
	if counts["ConfigMap"] != 1 || counts["Service"] != 1 || counts["Deployment"] != 1 {
		t.Errorf("counts = %+v", counts)
	}
}

func TestCountByKind_RepeatedKinds(t *testing.T) {
	input := "kind: Pod\n---\nkind: Pod\n---\nkind: Service\n"
	docs, _ := manifest.Split(strings.NewReader(input))
	counts := manifest.CountByKind(docs)
	if counts["Pod"] != 2 {
		t.Errorf("expected 2 Pods, got %d", counts["Pod"])
	}
}
