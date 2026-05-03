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

package labels_test

import (
	"strings"
	"testing"

	"github.com/Kitio-Tek/athos-kubernetes/internal/labels"
)

func TestMerge_OverlayWins(t *testing.T) {
	base := labels.Set{"a": "1", "b": "2"}
	overlay := labels.Set{"b": "X", "c": "3"}
	got := labels.Merge(base, overlay)
	if got["a"] != "1" || got["b"] != "X" || got["c"] != "3" {
		t.Errorf("Merge result = %v", got)
	}
}

func TestMergeAll_OrderRespected(t *testing.T) {
	got := labels.MergeAll(labels.Set{"a": "1"}, labels.Set{"a": "2"}, labels.Set{"a": "3"})
	if got["a"] != "3" {
		t.Errorf("expected last writer wins, got %v", got)
	}
}

func TestEqual_SameKeysAndValues(t *testing.T) {
	if !labels.Equal(labels.Set{"a": "1"}, labels.Set{"a": "1"}) {
		t.Error("identical sets should be equal")
	}
	if labels.Equal(labels.Set{"a": "1"}, labels.Set{"a": "2"}) {
		t.Error("differing values should not be equal")
	}
	if labels.Equal(labels.Set{"a": "1"}, labels.Set{"a": "1", "b": "2"}) {
		t.Error("differing key counts should not be equal")
	}
}

func TestHasAll_Present(t *testing.T) {
	got := labels.Set{"a": "1", "b": "2", "c": "3"}
	want := labels.Set{"a": "1", "b": "2"}
	if !labels.HasAll(got, want) {
		t.Error("HasAll should be true when all want keys match")
	}
}

func TestHasAll_Missing(t *testing.T) {
	got := labels.Set{"a": "1"}
	want := labels.Set{"a": "1", "b": "2"}
	if labels.HasAll(got, want) {
		t.Error("HasAll should be false when a key is missing")
	}
}

func TestSelectorString_Sorted(t *testing.T) {
	s := labels.Set{"b": "2", "a": "1", "c": "3"}
	got := labels.SelectorString(s)
	if !strings.HasPrefix(got, "a=1,") {
		t.Errorf("selector should be alphabetically sorted, got %q", got)
	}
}

func TestIsValidKey(t *testing.T) {
	cases := []struct {
		key string
		ok  bool
	}{
		{"foo", true},
		{"foo.bar/name", true},
		{"app.kubernetes.io/managed-by", true},
		{"", false},
		{"-bad", false},
		{"bad-", false},
		{"foo bar", false},
		{"x/y/z", false},
	}
	for _, tc := range cases {
		if got := labels.IsValidKey(tc.key); got != tc.ok {
			t.Errorf("IsValidKey(%q) = %v, want %v", tc.key, got, tc.ok)
		}
	}
}

func TestIsValidValue(t *testing.T) {
	cases := []struct {
		v  string
		ok bool
	}{
		{"", true},
		{"v1", true},
		{"my-thing.0", true},
		{"-bad", false},
		{strings.Repeat("a", 64), false},
	}
	for _, tc := range cases {
		if got := labels.IsValidValue(tc.v); got != tc.ok {
			t.Errorf("IsValidValue(%q) = %v, want %v", tc.v, got, tc.ok)
		}
	}
}

func TestValidate_Failures(t *testing.T) {
	if err := labels.Validate(labels.Set{"-bad": "ok"}); err == nil {
		t.Error("expected validation error on bad key")
	}
	if err := labels.Validate(labels.Set{"good": strings.Repeat("a", 65)}); err == nil {
		t.Error("expected validation error on long value")
	}
	if err := labels.Validate(labels.Set{"good": "v"}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSubtract(t *testing.T) {
	a := labels.Set{"a": "1", "b": "2", "c": "3"}
	b := labels.Set{"b": "x"}
	got := labels.Subtract(a, b)
	if _, ok := got["b"]; ok {
		t.Errorf("Subtract should drop common keys, got %v", got)
	}
}

func TestKeys_Sorted(t *testing.T) {
	got := labels.Keys(labels.Set{"c": "1", "a": "2", "b": "3"})
	want := []string{"a", "b", "c"}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Keys[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
