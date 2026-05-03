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

// Package imageutil parses and manipulates container image references using
// the same grammar as Docker. It is a small subset of
// github.com/google/go-containerregistry/pkg/name and is sufficient for the
// operator's needs without pulling in the full dependency.
package imageutil

import (
	"errors"
	"fmt"
	"strings"
)

// Reference is a parsed container image reference.
type Reference struct {
	// Registry is the host-and-port portion (e.g. "ghcr.io"). Empty means
	// the default registry.
	Registry string
	// Repository is the path component (e.g. "kitio-tek/athos").
	Repository string
	// Tag is the human-readable label (e.g. "v1.2.3"). Mutually exclusive
	// with Digest in well-formed references but both are stored.
	Tag string
	// Digest is the immutable content digest including the algorithm
	// prefix (e.g. "sha256:abc...").
	Digest string
}

// String returns the canonical reference string. If both Tag and Digest are
// set, Digest takes precedence.
func (r Reference) String() string {
	out := r.Repository
	if r.Registry != "" {
		out = r.Registry + "/" + out
	}
	if r.Digest != "" {
		return out + "@" + r.Digest
	}
	if r.Tag != "" {
		return out + ":" + r.Tag
	}
	return out
}

// IsTagged reports whether the reference uses a tag (no digest).
func (r Reference) IsTagged() bool { return r.Tag != "" && r.Digest == "" }

// IsPinned reports whether the reference uses a digest.
func (r Reference) IsPinned() bool { return r.Digest != "" }

// Parse parses a string into a Reference. The empty string returns
// ErrEmptyReference.
func Parse(s string) (Reference, error) {
	if s == "" {
		return Reference{}, ErrEmptyReference
	}

	r := Reference{}
	rest := s

	// Split off digest if present.
	if i := strings.Index(rest, "@"); i >= 0 {
		r.Digest = rest[i+1:]
		rest = rest[:i]
	}

	// Split off tag if present, but only if the colon is in the last segment
	// (to avoid mistaking a registry port for a tag).
	if last := strings.LastIndex(rest, "/"); last >= 0 {
		head := rest[:last]
		tail := rest[last:]
		if i := strings.LastIndex(tail, ":"); i >= 0 {
			r.Tag = tail[i+1:]
			rest = head + tail[:i]
		}
	} else if i := strings.LastIndex(rest, ":"); i >= 0 {
		r.Tag = rest[i+1:]
		rest = rest[:i]
	}

	// Decide whether the leading segment is a registry. It is treated as
	// a registry if it contains a "." or ":" or matches "localhost".
	if i := strings.Index(rest, "/"); i > 0 {
		head := rest[:i]
		if strings.ContainsAny(head, ".:") || head == "localhost" {
			r.Registry = head
			rest = rest[i+1:]
		}
	}

	r.Repository = rest
	if r.Repository == "" {
		return Reference{}, fmt.Errorf("imageutil: no repository in %q", s)
	}
	return r, nil
}

// ErrEmptyReference is returned when Parse is called with the empty string.
var ErrEmptyReference = errors.New("imageutil: empty image reference")

// WithTag returns a copy of r with the Tag replaced and Digest cleared.
func (r Reference) WithTag(tag string) Reference {
	r.Tag = tag
	r.Digest = ""
	return r
}

// WithDigest returns a copy of r with the Digest replaced.
func (r Reference) WithDigest(digest string) Reference {
	r.Digest = digest
	return r
}

// PostgresImage builds a canonical reference for the official postgres
// image at a given major version: "docker.io/library/postgres:<v>-alpine".
func PostgresImage(major int) string {
	return fmt.Sprintf("docker.io/library/postgres:%d-alpine", major)
}
