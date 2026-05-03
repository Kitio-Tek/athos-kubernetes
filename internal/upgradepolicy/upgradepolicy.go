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

// Package upgradepolicy classifies PostgreSQL version transitions and
// decides whether they can be performed in-place or require a logical dump
// and restore.
package upgradepolicy

import "fmt"

// Class describes the kind of upgrade between two PostgreSQL major versions.
type Class string

const (
	// ClassNoChange means both versions match.
	ClassNoChange Class = "NoChange"
	// ClassMinor means only the patch level changes; in-place restart works.
	ClassMinor Class = "Minor"
	// ClassMajorInPlace means a major version step that pg_upgrade can handle.
	ClassMajorInPlace Class = "MajorInPlace"
	// ClassMajorLogical means the step is large enough that pg_dump/pg_restore
	// is required.
	ClassMajorLogical Class = "MajorLogical"
	// ClassDowngrade indicates the target version is older than the current
	// one. Downgrades are not supported in-place.
	ClassDowngrade Class = "Downgrade"
)

// Classify reports the upgrade class for a transition from current to target.
func Classify(current, target int) (Class, error) {
	if current <= 0 || target <= 0 {
		return "", fmt.Errorf("upgradepolicy: versions must be positive")
	}
	if current == target {
		return ClassNoChange, nil
	}
	if target < current {
		return ClassDowngrade, nil
	}
	step := target - current
	switch {
	case step == 0:
		return ClassNoChange, nil
	case step <= 1:
		return ClassMajorInPlace, nil
	case step <= 3:
		return ClassMajorInPlace, nil
	default:
		return ClassMajorLogical, nil
	}
}

// SupportedTargets returns the major versions reachable from current via an
// in-place upgrade. The slice is sorted ascending.
func SupportedTargets(current int) []int {
	if current <= 0 {
		return nil
	}
	out := []int{}
	for v := current; v <= current+3; v++ {
		out = append(out, v)
	}
	return out
}

// IsSafe returns nil if the target is reachable from current. The error
// message includes the recommended path otherwise.
func IsSafe(current, target int) error {
	c, err := Classify(current, target)
	if err != nil {
		return err
	}
	switch c {
	case ClassNoChange, ClassMinor, ClassMajorInPlace:
		return nil
	case ClassMajorLogical:
		return fmt.Errorf("upgradepolicy: %d -> %d requires logical dump+restore", current, target)
	case ClassDowngrade:
		return fmt.Errorf("upgradepolicy: cannot downgrade from %d to %d", current, target)
	}
	return fmt.Errorf("upgradepolicy: unknown class for %d -> %d", current, target)
}
