package debrepo

import "fmt"

func ExampleSource() {
	s := &Source{
		RepoType:     RepoTypeBinary,
		URI:          "http://ftp.debian.org/debian",
		Distribution: "squeeze",
		Components:   []string{"main", "contrib", "non-free"},
	}
	fmt.Println(s)
	// Output:
	// deb http://ftp.debian.org/debian squeeze main contrib non-free
}
