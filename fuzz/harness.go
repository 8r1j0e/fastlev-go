package fuzz

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"fastlev-go/src/levenshtein"
)

// Differential fuzz = run BOTH the original Node lib AND the Go port
// on identical random inputs and compare results.
// Supports both 32-bit and 64-bit Myers algorithm families.

func levenshteinRef(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}
	if len(a) > len(b) {
		a, b = b, a
	}
	row := make([]int, len(a)+1)
	for i := range row {
		row[i] = i
	}
	for i := 1; i <= len(b); i++ {
		prev := i
		for j := 1; j <= len(a); j++ {
			val := row[j-1]
			if b[i-1] != a[j-1] {
				val = min(row[j-1]+1, prev+1, row[j]+1)
			}
			row[j-1] = prev
			prev = val
		}
		row[len(a)] = prev
	}
	return row[len(a)]
}

// findRepoRoot walks up from the working directory looking for go.mod
func findRepoRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for i := 0; i < 10; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return filepath.Dir(wd)
}

// jsDistance calls the original JavaScript via Node.
func jsDistance(a, b string) (int, error) {
	script := fmt.Sprintf(
		`const {distance:dist}=require("fastest-levenshtein");process.stdout.write(String(dist(%q,%q)));`,
		a, b,
	)
	cmd := exec.Command("node", "-e", script)
	repoRoot := findRepoRoot()
	if repoRoot != "" {
		cmd.Dir = repoRoot
	}
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return 0, fmt.Errorf("node error (%v): %s", err, out.String())
	}
	var dist int
	if _, err := fmt.Sscanf(out.String(), "%d", &dist); err != nil {
		return 0, fmt.Errorf("parse error: %s", out.String())
	}
	return dist, nil
}

// RunFuzz runs N iterations of random string pair tests, comparing
// Go port vs the original JS library (or reference DP as fallback).
// This exercises all algorithm paths: myers32/myers64 for short strings
// and myersX32/myersX64 for long strings.
func RunFuzz(iterations int, rng *rand.Rand) []string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	var divergences []string

	for i := 0; i < iterations; i++ {
		n := rng.Intn(500) + 1
		m := rng.Intn(500) + 1

		s1 := randomString(rng, charset, n)
		s2 := randomString(rng, charset, m)

		goDist := levenshtein.Distance(s1, s2)
		jsDist, err := jsDistance(s1, s2)
		if err != nil {
			// Node not available - fall back to DP reference
			refDist := levenshteinRef(s1, s2)
			if goDist != refDist {
				divergences = append(divergences,
					fmt.Sprintf("DIVERGENCE [%d]: Go=%d Ref=%d | len(%d,%d)", i, goDist, refDist, n, m))
			}
		} else {
			if goDist != jsDist {
				divergences = append(divergences,
					fmt.Sprintf("DIVERGENCE [%d]: Go=%d JS=%d | len(%d,%d)", i, goDist, jsDist, n, m))
			}
		}
	}
	return divergences
}

func randomString(rng *rand.Rand, charset string, length int) string {
	var sb strings.Builder
	sb.Grow(length)
	for i := 0; i < length; i++ {
		sb.WriteByte(charset[rng.Intn(len(charset))])
	}
	return sb.String()
}