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

package upgradepolicy_test

import (
	"strings"
	"testing"

	"github.com/Kitio-Tek/athos-kubernetes/internal/upgradepolicy"
)

func TestClassify_NoChange(t *testing.T) {
	c, err := upgradepolicy.Classify(16, 16)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if c != upgradepolicy.ClassNoChange {
		t.Errorf("got %s", c)
	}
}

func TestClassify_OneMajorIsInPlace(t *testing.T) {
	c, _ := upgradepolicy.Classify(16, 17)
	if c != upgradepolicy.ClassMajorInPlace {
		t.Errorf("got %s", c)
	}
}

func TestClassify_BigStepIsLogical(t *testing.T) {
	c, _ := upgradepolicy.Classify(12, 17)
	if c != upgradepolicy.ClassMajorLogical {
		t.Errorf("got %s", c)
	}
}

func TestClassify_Downgrade(t *testing.T) {
	c, _ := upgradepolicy.Classify(17, 16)
	if c != upgradepolicy.ClassDowngrade {
		t.Errorf("got %s", c)
	}
}

func TestClassify_RejectsZeroOrNegative(t *testing.T) {
	if _, err := upgradepolicy.Classify(0, 1); err == nil {
		t.Error("expected error")
	}
	if _, err := upgradepolicy.Classify(1, -1); err == nil {
		t.Error("expected error")
	}
}

func TestSupportedTargets(t *testing.T) {
	got := upgradepolicy.SupportedTargets(16)
	if len(got) != 4 {
		t.Fatalf("expected 4, got %v", got)
	}
	if got[0] != 16 || got[3] != 19 {
		t.Errorf("range = %v", got)
	}
}

func TestSupportedTargets_Zero(t *testing.T) {
	if got := upgradepolicy.SupportedTargets(0); got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestIsSafe_OK(t *testing.T) {
	if err := upgradepolicy.IsSafe(16, 17); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestIsSafe_LogicalRequired(t *testing.T) {
	err := upgradepolicy.IsSafe(12, 17)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "dump") {
		t.Errorf("error message = %q", err.Error())
	}
}

func TestIsSafe_Downgrade(t *testing.T) {
	if err := upgradepolicy.IsSafe(17, 16); err == nil {
		t.Error("expected error")
	}
}
