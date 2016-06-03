package debrepo

import (
	"bytes"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func TestRelease_ReadRelease(t *testing.T) {
	checkField := func(field, expected, actual string) {
		if expected != actual {
			t.Fatalf("%s: expected=%v actual=%v", field, expected, actual)
		}
	}
	f, err := os.Open("testdata/repo/root/debian/dists/jessie/Release")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
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

func TestRelease_Serialize(t *testing.T) {
	expected, err := ioutil.ReadFile("testdata/repo/root/debian/dists/jessie/Release")
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.Open("testdata/repo/root/debian/dists/jessie/Release")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	r, err := ReadRelease(f)
	if err != nil {
		t.Fatal(err)
	}
	var actual = &bytes.Buffer{}
	if err := r.Serialize(actual); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expected, actual.Bytes()) {
		t.Fatal("expected != actual")
	}
}
