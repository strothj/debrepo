package debrepo

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"errors"
	"fmt"
	"reflect"
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
	Components    []string
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

// Validate validates the field values in Release.
func (r *Release) Validate() error {
	return (&releaseValidator{Release: r}).validate()
}

// ReleaseValidator validates field values in a Release.
type releaseValidator struct {
	*Release
	err error
}

// Validate returns an error if field validation fails.
func (rv *releaseValidator) validate() error {
	rv.validateComponents()
	rv.validateArchitectures()
	rv.validateNoSupportForArchitectureAll()
	rv.validateOptionalSingleLineFields()
	rv.validateOptionalSingleWordFields()
	rv.validateDate()
	rv.validateValidUntil()
	rv.validateFileSums()
	return rv.err
}

func (rv *releaseValidator) validateComponents() {
	if len(rv.Components) == 0 {
		rv.err = errors.New("field components empty")
	}
	for _, v := range rv.Components {
		if len(v) == 0 {
			rv.err = errors.New("empty component")
		}
	}
}

func (rv *releaseValidator) validateArchitectures() {
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

func (rv *releaseValidator) validateNoSupportForArchitectureAll() {
	if rv.NoSupportForArchitectureAll != "" &&
		rv.NoSupportForArchitectureAll != "Packages" {
		rv.err = errors.New("invalid value for NoSupportForArchitectureAll")
	}
}

func (rv *releaseValidator) validateOptionalSingleLineFields() {
	rv.validateSingleLineOrEmpty("Description", rv.Description)
	rv.validateSingleLineOrEmpty("Origin", rv.Origin)
	rv.validateSingleLineOrEmpty("Label", rv.Label)
}

func (rv *releaseValidator) validateOptionalSingleWordFields() {
	rv.validateSingleWordOrEmpty("Suite", rv.Suite)
	rv.validateSingleWordOrEmpty("Codename", rv.Codename)
	rv.validateSingleWordOrEmpty("Version", rv.Version)
}

func (rv *releaseValidator) validateSingleLineOrEmpty(field, str string) {
	if strings.Index(str, "\n") != -1 {
		rv.err = fmt.Errorf("field %s can not contain multiple lines", field)
	}
}

func (rv *releaseValidator) validateSingleWordOrEmpty(field, str string) {
	if strings.Index(str, "\n") != -1 ||
		strings.Index(str, " ") != -1 {
		rv.err = fmt.Errorf("field %s can contain only a single word", field)
	}
}

func (rv *releaseValidator) validateDate() {
	if rv.Date.IsZero() {
		rv.err = errors.New("field date can not be empty")
	}
}

func (rv *releaseValidator) validateValidUntil() {
	if rv.ValidUntil.IsZero() {
		return
	}
	if time.Now().After(rv.ValidUntil) {
		rv.err = errors.New("release file is expired")
	}
}

func (rv *releaseValidator) validateFileSums() {
	if len(rv.MD5Sum) == 0 &&
		len(rv.SHA1) == 0 &&
		len(rv.SHA256) == 0 {
		rv.err = errors.New("no files in release file")
		return
	}
	validateNotZeroLength := func(fileSums interface{}) {
		if fileSums == nil {
			return
		}
		keys := reflect.ValueOf(fileSums).MapKeys()
		for _, k := range keys {
			if len(k.String()) == 0 {
				rv.err = errors.New("empty filename in release file")
				return
			}
		}
	}
	validateNotZeroLength(rv.MD5Sum)
	validateNotZeroLength(rv.SHA1)
	validateNotZeroLength(rv.SHA256)
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
