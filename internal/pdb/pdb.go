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

// Package pdb builds the PodDisruptionBudget resources that protect
// PostgresCluster pods from voluntary disruptions during node drains
// and rolling updates.
package pdb

import (
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// Spec captures the inputs the operator needs to build a PDB.
type Spec struct {
	Name      string
	Namespace string
	Selector  map[string]string
	// MinAvailable is interpreted as a count when >0 and ignored otherwise.
	// Mutually exclusive with MaxUnavailable.
	MinAvailable int
	// MaxUnavailable is interpreted as a count when >0 and ignored
	// otherwise. Mutually exclusive with MinAvailable.
	MaxUnavailable int
	Labels         map[string]string
}

// Build returns the PodDisruptionBudget for the given spec. When neither
// MinAvailable nor MaxUnavailable is set, MaxUnavailable defaults to 1 so
// at least one pod is always preserved during a drain.
func Build(s Spec) *policyv1.PodDisruptionBudget {
	pdb := &policyv1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.Name,
			Namespace: s.Namespace,
			Labels:    s.Labels,
		},
		Spec: policyv1.PodDisruptionBudgetSpec{
			Selector: &metav1.LabelSelector{MatchLabels: s.Selector},
		},
	}
	switch {
	case s.MinAvailable > 0:
		min := intstr.FromInt(s.MinAvailable)
		pdb.Spec.MinAvailable = &min
	case s.MaxUnavailable > 0:
		max := intstr.FromInt(s.MaxUnavailable)
		pdb.Spec.MaxUnavailable = &max
	default:
		max := intstr.FromInt(1)
		pdb.Spec.MaxUnavailable = &max
	}
	return pdb
}

// RecommendedFor returns a sensible Spec for a cluster with the given
// instance count. The rule is "MinAvailable = N - 1 for N>=3, otherwise
// MaxUnavailable = 1".
func RecommendedFor(name, namespace string, instances int, selector, labels map[string]string) Spec {
	s := Spec{
		Name:      name,
		Namespace: namespace,
		Selector:  selector,
		Labels:    labels,
	}
	if instances >= 3 {
		s.MinAvailable = instances - 1
	} else {
		s.MaxUnavailable = 1
	}
	return s
}

// IsAdvisory reports whether the spec configures only an advisory PDB
// (MaxUnavailable=0 and MinAvailable=0). Advisory PDBs are useful when
// the operator wants the resource to exist for observability without
// actually enforcing a disruption budget.
func IsAdvisory(s Spec) bool {
	return s.MinAvailable == 0 && s.MaxUnavailable == 0
}
