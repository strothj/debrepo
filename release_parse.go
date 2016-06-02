package debrepo

import (
	"bufio"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

// ReadRelease returns a Release from a Release file.
func ReadRelease(r io.Reader) (release *Release, err error) {
	defer func() {
		if p := recover(); p != nil {
			release = nil
			err = fmt.Errorf("parsing error: %s", p)
		}
	}()
	scanner := bufio.NewScanner(r)
	release = &Release{
		MD5Sum: make(map[string]MD5FileMetaData),
		SHA1:   make(map[string]SHA1FileMetaData),
		SHA256: make(map[string]SHA256FileMetaData),
	}

	var fileSum string
	for scanner.Scan() {
		line := scanner.Text()
		words := strings.Split(line, " ")
		if len(words) == 0 {
			continue
		}
		switch words[0] {
		case "Description:":
			release.Description = strings.Join(words[1:], " ")
		case "Origin:":
			release.Origin = strings.Join(words[1:], " ")
		case "Label:":
			release.Label = strings.Join(words[1:], " ")
		case "Version:":
			release.Version = words[1]
		case "Suite:":
			release.Suite = words[1]
		case "Codename:":
			release.Codename = words[1]
		case "No-Support-for-Architecture-all:":
			release.NoSupportForArchitectureAll = words[1]
		case "Components:":
			release.Components = words[1:]
		case "Architectures:":
			release.Architectures = words[1:]
		case "Date:":
			release.Date = parseDate(strings.Join(words[1:], " "))
		case "Valid-Until:":
			release.ValidUntil = parseDate(strings.Join(words[1:], " "))
		case "MD5Sum:":
			fileSum = "MD5Sum"
		case "SHA1:":
			fileSum = "SHA1"
		case "SHA256:":
			fileSum = "SHA256"
		case "":
			sum, length, path := parseFileSumParams(words)
			b, err := hex.DecodeString(sum)
			if err != nil {
				panic(err)
			}
			switch fileSum {
			case "MD5Sum":
				var bb [md5.Size]byte
				copy(bb[:], b)
				release.MD5Sum[path] = MD5FileMetaData{Length: length, Sum: bb}
			case "SHA1":
				var bb [sha1.Size]byte
				copy(bb[:], b)
				release.SHA1[path] = SHA1FileMetaData{Length: length, Sum: bb}
			case "SHA256":
				var bb [sha256.Size]byte
				copy(bb[:], b)
				release.SHA256[path] = SHA256FileMetaData{Length: length, Sum: bb}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if err := release.Validate(); err != nil {
		return nil, err
	}
	return release, nil
}

func parseFileSumParams(words []string) (sum string, length int64, path string) {
	var parsed int
	for _, w := range words {
		if len(w) > 0 {
			switch parsed {
			case 0:
				sum = w
			case 1:
				i, err := strconv.ParseInt(w, 10, 64)
				if err != nil {
					panic(err)
				}
				length = i
			case 2:
				path = w
			}
			parsed++
		}
	}
	return
}

func parseDate(value string) time.Time {
	date, err := time.Parse(time.RFC1123, value)
	if err != nil {
		date, err = time.Parse(time.RFC1123Z, value)
		if err != nil {
			panic(err)
		}
	}
	return date
}