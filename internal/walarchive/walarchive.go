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

// Package walarchive contains helpers used by the cluster controller and the
// recovery flow to construct PostgreSQL archive_command and restore_command
// strings against the supported object stores.
//
// The package is provider-agnostic: each store is described by a small
// Endpoint struct and the helpers compose shell commands around tools such
// as wal-g and barman-cloud-wal-archive. Callers can swap providers by
// changing the Endpoint and reusing the same archive/restore command.
package walarchive

import (
	"fmt"
	"strings"
)

// Provider identifies a supported object-storage backend for WAL files.
type Provider string

const (
	// ProviderS3 is any S3-compatible store (AWS S3, GCS via interop, MinIO).
	ProviderS3 Provider = "s3"
	// ProviderGCS is Google Cloud Storage with native GCS authentication.
	ProviderGCS Provider = "gcs"
	// ProviderAzure is Azure Blob Storage.
	ProviderAzure Provider = "azure"
	// ProviderFile is a local POSIX directory; useful for testing.
	ProviderFile Provider = "file"
)

// Endpoint describes the target object store. Not all fields apply to every
// provider; the Validate method only checks the fields that are required for
// the provider in use.
type Endpoint struct {
	Provider     Provider
	Bucket       string
	Prefix       string
	Region       string
	Endpoint     string
	StorageClass string
	Path         string
}

// Validate ensures the endpoint has the fields needed for its provider.
func (e Endpoint) Validate() error {
	switch e.Provider {
	case ProviderS3, ProviderGCS, ProviderAzure:
		if e.Bucket == "" {
			return fmt.Errorf("walarchive: bucket is required for provider %q", e.Provider)
		}
	case ProviderFile:
		if e.Path == "" {
			return fmt.Errorf("walarchive: path is required for provider file")
		}
	default:
		return fmt.Errorf("walarchive: unknown provider %q", e.Provider)
	}
	return nil
}

// URL returns the canonical wal-g style URL for the endpoint, e.g.
// s3://my-bucket/cluster/wal or file:///var/lib/walarchive.
func (e Endpoint) URL() string {
	switch e.Provider {
	case ProviderS3:
		return joinURL("s3://", e.Bucket, e.Prefix)
	case ProviderGCS:
		return joinURL("gs://", e.Bucket, e.Prefix)
	case ProviderAzure:
		return joinURL("azure://", e.Bucket, e.Prefix)
	case ProviderFile:
		return "file://" + strings.TrimRight(e.Path, "/")
	default:
		return ""
	}
}

func joinURL(scheme, bucket, prefix string) string {
	out := scheme + bucket
	prefix = strings.TrimLeft(strings.TrimSpace(prefix), "/")
	if prefix != "" {
		out += "/" + prefix
	}
	return out
}

// Tool identifies the binary used to push or pull WAL files.
type Tool string

const (
	// ToolWALG is the github.com/wal-g/wal-g project.
	ToolWALG Tool = "wal-g"
	// ToolBarman is barman-cloud-wal-archive.
	ToolBarman Tool = "barman"
)

// ArchiveCommand returns the value to set as PostgreSQL's archive_command
// configuration option. The %p placeholder is the file path of the WAL
// segment and %f is the segment file name; they are substituted by
// PostgreSQL at archive time, not by this helper.
func ArchiveCommand(tool Tool, ep Endpoint) (string, error) {
	if err := ep.Validate(); err != nil {
		return "", err
	}
	switch tool {
	case ToolWALG:
		return fmt.Sprintf("WALE_S3_PREFIX=%s wal-g wal-push %%p", ep.URL()), nil
	case ToolBarman:
		return fmt.Sprintf("barman-cloud-wal-archive --endpoint-url %s %s %%p",
			shellQuote(ep.Endpoint), shellQuote(ep.URL())), nil
	default:
		return "", fmt.Errorf("walarchive: unknown tool %q", tool)
	}
}

// RestoreCommand returns the value to set as PostgreSQL's restore_command
// configuration option for the given tool and endpoint.
func RestoreCommand(tool Tool, ep Endpoint) (string, error) {
	if err := ep.Validate(); err != nil {
		return "", err
	}
	switch tool {
	case ToolWALG:
		return fmt.Sprintf("WALE_S3_PREFIX=%s wal-g wal-fetch %%f %%p", ep.URL()), nil
	case ToolBarman:
		return fmt.Sprintf("barman-cloud-wal-restore --endpoint-url %s %s %%f %%p",
			shellQuote(ep.Endpoint), shellQuote(ep.URL())), nil
	default:
		return "", fmt.Errorf("walarchive: unknown tool %q", tool)
	}
}

func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	if !strings.ContainsAny(s, " '\"$\\") {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

// SegmentFileName returns the canonical 24-character WAL segment name
// used as %f by PostgreSQL: timeline (8 hex), log id (8 hex), seg id (8 hex).
func SegmentFileName(timeline, logID, segID uint32) string {
	return fmt.Sprintf("%08X%08X%08X", timeline, logID, segID)
}

// ParseSegmentName parses a 24-character WAL segment name back into its
// (timeline, logID, segID) components. It returns an error if name does not
// match the expected length or hex format.
func ParseSegmentName(name string) (timeline, logID, segID uint32, err error) {
	if len(name) != 24 {
		return 0, 0, 0, fmt.Errorf("walarchive: segment name must be 24 chars, got %d", len(name))
	}
	if _, err := fmt.Sscanf(name, "%08X%08X%08X", &timeline, &logID, &segID); err != nil {
		return 0, 0, 0, fmt.Errorf("walarchive: invalid segment name %q: %w", name, err)
	}
	return timeline, logID, segID, nil
}

// ProviderForScheme returns the Provider associated with a URL scheme.
func ProviderForScheme(scheme string) (Provider, bool) {
	switch strings.ToLower(scheme) {
	case "s3":
		return ProviderS3, true
	case "gs", "gcs":
		return ProviderGCS, true
	case "azure":
		return ProviderAzure, true
	case "file":
		return ProviderFile, true
	}
	return "", false
}
