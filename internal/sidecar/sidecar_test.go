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

package sidecar_test

import (
	"strings"
	"testing"

	"github.com/Kitio-Tek/athos-kubernetes/internal/sidecar"
)

func TestExporterContainer_Defaults(t *testing.T) {
	c := sidecar.ExporterContainer(sidecar.ExporterSpec{DataSourceSecret: "creds"})
	if c.Name != "exporter" {
		t.Errorf("name = %q", c.Name)
	}
	if c.Image != sidecar.DefaultExporterImage {
		t.Errorf("image = %q", c.Image)
	}
	if len(c.Ports) != 1 || c.Ports[0].ContainerPort != sidecar.ExporterPort {
		t.Errorf("ports = %+v", c.Ports)
	}
	if c.LivenessProbe == nil || c.LivenessProbe.TCPSocket == nil {
		t.Error("expected TCP liveness probe")
	}
}

func TestExporterContainer_DataSourceEnvFromSecret(t *testing.T) {
	c := sidecar.ExporterContainer(sidecar.ExporterSpec{DataSourceSecret: "creds"})
	if len(c.Env) == 0 {
		t.Fatal("no env vars")
	}
	if c.Env[0].Name != "DATA_SOURCE_NAME" {
		t.Errorf("env name = %q", c.Env[0].Name)
	}
	if c.Env[0].ValueFrom == nil || c.Env[0].ValueFrom.SecretKeyRef == nil {
		t.Error("expected SecretKeyRef")
	}
	if c.Env[0].ValueFrom.SecretKeyRef.Name != "creds" {
		t.Errorf("secret = %q", c.Env[0].ValueFrom.SecretKeyRef.Name)
	}
}

func TestExporterContainer_KeyOverride(t *testing.T) {
	c := sidecar.ExporterContainer(sidecar.ExporterSpec{
		DataSourceSecret: "creds",
		DataSourceKey:    "custom-key",
	})
	if c.Env[0].ValueFrom.SecretKeyRef.Key != "custom-key" {
		t.Errorf("key = %q", c.Env[0].ValueFrom.SecretKeyRef.Key)
	}
}

func TestExporterContainer_ImageOverride(t *testing.T) {
	c := sidecar.ExporterContainer(sidecar.ExporterSpec{
		Image:            "custom/exporter:1",
		DataSourceSecret: "creds",
	})
	if c.Image != "custom/exporter:1" {
		t.Errorf("image override lost: %q", c.Image)
	}
}

func TestWALUploaderContainer_RequiresBucket(t *testing.T) {
	if _, err := sidecar.WALUploaderContainer(sidecar.WALUploaderSpec{}); err == nil {
		t.Error("expected error when bucket missing")
	}
}

func TestWALUploaderContainer_BasicShape(t *testing.T) {
	c, err := sidecar.WALUploaderContainer(sidecar.WALUploaderSpec{
		BucketURL: "s3://bucket/prefix",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Name != "wal-uploader" {
		t.Errorf("name = %q", c.Name)
	}
	if !strings.Contains(strings.Join(c.Command, " "), "wal-push-stream") {
		t.Errorf("command = %v", c.Command)
	}
	found := false
	for _, e := range c.Env {
		if e.Name == "WALG_S3_PREFIX" && e.Value == "s3://bucket/prefix" {
			found = true
		}
	}
	if !found {
		t.Errorf("WALG_S3_PREFIX env not set: %+v", c.Env)
	}
}

func TestWALUploaderContainer_WithCreds(t *testing.T) {
	c, err := sidecar.WALUploaderContainer(sidecar.WALUploaderSpec{
		BucketURL:   "s3://bucket",
		CredsSecret: "s3-creds",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := map[string]bool{
		"AWS_ACCESS_KEY_ID":     false,
		"AWS_SECRET_ACCESS_KEY": false,
	}
	for _, e := range c.Env {
		if _, ok := want[e.Name]; ok {
			want[e.Name] = true
		}
	}
	for k, found := range want {
		if !found {
			t.Errorf("env %q not found", k)
		}
	}
}
