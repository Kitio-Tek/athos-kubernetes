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

// Package clusterstate computes the next reconciliation action for a
// PostgresCluster from the observed state of its sub-resources. It is a pure
// function package: callers gather the inputs (the cluster spec, the current
// StatefulSet, the list of pods) and the package returns a Decision describing
// what the controller should do next.
//
// Keeping this logic in a separate package makes it possible to write
// table-driven unit tests over a wide range of degraded cluster states
// without spinning up an envtest control plane.
package clusterstate

import (
	"fmt"
)

// Action enumerates the high-level decisions the controller can make on a
// given reconcile pass.
type Action string

const (
	// ActionNothing means the cluster is in the desired state.
	ActionNothing Action = "Nothing"
	// ActionScaleUp means more replicas should be created.
	ActionScaleUp Action = "ScaleUp"
	// ActionScaleDown means replicas should be removed.
	ActionScaleDown Action = "ScaleDown"
	// ActionUpdateImage means a rolling restart is required to pick up
	// a new container image.
	ActionUpdateImage Action = "UpdateImage"
	// ActionFailover means the primary should be replaced by a replica.
	ActionFailover Action = "Failover"
	// ActionAwaitReady means the controller should wait without acting.
	ActionAwaitReady Action = "AwaitReady"
	// ActionRecover means the cluster has lost data integrity and must be
	// restored from a backup.
	ActionRecover Action = "Recover"
)

// Observed captures the current observed state of a cluster.
type Observed struct {
	DesiredReplicas  int32
	CurrentReplicas  int32
	ReadyReplicas    int32
	DesiredImage     string
	CurrentImage     string
	PrimaryHealthy   bool
	HasPersistedData bool
}

// Decision is the outcome of evaluating Observed against the desired state.
type Decision struct {
	Action Action
	Reason string
}

// String returns a "Action: Reason" representation suitable for log lines.
func (d Decision) String() string {
	return fmt.Sprintf("%s: %s", d.Action, d.Reason)
}

// Evaluate returns the next reconciliation Decision from the observed state.
// The order of checks matters: the most catastrophic states (recover, failover)
// are detected first.
func Evaluate(o Observed) Decision {
	if !o.HasPersistedData && o.CurrentReplicas > 0 {
		return Decision{
			Action: ActionRecover,
			Reason: "PVCs exist but contain no PostgreSQL data; restore from backup",
		}
	}
	if !o.PrimaryHealthy && o.ReadyReplicas > 0 {
		return Decision{
			Action: ActionFailover,
			Reason: "primary is unhealthy and at least one replica is ready",
		}
	}
	if o.CurrentImage != "" && o.DesiredImage != "" && o.CurrentImage != o.DesiredImage {
		return Decision{
			Action: ActionUpdateImage,
			Reason: fmt.Sprintf("image change from %q to %q", o.CurrentImage, o.DesiredImage),
		}
	}
	if o.DesiredReplicas > o.CurrentReplicas {
		return Decision{
			Action: ActionScaleUp,
			Reason: fmt.Sprintf("%d/%d replicas present", o.CurrentReplicas, o.DesiredReplicas),
		}
	}
	if o.DesiredReplicas < o.CurrentReplicas {
		return Decision{
			Action: ActionScaleDown,
			Reason: fmt.Sprintf("scale from %d to %d replicas", o.CurrentReplicas, o.DesiredReplicas),
		}
	}
	if o.ReadyReplicas < o.DesiredReplicas {
		return Decision{
			Action: ActionAwaitReady,
			Reason: fmt.Sprintf("waiting for %d more replicas to become ready",
				o.DesiredReplicas-o.ReadyReplicas),
		}
	}
	return Decision{
		Action: ActionNothing,
		Reason: "cluster is in the desired state",
	}
}

// IsTerminal returns true if the decision indicates no further work is
// required for this reconcile cycle.
func (d Decision) IsTerminal() bool {
	switch d.Action {
	case ActionNothing, ActionAwaitReady:
		return true
	}
	return false
}
