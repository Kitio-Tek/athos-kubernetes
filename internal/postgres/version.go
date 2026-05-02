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

package postgres

import "fmt"

const (
	// defaultImageRepo is the container registry prefix used for PostgreSQL images.
	defaultImageRepo = "docker.io/library/postgres"

	// exporterImageRepo is the container image for the postgres_exporter sidecar.
	exporterImageRepo = "quay.io/prometheuscommunity/postgres-exporter"

	// defaultExporterTag is the postgres_exporter version to deploy.
	defaultExporterTag = "v0.15.0"
)

// PostgresImageTag returns the fully qualified container image reference for the
// given PostgreSQL major version.
func PostgresImageTag(version int32) string {
	switch version {
	case 14:
		return fmt.Sprintf("%s:14-alpine", defaultImageRepo)
	case 15:
		return fmt.Sprintf("%s:15-alpine", defaultImageRepo)
	case 16:
		return fmt.Sprintf("%s:16-alpine", defaultImageRepo)
	case 17:
		return fmt.Sprintf("%s:17-alpine", defaultImageRepo)
	default:
		return fmt.Sprintf("%s:%d-alpine", defaultImageRepo, version)
	}
}

// ExporterImageTag returns the container image reference for the
// postgres_exporter Prometheus sidecar.
func ExporterImageTag() string {
	return fmt.Sprintf("%s:%s", exporterImageRepo, defaultExporterTag)
}
