package debrepo

import (
	"fmt"
	"strings"

	"github.com/asaskevich/govalidator"
)

const (
	// InvalidSourceEntry is returned on malformed Source entries.
	InvalidSourceEntry = Error("unable to parse source")
)

// A Source is an entry in the package resource list. It represents a line in
// the control file "sources.list".
type Source struct {
	repoType     string
	baseURI      string
	distribution string
	components   []string
}

func (s Source) String() string {
	if len(s.components) == 0 {
		return ""
	}
	return fmt.Sprintf("%s %s %s %s",
		s.repoType,
		s.baseURI,
		s.distribution,
		strings.Join(s.components, " "))
}

// ParseSource parses entry to create a Source.
// entry must be in the format:
// 	deb http://ftp.debian.org/debian squeeze main contrib non-free
func ParseSource(entry string) (*Source, error) {
	ss := strings.Split(entry, " ")
	if len(ss) < 4 {
		return nil, InvalidSourceEntry
	}
	for _, f := range ss {
		if len(f) == 0 {
			return nil, InvalidSourceEntry
		}
	}
	if ss[0] != "deb" && ss[0] != "deb-src" {
		return nil, InvalidSourceEntry
	}
	if !govalidator.IsURL(ss[1]) {
		return nil, InvalidSourceEntry
	}
	return &Source{
		repoType:     ss[0],
		baseURI:      ss[1],
		distribution: ss[2],
		components:   ss[3:],
	}, nil
}

// SourceList is a list of APT data sources. It is equivalent to the file
// "sources.list" on Debian style Linux distributions.
type SourceList []*Source
