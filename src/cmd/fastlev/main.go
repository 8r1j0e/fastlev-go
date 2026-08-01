package main

import (
	"fmt"

	levenshtien "fastlev-go/src/levenshtein"
)

func main() {
	a := "abcdefghijklmnopqrstuvwxyz1234567890"
	b := "abcdefghijklsfrtqrstuvwxyz1234567891"

	fmt.Println(levenshtien.Distance(a, b))
}
