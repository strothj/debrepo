package debrepo

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/openpgp"
)

//go:generate -command generate-test-keyring go run tools/generate-test-keyring/main.go
//go:generate -command generate-test-repo-contents go run tools/generate-test-repo-contents/main.go
//go:generate generate-test-keyring
//go:generate generate-test-repo-contents

type testserver struct {
	*httptest.Server
}

func newTestServer() *testserver {
	return &testserver{
		Server: httptest.NewServer(http.FileServer(http.Dir("testdata/repo/root"))),
	}
}

func (ts *testserver) Time() time.Time {
	t, err := time.Parse(time.UnixDate, testserverGenTime)
	if err != nil {
		panic(err)
	}
	return t
}

func (ts *testserver) Distribution() string { return testserverDistribution }

func (ts *testserver) URIRoot() string { return testserverURIRoot }

func (ts *testserver) KeyRing() openpgp.EntityList {
	r := strings.NewReader(testserverKeyRing)
	el, err := openpgp.ReadArmoredKeyRing(r)
	if err != nil {
		panic(err)
	}
	return el
}

func Test_TestServer_Time(t *testing.T) {
	ts := &testserver{}
	if expected, actual := testserverGenTime, ts.Time().Format(time.UnixDate); expected != actual {
		t.Fatalf("expected=%v actual=%v", expected, actual)
	}
}

func Test_TestServer_KeyRing(t *testing.T) {
	ts := &testserver{}
	el := ts.KeyRing()
	if len(el) < 1 {
		t.Fatalf("expected=>0 actual=%v", len(el))
	}
}

func Test_TestServer_FileServer(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()
	u, _ := url.Parse(ts.URL)
	u.Path = path.Join(ts.URIRoot(), "dists", ts.Distribution(), "Release")
	resp, err := http.Get(u.String())
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		t.Fatal("expected status OK")
	}
}
