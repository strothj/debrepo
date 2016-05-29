package debrepo

import (
	"fmt"
	"strings"
)

// RepoType specifies whether a repository provides binary or source code
// packages.
type RepoType int

func (r RepoType) String() string {
	if r == RepoTypeBinary {
		return "deb"
	}
	if r == RepoTypeSource {
		return "deb-src"
	}
	return "unknown repository type"
}

const (
	// RepoTypeBinary specifies a repository for binary packages.
	RepoTypeBinary RepoType = iota
	// RepoTypeSource specifies a repository for source code packages.
	RepoTypeSource
)

// A Source represents the components of a package source.
type Source struct {
	RepoType     RepoType
	URI          string
	Distribution string
	Components   []string
}

func (s *Source) String() string {
	return fmt.Sprintf("%s %s %s %s", s.RepoType, s.URI, s.Distribution, strings.Join(s.Components, " "))
}

// A SourceList contains one or more Sources.
type SourceList []*Source
