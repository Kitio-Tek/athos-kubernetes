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

// Package probe builds the corev1.Probe definitions used by the operator
// for liveness, readiness and startup checks against PostgreSQL pods.
//
// The package centralises the choice of command, port and timing constants so
// callers (StatefulSet builders and helm charts) do not duplicate them.
package probe

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// Default timing values, expressed in seconds, applied to every probe unless
// overridden by the caller. Values were chosen to work with PostgreSQL's
// startup time on a fresh PVC and to fail over within a single replication
// cycle when an instance becomes unreachable.
const (
	DefaultInitialDelaySeconds = 30
	DefaultPeriodSeconds       = 10
	DefaultTimeoutSeconds      = 5
	DefaultSuccessThreshold    = 1
	DefaultFailureThreshold    = 6

	// PostgresPort is the TCP port the postmaster listens on inside the pod.
	PostgresPort = 5432
)

// Timing groups the four duration knobs every kubelet probe takes. Zero
// values mean "use the default".
type Timing struct {
	InitialDelaySeconds int32
	PeriodSeconds       int32
	TimeoutSeconds      int32
	SuccessThreshold    int32
	FailureThreshold    int32
}

func (t Timing) withDefaults() Timing {
	if t.InitialDelaySeconds == 0 {
		t.InitialDelaySeconds = DefaultInitialDelaySeconds
	}
	if t.PeriodSeconds == 0 {
		t.PeriodSeconds = DefaultPeriodSeconds
	}
	if t.TimeoutSeconds == 0 {
		t.TimeoutSeconds = DefaultTimeoutSeconds
	}
	if t.SuccessThreshold == 0 {
		t.SuccessThreshold = DefaultSuccessThreshold
	}
	if t.FailureThreshold == 0 {
		t.FailureThreshold = DefaultFailureThreshold
	}
	return t
}

// LivenessProbe returns a Probe that runs pg_isready against the local
// instance. A failure will cause the kubelet to restart the container.
func LivenessProbe(t Timing) *corev1.Probe {
	t = t.withDefaults()
	return &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			Exec: &corev1.ExecAction{
				Command: []string{"pg_isready", "-U", "postgres", "-h", "127.0.0.1"},
			},
		},
		InitialDelaySeconds: t.InitialDelaySeconds,
		PeriodSeconds:       t.PeriodSeconds,
		TimeoutSeconds:      t.TimeoutSeconds,
		SuccessThreshold:    t.SuccessThreshold,
		FailureThreshold:    t.FailureThreshold,
	}
}

// ReadinessProbe returns a Probe that confirms the instance is accepting
// connections. Failures temporarily remove the pod from Service endpoints.
func ReadinessProbe(t Timing) *corev1.Probe {
	t = t.withDefaults()
	return &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			Exec: &corev1.ExecAction{
				Command: []string{"pg_isready", "-U", "postgres", "-h", "127.0.0.1"},
			},
		},
		InitialDelaySeconds: t.InitialDelaySeconds,
		PeriodSeconds:       t.PeriodSeconds,
		TimeoutSeconds:      t.TimeoutSeconds,
		SuccessThreshold:    t.SuccessThreshold,
		FailureThreshold:    t.FailureThreshold,
	}
}

// StartupProbe returns a Probe that gives a fresh pod time to initialize
// before liveness probes start running. Useful when restoring large WAL
// archives at startup.
func StartupProbe(t Timing) *corev1.Probe {
	t = t.withDefaults()
	if t.FailureThreshold < 30 {
		t.FailureThreshold = 30
	}
	return &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			TCPSocket: &corev1.TCPSocketAction{Port: intstr.FromInt(PostgresPort)},
		},
		InitialDelaySeconds: t.InitialDelaySeconds,
		PeriodSeconds:       t.PeriodSeconds,
		TimeoutSeconds:      t.TimeoutSeconds,
		SuccessThreshold:    t.SuccessThreshold,
		FailureThreshold:    t.FailureThreshold,
	}
}

// PgBouncerLivenessProbe returns a TCP probe targeting the PgBouncer port.
func PgBouncerLivenessProbe(port int32) *corev1.Probe {
	if port == 0 {
		port = 6432
	}
	t := Timing{}.withDefaults()
	return &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			TCPSocket: &corev1.TCPSocketAction{Port: intstr.FromInt(int(port))},
		},
		InitialDelaySeconds: t.InitialDelaySeconds,
		PeriodSeconds:       t.PeriodSeconds,
		TimeoutSeconds:      t.TimeoutSeconds,
		SuccessThreshold:    t.SuccessThreshold,
		FailureThreshold:    t.FailureThreshold,
	}
}

// PgBouncerReadinessProbe is currently identical to the liveness probe; it is
// kept as a separate constructor so future probe shapes can diverge without
// changing call sites.
func PgBouncerReadinessProbe(port int32) *corev1.Probe {
	return PgBouncerLivenessProbe(port)
}

// HTTPGetProbe returns a generic HTTP-based probe pointing at the given path
// and port on the local pod. Used by the manager metrics endpoint.
func HTTPGetProbe(path string, port int32, t Timing) *corev1.Probe {
	t = t.withDefaults()
	return &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: path,
				Port: intstr.FromInt(int(port)),
			},
		},
		InitialDelaySeconds: t.InitialDelaySeconds,
		PeriodSeconds:       t.PeriodSeconds,
		TimeoutSeconds:      t.TimeoutSeconds,
		SuccessThreshold:    t.SuccessThreshold,
		FailureThreshold:    t.FailureThreshold,
	}
}
