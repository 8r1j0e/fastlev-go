package main

import (
	"fmt"
	"os"
	"strings"

	"fastlev-go/src/levenshtein"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: fastlev distance <string1> <string2>\n")
		fmt.Fprintf(os.Stderr, "       fastlev closest <string> <arg1,arg2,...>\n")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "distance":
		if len(os.Args) != 4 {
			fmt.Fprintln(os.Stderr, "Usage: fastlev distance <string1> <string2>")
			os.Exit(1)
		}
		fmt.Println(levenshtein.Distance(os.Args[2], os.Args[3]))

	case "closest":
		if len(os.Args) != 4 {
			fmt.Fprintln(os.Stderr, "Usage: fastlev closest <string> <arg1,arg2,...>")
			os.Exit(1)
		}
		arr := strings.Split(os.Args[3], ",")
		fmt.Println(levenshtein.Closest(os.Args[2], arr))

	case "--help":
		fmt.Fprintf(os.Stderr, "Usage: fastlev distance <string1> <string2>\n")
		fmt.Fprintf(os.Stderr, "       fastlev closest <string> <arg1,arg2,...>\n")
		os.Exit(1)

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}
