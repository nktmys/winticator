package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsNewer(t *testing.T) {
	tests := []struct {
		name    string
		latest  string
		current string
		want    bool
	}{
		{
			name:    "same version",
			latest:  "1.0.0",
			current: "1.0.0",
			want:    false,
		},
		{
			name:    "newer major",
			latest:  "2.0.0",
			current: "1.0.0",
			want:    true,
		},
		{
			name:    "newer minor",
			latest:  "1.1.0",
			current: "1.0.0",
			want:    true,
		},
		{
			name:    "newer patch",
			latest:  "1.0.1",
			current: "1.0.0",
			want:    true,
		},
		{
			name:    "older version",
			latest:  "1.0.0",
			current: "2.0.0",
			want:    false,
		},
		{
			name:    "multi-digit version comparison - minor",
			latest:  "1.23.4",
			current: "1.5.6",
			want:    true,
		},
		{
			name:    "multi-digit version comparison - patch",
			latest:  "1.0.100",
			current: "1.0.99",
			want:    true,
		},
		{
			name:    "multi-digit all parts",
			latest:  "12.34.56",
			current: "12.34.55",
			want:    true,
		},
		{
			name:    "latest has more parts",
			latest:  "1.0.1",
			current: "1.0",
			want:    true,
		},
		{
			name:    "current has more parts",
			latest:  "1.0",
			current: "1.0.1",
			want:    false,
		},
		{
			name:    "version with pre-release suffix",
			latest:  "1.0.1-beta",
			current: "1.0.0",
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsNewer(tt.latest, tt.current)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    []int
	}{
		{
			name:    "standard version",
			version: "1.2.3",
			want:    []int{1, 2, 3},
		},
		{
			name:    "multi-digit",
			version: "12.34.56",
			want:    []int{12, 34, 56},
		},
		{
			name:    "with pre-release",
			version: "1.0.0-beta",
			want:    []int{1, 0, 0},
		},
		{
			name:    "two parts",
			version: "1.0",
			want:    []int{1, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Parse(tt.version)
			assert.Equal(t, tt.want, got)
		})
	}
}
