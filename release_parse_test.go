package debrepo

import (
	"os"
	"testing"
)

func TestReadRelease(t *testing.T) {
	checkField := func(field, expected, actual string) {
		if expected != actual {
			t.Fatalf("%s: expected=%v actual=%v", field, expected, actual)
		}
	}
	f, err := os.Open("testdata/repo/root/debian/dists/jessie/Release")
	if err != nil {
		t.Fatal(err)
	}
	r, err := ReadRelease(f)
	if err != nil {
		t.Fatal(err)
	}
	if r == nil {
		t.Fatal("expected non-nil Release")
	}
	checkField("Origin", "Debian", r.Origin)
	checkField("Label", "Debian", r.Label)
	checkField("Suite", "stable", r.Suite)
	checkField("Codename", "jessie", r.Codename)
	if len(r.Version) == 0 {
		t.Fatal("version empty")
	}
	if len(r.MD5Sum) < 500 || len(r.SHA1) < 500 || len(r.SHA256) < 500 {
		t.Fatal("missing file sums")
	}
}
