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

package duration_test

import (
	"testing"
	"time"

	"github.com/Kitio-Tek/athos-kubernetes/internal/duration"
)

func TestParse_StandardUnits(t *testing.T) {
	got, err := duration.Parse("1h30m")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	want := 90 * time.Minute
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestParse_DayUnit(t *testing.T) {
	got, err := duration.Parse("3d")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got != 72*time.Hour {
		t.Errorf("3d = %v", got)
	}
}

func TestParse_WeekUnit(t *testing.T) {
	got, err := duration.Parse("2w")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got != 14*duration.Day {
		t.Errorf("2w = %v", got)
	}
}

func TestParse_MonthUnit(t *testing.T) {
	got, err := duration.Parse("1mo")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got != duration.Month {
		t.Errorf("1mo = %v", got)
	}
}

func TestParse_Compound(t *testing.T) {
	got, err := duration.Parse("1w2d3h")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	want := duration.Week + 2*duration.Day + 3*time.Hour
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestParse_Invalid(t *testing.T) {
	if _, err := duration.Parse(""); err == nil {
		t.Error("empty should error")
	}
	if _, err := duration.Parse("abc"); err == nil {
		t.Error("non-numeric should error")
	}
	if _, err := duration.Parse("12"); err == nil {
		t.Error("missing unit should error")
	}
	if _, err := duration.Parse("12xyz"); err == nil {
		t.Error("unknown unit should error")
	}
}

func TestMustParse_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()
	duration.MustParse("nope")
}

func TestFormat(t *testing.T) {
	cases := []struct {
		in   time.Duration
		want string
	}{
		{0, "0s"},
		{30 * time.Second, "30s"},
		{90 * time.Second, "1m30s"},
		{72 * time.Hour, "3d"},
		{73 * time.Hour, "3d1h"},
		{-time.Minute, "-1m"},
	}
	for _, tc := range cases {
		if got := duration.Format(tc.in); got != tc.want {
			t.Errorf("Format(%v) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestParse_NegativeStandard(t *testing.T) {
	got, err := duration.Parse("-30m")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got != -30*time.Minute {
		t.Errorf("got %v, want -30m", got)
	}
}
