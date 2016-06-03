package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var (
	fileMode = os.FileMode(0644)

	directories = []string{
		"testdata/repo/root/debian/dists/jessie/main/binary-amd64",
	}

	files = []struct {
		url  string
		path string
	}{
		{"http://ftp.debian.org/debian/dists/jessie/Release", "testdata/repo/root/debian/dists/jessie/Release"},
		{"http://ftp.debian.org/debian/dists/jessie/Release.gpg", "testdata/repo/root/debian/dists/jessie/Release.gpg"},
		{"http://ftp.debian.org/debian/dists/jessie/main/binary-amd64/Packages.gz", "testdata/repo/root/debian/dists/jessie/main/binary-amd64/Packages.gz"},
		{"http://ftp.debian.org/debian/dists/jessie/main/binary-amd64/Packages.xz", "testdata/repo/root/debian/dists/jessie/main/binary-amd64/Packages.xz"},
	}
)

func main() {
	removeOldDirectories()
	createDirectories()
	downloadFiles()
}

func removeOldDirectories() {
	if err := os.RemoveAll("testdata/repo"); err != nil {
		log.Print(err)
	}
}

func createDirectories() {
	for _, dir := range directories {
		if err := os.MkdirAll(dir, fileMode); err != nil {
			log.Fatal(err)
		}
	}
}

func downloadFiles() {
	for _, file := range files {
		resp, err := http.Get(file.url)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			log.Fatalf("Error downloading file %s: %s", file.url, resp.Status)
		}
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("Error reading data for file %s: %v", file.url, err)
		}
		if err = ioutil.WriteFile(file.path, b, fileMode); err != nil {
			log.Fatalf("Error writing file %s: %v", file.path, err)
		}
	}
}
