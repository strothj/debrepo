package debrepo

import (
	"testing"
)

func TestReleaseValidator_ValidateArchitectures(t *testing.T) {
	tests := []struct {
		archs []string
		valid bool
	}{
		{nil, false},
		{[]string{}, false},
		{[]string{"unsupportedArch", Architectures[0]}, false},
		{[]string{Architectures[0], Architectures[1]}, true},
	}
	for i, v := range tests {
		rv := &ReleaseValidator{Release: &Release{Architectures: v.archs}}
		rv.validateArchitectures()
		if expected, actual := v.valid, rv.err == nil; expected != actual {
			t.Fatalf("test(%v): expected=%v actual=%v", i, expected, actual)
		}
	}
}
