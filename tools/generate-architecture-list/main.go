package main

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"
)

func main() {
	out, err := exec.Command("dpkg-architecture", "-L").Output()
	if err != nil {
		panic(err)
	}
	outS := string(out)
	archs := strings.Split(outS, "\n")
	output := `package debrepo
    
//go:generate -command generate-archs go run tools/generate-architecture-list/main.go
//go:generate generate-archs

/* GENERATED FILE, DO NOT EDIT */

// Architectures is a list of supported repository architecture types.
// See https://www.debian.org/doc/debian-policy/ch-customized-programs.html
var Architectures = [...]string{
`
	for _, a := range archs {
        if len(a) > 0 {
		    output += fmt.Sprintf("    \"%s\",\n", a)
        }
	}
	output += "}"
	if err := ioutil.WriteFile("architectures.go", []byte(output), 0644); err != nil {
		panic(err)
	}

}
