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

// Package storageclass keeps the per-cluster decisions about which
// StorageClass and access modes to use in one place. Reconcilers consult
// this package when they materialise PVCs so the same defaulting rules
// apply across the data, WAL and backup volumes.
package storageclass

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VolumeRequest captures the inputs the cluster reconciler has when it
// builds a PersistentVolumeClaim template.
type VolumeRequest struct {
	Name             string
	Size             resource.Quantity
	StorageClassName string
	AccessMode       corev1.PersistentVolumeAccessMode
	Labels           map[string]string
}

// DefaultAccessMode is the access mode used when the caller does not
// specify one. PostgreSQL is single-writer so ReadWriteOnce is the
// appropriate default.
const DefaultAccessMode = corev1.ReadWriteOnce

// PVCTemplate returns the volumeClaimTemplate entry for a StatefulSet.
// AccessMode falls back to DefaultAccessMode when unset.
func PVCTemplate(req VolumeRequest) corev1.PersistentVolumeClaim {
	mode := req.AccessMode
	if mode == "" {
		mode = DefaultAccessMode
	}
	pvc := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:   req.Name,
			Labels: req.Labels,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{mode},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: req.Size,
				},
			},
		},
	}
	if req.StorageClassName != "" {
		sc := req.StorageClassName
		pvc.Spec.StorageClassName = &sc
	}
	return pvc
}

// IsExpandable reports whether the PVC's size is being increased compared
// to the existing claim. This helper is used by the cluster controller to
// decide whether a hot expansion is possible (storage class must allow it)
// or whether the PVC needs to be re-created.
func IsExpandable(existing, desired resource.Quantity) bool {
	return desired.Cmp(existing) > 0
}

// DefaultClassName returns the value treated as "use the cluster default"
// — i.e. the empty string. Centralising the value avoids each call site
// inventing its own sentinel.
func DefaultClassName() string { return "" }
