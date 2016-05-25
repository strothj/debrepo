package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	fileMode = os.FileMode(0644)

	directories = []string{
		"testdata/repo/root/debian/dists",
	}

	keys = []string{
		"https://ftp-master.debian.org/keys/archive-key-8.asc",
		"https://ftp-master.debian.org/keys/archive-key-8-security.asc",
	}
)

func main() {
	contents := `package debrepo

/* GENERATED FILE, DO NOT EDIT */

const testserverGenTime = "` + time.Now().UTC().Format(time.UnixDate) + `"

const testserverDistribution = "jessie"

const testserverURIRoot = "/debian"

const testserverKeyRing = ` + "`" + getKeys() + "`"

	if err := ioutil.WriteFile("testserver_gen_test.go", []byte(contents), fileMode); err != nil {
		log.Fatal(err)
	}
}

func getKeys() string {
	var keyRing []byte
	for _, keyURL := range keys {
		resp, err := http.Get(keyURL)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			log.Fatalf("Failed to retrieve key %s: %s", keyURL, resp.Status)
		}
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		keyRing = append(keyRing, b...)
	}
	return string(keyRing)
}
