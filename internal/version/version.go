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

// Package version exposes the operator's build-time identity. The package
// variables Version, Commit and BuildDate are intended to be set with
// -ldflags at link time, e.g.
//
//	-ldflags "-X github.com/Kitio-Tek/athos-kubernetes/internal/version.Version=1.2.3 \
//	          -X github.com/Kitio-Tek/athos-kubernetes/internal/version.Commit=$(git rev-parse HEAD) \
//	          -X github.com/Kitio-Tek/athos-kubernetes/internal/version.BuildDate=$(date -u +%FT%TZ)"
//
// The accessors in this file produce structured and string forms suitable
// for log lines, status fields and HTTP user agents.
package version

import (
	"fmt"
	"runtime"
	"strings"
)

// Product is the canonical short name embedded in user-agent strings.
const Product = "athos-operator"

// Default placeholder values used when the binary is built without
// -ldflags (for example during `go test` or local development).
const (
	defaultVersion   = "v0.0.0-dev"
	defaultCommit    = "unknown"
	defaultBuildDate = "unknown"
)

// Version is the human-readable operator version. Override at link time.
var Version = defaultVersion

// Commit is the git commit SHA the binary was built from. Override at link time.
var Commit = defaultCommit

// BuildDate is an RFC3339 timestamp of when the binary was built. Override
// at link time.
var BuildDate = defaultBuildDate

// BuildInfo bundles the build-time metadata in a single struct so callers
// can serialise or log all fields at once.
type BuildInfo struct {
	// Version is the semantic version string (e.g. "v1.2.3").
	Version string `json:"version"`
	// Commit is the source revision identifier.
	Commit string `json:"commit"`
	// BuildDate is the build timestamp in RFC3339.
	BuildDate string `json:"buildDate"`
	// GoVersion records the Go toolchain that produced the binary.
	GoVersion string `json:"goVersion"`
	// Platform is GOOS/GOARCH.
	Platform string `json:"platform"`
}

// Info returns a snapshot of the current build metadata.
func Info() BuildInfo {
	return BuildInfo{
		Version:   Version,
		Commit:    Commit,
		BuildDate: BuildDate,
		GoVersion: runtime.Version(),
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

// String returns a single-line human-readable representation suitable for
// log lines such as "athos-operator v1.2.3 (abc1234) built 2026-04-01T...".
func String() string {
	return fmt.Sprintf("%s %s (%s) built %s on %s", Product, Version, shortCommit(Commit), BuildDate, runtime.Version())
}

// ShortVersion returns the major.minor portion of Version, dropping the
// leading "v" if present and any patch / pre-release suffix. If the input
// cannot be parsed, the original Version is returned unchanged.
func ShortVersion() string {
	v := strings.TrimPrefix(Version, "v")
	// Drop pre-release / build metadata.
	if i := strings.IndexAny(v, "-+"); i >= 0 {
		v = v[:i]
	}
	parts := strings.Split(v, ".")
	if len(parts) < 2 {
		return Version
	}
	return parts[0] + "." + parts[1]
}

// UserAgent returns an HTTP User-Agent string composed of the product name
// and short version, e.g. "athos-operator/1.2".
func UserAgent() string {
	return Product + "/" + ShortVersion()
}

// shortCommit returns the first 7 characters of a commit hash, or the
// input unchanged when shorter.
func shortCommit(c string) string {
	const shortLen = 7
	if len(c) <= shortLen {
		return c
	}
	return c[:shortLen]
}
