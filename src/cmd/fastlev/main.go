package main

import (
	"fmt"

	levenshtien "fastlev-go/src/levenshtein"
)

func main() {
	fmt.Println(levenshtien.Distance("kitten", "sitting"))
}
