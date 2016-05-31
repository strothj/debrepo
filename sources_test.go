package debrepo

import (
	"reflect"
	"testing"
)

var sourceTests = []struct {
	entry  string
	source *Source
	str    string
	err    error
}{
	{
		entry: "deb http://ftp.debian.org/debian squeeze main contrib non-free",
		source: &Source{
			repoType:     "deb",
			baseURI:      "http://ftp.debian.org/debian",
			distribution: "squeeze",
			components:   []string{"main", "contrib", "non-free"},
		},
		str: "deb http://ftp.debian.org/debian squeeze main contrib non-free",
		err: nil,
	},
	{
		entry: "deb-src http://us.archive.ubuntu.com/ubuntu/ saucy universe",
		source: &Source{
			repoType:     "deb-src",
			baseURI:      "http://us.archive.ubuntu.com/ubuntu/",
			distribution: "saucy",
			components:   []string{"universe"},
		},
		str: "deb-src http://us.archive.ubuntu.com/ubuntu/ saucy universe",
		err: nil,
	},
	{
		entry:  "deb-invalid http://us.archive.ubuntu.com/ubuntu/ saucy universe",
		source: nil,
		str:    "",
		err:    InvalidSourceEntry,
	},
	{
		entry:  "deb http://us.archive.ubuntu.com/ubuntu/ saucy  ", // extra whitespace
		source: nil,
		str:    "",
		err:    InvalidSourceEntry,
	},
	{
		entry:  "deb #notURL saucy universe",
		source: nil,
		str:    "",
		err:    InvalidSourceEntry,
	},
}

func TestSource_ParseSource(t *testing.T) {
	for i, tt := range sourceTests {
		source, err := ParseSource(tt.entry)
		if expected, actual := tt.source, source; !reflect.DeepEqual(expected, actual) {
			t.Fatalf("test(%v): source: expected=%v actual=%v", i, expected, actual)
		}
		if expected, actual := tt.err, err; !reflect.DeepEqual(expected, actual) {
			t.Fatalf("test(%v): error: expected=%v actual=%v", i, expected, actual)
		}
	}
}

func TestSource_String(t *testing.T) {
	for i, tt := range sourceTests {
		var str string
		if tt.source != nil {
			str = tt.source.String()
		} else {
			str = ""
		}
		if expected, actual := tt.str, str; expected != actual {
			t.Fatalf("test(%v): expected=%v actual=%v", i, expected, actual)
		}
	}
}

func TestSource_EmptySource_StringReturnsEmptyString(t *testing.T) {
	if expected, actual := "", (Source{}).String(); expected != actual {
		t.Fatalf("expected=\"%s\" actual=\"%s\"", expected, actual)
	}
}
