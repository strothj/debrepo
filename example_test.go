package debrepo_test

import (
	"fmt"

	"github.com/strothj/debrepo"
)

func ExampleParseSource() {
	repoLine := "deb http://ftp.debian.org/debian squeeze main contrib non-free"
	source, err := debrepo.ParseSource(repoLine)
	if err != nil {
		// Error parsing repository line
	}
	fmt.Println(source)
	// Output: deb http://ftp.debian.org/debian squeeze main contrib non-free
}
