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

package patcher

import (
	"encoding/json"
	"testing"
)

func TestCreateMergePatch_NoChange(t *testing.T) {
	obj := map[string]interface{}{"key": "value", "num": float64(1)}
	orig, _ := json.Marshal(obj)
	mod, _ := json.Marshal(obj)
	patch, err := createMergePatch(orig, mod)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var patchMap map[string]interface{}
	if err := json.Unmarshal(patch, &patchMap); err != nil {
		t.Fatalf("unmarshal patch: %v", err)
	}
	if len(patchMap) != 0 {
		t.Errorf("expected empty patch, got %v", patchMap)
	}
}

func TestCreateMergePatch_ValueChanged(t *testing.T) {
	orig, _ := json.Marshal(map[string]interface{}{"replicas": float64(1)})
	mod, _ := json.Marshal(map[string]interface{}{"replicas": float64(3)})
	patch, err := createMergePatch(orig, mod)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var patchMap map[string]interface{}
	_ = json.Unmarshal(patch, &patchMap)
	if patchMap["replicas"] != float64(3) {
		t.Errorf("expected replicas=3 in patch, got %v", patchMap)
	}
}

func TestCreateMergePatch_KeyAdded(t *testing.T) {
	orig, _ := json.Marshal(map[string]interface{}{"a": "1"})
	mod, _ := json.Marshal(map[string]interface{}{"a": "1", "b": "2"})
	patch, err := createMergePatch(orig, mod)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var patchMap map[string]interface{}
	_ = json.Unmarshal(patch, &patchMap)
	if patchMap["b"] != "2" {
		t.Errorf("expected b=2 in patch, got %v", patchMap)
	}
	if _, ok := patchMap["a"]; ok {
		t.Errorf("expected a to be absent from patch, got %v", patchMap)
	}
}

func TestCreateMergePatch_KeyRemoved(t *testing.T) {
	orig, _ := json.Marshal(map[string]interface{}{"a": "1", "b": "2"})
	mod, _ := json.Marshal(map[string]interface{}{"a": "1"})
	patch, err := createMergePatch(orig, mod)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var patchMap map[string]interface{}
	_ = json.Unmarshal(patch, &patchMap)
	if patchMap["b"] != nil {
		t.Errorf("expected b=null in patch, got %v", patchMap["b"])
	}
}

func TestCreateMergePatch_NestedChange(t *testing.T) {
	orig, _ := json.Marshal(map[string]interface{}{
		"spec": map[string]interface{}{"replicas": float64(1)},
	})
	mod, _ := json.Marshal(map[string]interface{}{
		"spec": map[string]interface{}{"replicas": float64(2)},
	})
	patch, err := createMergePatch(orig, mod)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var patchMap map[string]interface{}
	_ = json.Unmarshal(patch, &patchMap)
	spec, ok := patchMap["spec"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected spec in patch, got %v", patchMap)
	}
	if spec["replicas"] != float64(2) {
		t.Errorf("expected spec.replicas=2, got %v", spec["replicas"])
	}
}

func TestDiffMaps_Empty(t *testing.T) {
	result := diffMaps(map[string]interface{}{}, map[string]interface{}{})
	if len(result) != 0 {
		t.Errorf("expected empty diff, got %v", result)
	}
}

func TestDiffMaps_NilValues(t *testing.T) {
	orig := map[string]interface{}{"key": nil}
	mod := map[string]interface{}{"key": nil}
	result := diffMaps(orig, mod)
	if len(result) != 0 {
		t.Errorf("expected empty diff for matching nil values, got %v", result)
	}
}

func TestNew(t *testing.T) {
	p := New(nil)
	if p == nil {
		t.Error("expected non-nil Patcher")
	}
}
