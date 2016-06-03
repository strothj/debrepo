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

	// The NotAutomatic and ButAutomaticUpgrades fields are optional boolean
	// fields instructing the package manager. They may contain the values "yes"
	// and "no" (true or false). If one the fields is not specified, this has
	// the same meaning as a value of "no".
	//
	// If a value of "yes" is specified for the NotAutomatic field, a package
	// manager should not install packages (or upgrade to newer versions) from
	// this repository without explicit user consent (APT assigns priority 1 to
	// this) If the field ButAutomaticUpgrades is specified as well and has the
	// value "yes", the package manager should automatically install package
	// upgrades from this repository, if the installed version of the package is
	// higher than the version of the package in other sources (APT assigns
	// priority 100).
	//
	// Specifying "yes" for ButAutomaticUpgrades without specifying "yes" for
	// NotAutomatic is invalid.
	NotAutomatic         bool
	ButAutomaticUpgrades bool

	// An optional boolean field with the default value "no" (false). A value of
	// "yes" (true) indicates that the server supports the optional "by-hash"
	// locations as an alternative to the canonical location (and name) of an
	// index file. A client is free to choose which locations it will try to get
	// indexes from, but it is recommend to use the "by-hash" location if
	// supported by the server for its benefits for servers and clients. A
	// client may fallback to the canonical location if by-hash fails.
	AcquireByHash bool

	// An optional field containing a comma separated list of GPG key
	// fingerprints to be used for validating the next Release file.
	// The fingerprints must consist only of hex digits and may not contain
	// spaces.
	//
	// If the field is present, a client should only accept updates to the
	// repository that are signed with keys listed in the field.
	//
	// Compatibility: This feature is introduced in APT 1.3. APT
	// (as of 2016-05-01/2e49f51) requires the concrete key used to sign the
	// repository to be listed, that is, if a subkey is used, the subkey
	// fingerprint must be listed in the field.
	SignedBy [][20]byte
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
	rv.validateAutomatic()
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

func (rv *releaseValidator) validateAutomatic() {
	if !rv.NotAutomatic && rv.ButAutomaticUpgrades {
		rv.err = errors.New("can not set ButAutomaticUpgrades without NotAutomatic")
	}
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
