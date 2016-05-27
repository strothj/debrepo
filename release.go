package debrepo

import (
	"errors"
	"fmt"
)

// Release contains meta-information about the distribution and the checksums
// for the indices.
// It is decoded from the Release or InRelease file present at
// "dists/$DIST/InRelease" or "dists/$DIST/Release".
// See https://wiki.debian.org/RepositoryFormat#A.22Release.22_files
type Release struct {
	// These fields are optional.
	Description                 string
	Origin                      string
	Label                       string
	Version                     string
	Suite                       string
	Codename                    string
	NoSupportForArchitectureAll string

	// These fields determine the layout of the repository and should contain
	// something meaningful to the user.
	Components    string
	Architectures []string

	// These fields are purely functional and used mostly internally by
	// packaging tools.
	Date       string
	ValidUntil string
	MD5Sum     string
	SHA1       string
	SHA256     string
}

// ReleaseValidator validates field values in a Release.
type ReleaseValidator struct {
	*Release
	err error
}

func (rv *ReleaseValidator) validateArchitectures() {
	if rv.Architectures == nil || len(rv.Architectures) == 0 {
		rv.err = errors.New("field Architectures empty")
		return
	}
	for _, v := range rv.Architectures {
		valid := false
		for i := 0; i < len(Architectures); i++ {
			if v == Architectures[i] {
				valid = true
				break
			}
		}
		if !valid {
			rv.err = fmt.Errorf("unsupported architecture: %s", v)
			return
		}
	}
	return
}
