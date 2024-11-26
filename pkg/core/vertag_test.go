package core

import (
	"testing"
)

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name     string
		tag1     string
		tag2     string
		expected bool
	}{
		{
			name:     "unstable versions - newer patch",
			tag1:     "/refs/tags/module/1.1.9-unstable",
			tag2:     "/refs/tags/module/1.1.10-unstable",
			expected: true,
		},
		{
			name:     "unstable versions - newer minor",
			tag1:     "/refs/tags/module/1.1.10-unstable",
			tag2:     "/refs/tags/module/1.2.0-unstable",
			expected: true,
		},
		{
			name:     "stable versions",
			tag1:     "/refs/tags/module/1.1.9",
			tag2:     "/refs/tags/module/1.1.10",
			expected: true,
		},
		{
			name:     "mixed stable and unstable",
			tag1:     "/refs/tags/module/1.1.9",
			tag2:     "/refs/tags/module/1.1.9-unstable",
			expected: false,
		},
		{
			name:     "equal versions",
			tag1:     "/refs/tags/module/1.1.9-unstable",
			tag2:     "/refs/tags/module/1.1.9-unstable",
			expected: false,
		},
		{
			name:     "first version greater",
			tag1:     "/refs/tags/module/1.1.10-unstable",
			tag2:     "/refs/tags/module/1.1.9-unstable",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareVersions(tt.tag1, tt.tag2)
			if result != tt.expected {
				t.Errorf("compareVersions(%s, %s) = %v, want %v", tt.tag1, tt.tag2, result, tt.expected)
			}
		})
	}
}
