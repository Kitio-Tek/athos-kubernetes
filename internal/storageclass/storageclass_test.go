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

package storageclass_test

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/Kitio-Tek/athos-kubernetes/internal/storageclass"
)

func TestPVCTemplate_Defaults(t *testing.T) {
	pvc := storageclass.PVCTemplate(storageclass.VolumeRequest{
		Name: "data",
		Size: resource.MustParse("10Gi"),
	})
	if pvc.Name != "data" {
		t.Errorf("name = %q", pvc.Name)
	}
	if len(pvc.Spec.AccessModes) != 1 || pvc.Spec.AccessModes[0] != corev1.ReadWriteOnce {
		t.Errorf("access modes = %v", pvc.Spec.AccessModes)
	}
	if pvc.Spec.StorageClassName != nil {
		t.Errorf("expected nil StorageClassName for default class, got %v", pvc.Spec.StorageClassName)
	}
	if got := pvc.Spec.Resources.Requests[corev1.ResourceStorage]; got.Cmp(resource.MustParse("10Gi")) != 0 {
		t.Errorf("storage size = %s", got.String())
	}
}

func TestPVCTemplate_StorageClass(t *testing.T) {
	pvc := storageclass.PVCTemplate(storageclass.VolumeRequest{
		Name:             "data",
		Size:             resource.MustParse("1Gi"),
		StorageClassName: "ssd",
	})
	if pvc.Spec.StorageClassName == nil || *pvc.Spec.StorageClassName != "ssd" {
		t.Errorf("StorageClassName = %v", pvc.Spec.StorageClassName)
	}
}

func TestPVCTemplate_ExplicitAccessMode(t *testing.T) {
	pvc := storageclass.PVCTemplate(storageclass.VolumeRequest{
		Name:       "data",
		Size:       resource.MustParse("1Gi"),
		AccessMode: corev1.ReadOnlyMany,
	})
	if pvc.Spec.AccessModes[0] != corev1.ReadOnlyMany {
		t.Errorf("access mode = %v", pvc.Spec.AccessModes[0])
	}
}

func TestPVCTemplate_LabelsApplied(t *testing.T) {
	pvc := storageclass.PVCTemplate(storageclass.VolumeRequest{
		Name:   "data",
		Size:   resource.MustParse("1Gi"),
		Labels: map[string]string{"a": "b"},
	})
	if pvc.Labels["a"] != "b" {
		t.Errorf("label not applied: %v", pvc.Labels)
	}
}

func TestIsExpandable(t *testing.T) {
	if !storageclass.IsExpandable(resource.MustParse("1Gi"), resource.MustParse("2Gi")) {
		t.Error("2Gi from 1Gi should be expandable")
	}
	if storageclass.IsExpandable(resource.MustParse("2Gi"), resource.MustParse("1Gi")) {
		t.Error("shrink should not be expandable")
	}
	if storageclass.IsExpandable(resource.MustParse("1Gi"), resource.MustParse("1Gi")) {
		t.Error("equal sizes should not be expandable")
	}
}

func TestDefaultClassName_Empty(t *testing.T) {
	if storageclass.DefaultClassName() != "" {
		t.Errorf("DefaultClassName = %q, want empty", storageclass.DefaultClassName())
	}
}
