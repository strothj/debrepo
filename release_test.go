package debrepo

import (
	"testing"
	"time"
)

func TestReleaseValidator_ValidateArchitectures(t *testing.T) {
	tests := []struct {
		archs []string
		valid bool
	}{
		{nil, false},
		{[]string{}, false},
		{[]string{"unsupportedArch", architectures[0]}, false},
		{[]string{architectures[0], architectures[1]}, true},
	}
	for i, v := range tests {
		rv := &releaseValidator{Release: &Release{Architectures: v.archs}}
		rv.validateArchitectures()
		if expected, actual := v.valid, rv.err == nil; expected != actual {
			t.Fatalf("test(%v): expected=%v actual=%v", i, expected, actual)
		}
	}
}

func TestReleaseValidator_ValidateNoSupportForArchitectureAll(t *testing.T) {
	tests := []struct {
		value string
		valid bool
	}{
		{"", true},
		{"Packages", true},
		{"InvalidValue", false},
	}
	for i, v := range tests {
		rv := &releaseValidator{Release: &Release{NoSupportForArchitectureAll: v.value}}
		rv.validateNoSupportForArchitectureAll()
		if expected, actual := v.valid, rv.err == nil; expected != actual {
			t.Fatalf("test(%v): expected=%v actual=%v", i, expected, actual)
		}
	}
}

func TestReleaseValidator_ValidateOptionalSingleLineFields(t *testing.T) {
	tests := []func(r *Release) *string{
		func(r *Release) *string { return &r.Origin },
		func(r *Release) *string { return &r.Label },
	}
	for i, field := range tests {
		rv := &releaseValidator{Release: &Release{}}
		*field(rv.Release) = "single line value"
		rv.validateOptionalSingleLineFields()
		if expected, actual := false, rv.err != nil; expected != actual {
			t.Fatalf("test(%v): valid line: expected=%v actual=%v", i, expected, actual)
		}

		rv = &releaseValidator{Release: &Release{}}
		*field(rv.Release) = "first line\nsecond line"
		rv.validateOptionalSingleLineFields()
		if expected, actual := true, rv.err != nil; expected != actual {
			t.Fatalf("test(%v): multiple lines: expected=%v actual=%v", i, expected, actual)
		}
	}
}

func TestReleaseValidator_ValidateOptionalSingleWordFields(t *testing.T) {
	tests := []func(r *Release) *string{
		func(r *Release) *string { return &r.Suite },
		func(r *Release) *string { return &r.Codename },
		func(r *Release) *string { return &r.Version },
	}
	for i, field := range tests {
		rv := &releaseValidator{Release: &Release{}}
		*field(rv.Release) = "single-word-value"
		rv.validateOptionalSingleWordFields()
		if expected, actual := false, rv.err != nil; expected != actual {
			t.Fatalf("test(%v): valid word: expected=%v actual=%v", i, expected, actual)
		}

		rv = &releaseValidator{Release: &Release{}}
		*field(rv.Release) = "multiple words"
		rv.validateOptionalSingleWordFields()
		if expected, actual := true, rv.err != nil; expected != actual {
			t.Fatalf("test(%v): multiple words: expected=%v actual=%v", i, expected, actual)
		}
	}
}

func TestReleaseValidator_ValidateDate(t *testing.T) {
	rv := &releaseValidator{Release: &Release{}}
	rv.Date = time.Time{} // Empty time value
	rv.validateDate()
	if expected, actual := true, rv.err != nil; expected != actual {
		t.Fatalf("expected=%v actual=%v", expected, actual)
	}
	rv = &releaseValidator{Release: &Release{}}
	rv.Date = time.Now()
	rv.validateDate()
	if expected, actual := false, rv.err != nil; expected != actual {
		t.Fatalf("valid date: expected=%v actual=%v", expected, actual)
	}
}

func TestReleaseValidator_ValidateValidUntil(t *testing.T) {
	tests := []struct {
		time  time.Time
		valid bool
	}{
		{time.Time{}, true}, // optional field, empty time means unset
		{time.Now().Add(-30 * time.Minute), false},
		{time.Now().Add(30 * time.Minute), true},
	}
	for i, v := range tests {
		rv := &releaseValidator{Release: &Release{}}
		rv.ValidUntil = v.time
		rv.validateValidUntil()
		if expected, actual := v.valid, rv.err == nil; expected != actual {
			t.Fatalf("test(%v): expected=%v actual=%v", i, expected, actual)
		}
	}
}

func TestReleaseValidator_ValidateFileSums(t *testing.T) {
	tests := []struct {
		release *Release
		valid   bool
	}{
		{&Release{}, false},
		{&Release{MD5Sum: map[string]MD5FileMetaData{}}, false},
		{&Release{SHA1: map[string]SHA1FileMetaData{}}, false},
		{&Release{SHA256: map[string]SHA256FileMetaData{}}, false},
		{&Release{MD5Sum: map[string]MD5FileMetaData{"": MD5FileMetaData{}}}, false},
		{&Release{SHA1: map[string]SHA1FileMetaData{"": SHA1FileMetaData{}}}, false},
		{&Release{SHA256: map[string]SHA256FileMetaData{"": SHA256FileMetaData{}}}, false},
		{&Release{MD5Sum: map[string]MD5FileMetaData{"main/binary-all/Packages": MD5FileMetaData{}}}, true},
	}
	for i, v := range tests {
		rv := &releaseValidator{Release: v.release}
		rv.validateFileSums()
		if expected, actual := v.valid, rv.err == nil; expected != actual {
			t.Fatalf("test(%v): expected=%v actual=%v", i, expected, actual)
		}
	}
}
