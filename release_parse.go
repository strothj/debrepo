package debrepo

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"text/template"
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
		MD5Sum:   make(map[string]MD5FileMetaData),
		SHA1:     make(map[string]SHA1FileMetaData),
		SHA256:   make(map[string]SHA256FileMetaData),
		SignedBy: make([][20]byte, 0),
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
		case "NotAutomatic:":
			release.NotAutomatic = parseOptionalBool(words[1], "NotAutomatic")
		case "ButAutomaticUpgrades:":
			release.ButAutomaticUpgrades = parseOptionalBool(words[1], "ButAutomaticUpgrades")
		case "Acquire-By-Hash:":
			release.AcquireByHash = parseOptionalBool(words[1], "Acquire-By-Hash")
		case "Signed-By:":
			fingerprintLine := strings.Join(words[1:], "")
			fingerprints := strings.Split(fingerprintLine, ",")
			for _, f := range fingerprints {
				f = strings.TrimSpace(f)
				b, err := hex.DecodeString(f)
				if err != nil {
					return nil, err
				}
				if len(b) != 20 {
					return nil, errors.New("invalid fingerprint in Signed-By")
				}
				var bb [20]byte
				copy(bb[:], b)
				release.SignedBy = append(release.SignedBy, bb)
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

// Serialize saves a Release to a file.
func (r *Release) Serialize(out io.Writer) error {
	buf := &bytes.Buffer{}
	err := releaseTemplate.Execute(buf, r)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(buf)
	w := bufio.NewWriter(out)
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), " ")
		if len(line) == 0 {
			continue
		}
		if _, err := w.WriteString(fmt.Sprintf("%s\n", line)); err != nil {
			return err
		}
	}
	if err := w.Flush(); err != nil {
		return err
	}
	return scanner.Err()
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

func parseOptionalBool(word, field string) bool {
	if word == "yes" {
		return true
	} else if word == "no" {
		return false
	}
	panic(fmt.Errorf("invalid value for %s", field))
}

const releaseTemplateStr = `{{with .Origin}}Origin: {{.}}{{end}}
{{with .Label}}Label: {{.}}{{end}}
{{with .Suite}}Suite: {{.}}{{end}}
{{with .Version}}Version: {{.}}{{end}}
{{with .Codename}}Codename: {{.}}{{end}}
{{with .Date}}Date: {{date .}}{{end}}
{{with .Architectures}}Architectures: {{range .}}{{.}} {{end}}{{end}}
{{with .Components}}Components: {{range .}}{{.}} {{end}}{{end}}
{{with .Description}}Description: {{.}}{{end}}
{{with .NoSupportForArchitectureAll}}No-Support-for-Architecture-all: {{.}}{{end}}
{{validUntil .ValidUntil}}
{{with .NotAutomatic}}NotAutomatic: {{.}}{{end}}
{{with .ButAutomaticUpgrades}}ButAutomaticUpgrades: {{.}}{{end}}
{{with .AcquireByHash}}Acquire-By-Hash: {{.}}{{end}}
{{signedBy .SignedBy}}
{{with .MD5Sum -}}
MD5Sum:
{{range $key, $value := .}} {{printf "%s %8d %s" (hex16 .Sum) .Length $key}}
{{end}}
{{- end -}}
{{with .SHA1 -}}
SHA1:
{{range $key, $value := .}} {{printf "%s %8d %s" (hex20 .Sum) .Length $key}}
{{end}}
{{- end -}}
{{with .SHA256 -}}
SHA256:
{{range $key, $value := .}} {{printf "%s %8d %s" (hex32 .Sum) .Length $key}}
{{end}}
{{- end -}}
`

var releaseTemplateFuncs = template.FuncMap{
	"date": func(t time.Time) string {
		if t.IsZero() {
			return time.Now().Format(time.RFC1123)
		}
		return t.Format(time.RFC1123)
	},
	"validUntil": func(t time.Time) string {
		if t.IsZero() {
			return ""
		}
		return "Valid-Until: " + t.Format(time.RFC1123)
	},
	"signedBy": func(signers [][20]byte) string {
		if len(signers) == 0 {
			return ""
		}
		strSigners := make([]string, 0, len(signers))
		for _, v := range signers {
			strSigners = append(strSigners, hex.EncodeToString(v[:]))
		}
		return "Signed-By: " + strings.Join(strSigners, ",")
	},
	"hex16": func(b [md5.Size]byte) string {
		bb := make([]byte, 16)
		copy(bb, b[:])
		return hex.EncodeToString(bb)
	},
	"hex20": func(b [sha1.Size]byte) string {
		bb := make([]byte, sha1.Size)
		copy(bb, b[:])
		return hex.EncodeToString(bb)
	},
	"hex32": func(b [sha256.Size]byte) string {
		bb := make([]byte, sha256.Size)
		copy(bb, b[:])
		return hex.EncodeToString(bb)
	},
}

var releaseTemplate = template.Must(template.New("").Funcs(releaseTemplateFuncs).Parse(releaseTemplateStr))
