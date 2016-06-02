package debrepo

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"
	"time"
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
	Date time.Time
	// ValidUntil is an optional field which specifies at which time the Release
	// file should be considered expired by the client. Client behaviour on
	// expired Release files is unspecified. An empty Time means no expiration
	// time was set.
	ValidUntil time.Time
	MD5Sum     map[string]MD5FileMetaData
	SHA1       map[string]SHA1FileMetaData
	SHA256     map[string]SHA256FileMetaData
}

// ReleaseValidator validates field values in a Release.
type ReleaseValidator struct {
	*Release
	err error
}

// Validate returns an error if field validation fails.
func (rv *ReleaseValidator) Validate() error {
	rv.validateArchitectures()
	rv.validateNoSupportForArchitectureAll()
	rv.validateOptionalSingleLineFields()
	rv.validateOptionalSingleWordFields()
	rv.validateDate()
	rv.validateValidUntil()
	return rv.err
}

func (rv *ReleaseValidator) validateArchitectures() {
	if rv.Architectures == nil || len(rv.Architectures) == 0 {
		rv.err = errors.New("field Architectures empty")
		return
	}
	for _, v := range rv.Architectures {
		valid := false
		for i := 0; i < len(architectures); i++ {
			if v == architectures[i] {
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

func (rv *ReleaseValidator) validateNoSupportForArchitectureAll() {
	if rv.NoSupportForArchitectureAll != "" &&
		rv.NoSupportForArchitectureAll != "Packages" {
		rv.err = errors.New("invalid value for NoSupportForArchitectureAll")
	}
}

func (rv *ReleaseValidator) validateOptionalSingleLineFields() {
	rv.validateSingleLineOrEmpty("Origin", rv.Origin)
	rv.validateSingleLineOrEmpty("Label", rv.Label)
}

func (rv *ReleaseValidator) validateOptionalSingleWordFields() {
	rv.validateSingleWordOrEmpty("Suite", rv.Suite)
	rv.validateSingleWordOrEmpty("Codename", rv.Codename)
	rv.validateSingleWordOrEmpty("Version", rv.Version)
}

func (rv *ReleaseValidator) validateSingleLineOrEmpty(field, str string) {
	if strings.Index(str, "\n") != -1 {
		rv.err = fmt.Errorf("field %s can not contain multiple lines", field)
	}
}

func (rv *ReleaseValidator) validateSingleWordOrEmpty(field, str string) {
	if strings.Index(str, "\n") != -1 ||
		strings.Index(str, " ") != -1 {
		rv.err = fmt.Errorf("field %s can contain only a single word", field)
	}
}

func (rv *ReleaseValidator) validateDate() {
	if rv.Date.IsZero() {
		rv.err = errors.New("field date can not be empty")
	}
}

func (rv *ReleaseValidator) validateValidUntil() {
	if rv.ValidUntil.IsZero() {
		return
	}
	if time.Now().After(rv.ValidUntil) {
		rv.err = errors.New("release file is expired")
	}
	return
}

// MD5FileMetaData stores the MD5 sum and file length of a file in a repository
// Release file.
type MD5FileMetaData struct {
	Length int64
	Sum    [md5.Size]byte
}

// SHA1FileMetaData stores the SHA1 sum and file length of a file in a
// repository Release file.
type SHA1FileMetaData struct {
	Length int64
	Sum    [sha1.Size]byte
}

// SHA256FileMetaData stores the SHA256 sum and file length of a file in a
// repository Release file.
type SHA256FileMetaData struct {
	Length int64
	Sum    [sha256.Size]byte
}
