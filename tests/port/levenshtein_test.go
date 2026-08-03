package levenshtein_test

import (
	"math/rand"
	"testing"

	lev "fastlev-go/src/levenshtein"
)

// ---- Reference implementation (naive DP for verification) ----

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

// ---- Table-driven tests (32-bit and 64-bit boundaries) ----

func TestDistance_KnownValues(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"", "", 0},
		{"a", "", 1},
		{"", "a", 1},
		{"a", "a", 0},
		{"a", "b", 1},
		{"ab", "ac", 1},
		{"fast", "faster", 2},
		{"faster", "fast", 2},
		{"kitten", "sitting", 3},
		{"saturday", "sunday", 3},
		{"", "hello", 5},
		{"hello", "", 5},
		{"same", "same", 0},
		{"abc", "xyz", 3},
		{"aaaa", "aaaa", 0},
		{"abcd", "dcba", 4},
		// 32-char case: exercises myers32 boundary on 32-bit, myers64 on 64-bit
		{"abcdefghijklmnopqrstuvwxyzABCDEF", "bacdefghijklmnopqrstuvwxyzABCDEF", 2},
		// 64-char case: exercises myers64 boundary on 64-bit, myersX32 on 32-bit
		{"abcdefghijklmnopqrstuvwxyzABCDEFabcdefghijklmnopqrstuvwxyzABCDEF",
			"bacdefghijklmnopqrstuvwxyzABCDEFbacdefghijklmnopqrstuvwxyzABCDEF", 4},
		// 65-char case: exercises myersX64 on 64-bit
		{"abcdefghijklmnopqrstuvwxyzABCDEFabcdefghijklmnopqrstuvwxyzABCDEFF",
			"bacdefghijklmnopqrstuvwxyzABCDEFbacdefghijklmnopqrstuvwxyzABCDEFF", 4},
	}
	for _, tt := range tests {
		got := lev.Distance(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("Distance(%q, %q) = %d; want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestClosest_Basic(t *testing.T) {
	got := lev.Closest("fast", []string{"slow", "faster", "fastest"})
	want := "faster"
	if got != want {
		t.Errorf("Closest(%q, [slow, faster, fastest]) = %q; want %q", "fast", got, want)
	}
}

func TestClosest_EmptySlice(t *testing.T) {
	got := lev.Closest("anything", []string{})
	want := ""
	if got != want {
		t.Errorf("Closest with empty slice = %q; want %q", got, want)
	}
}

func TestClosest_ExactMatch(t *testing.T) {
	got := lev.Closest("exact", []string{"not", "exact", "close"})
	want := "exact"
	if got != want {
		t.Errorf("Closest with exact match = %q; want %q", got, want)
	}
}

// ---- Property-based random tests (1000 pairs against reference) ----

func TestDistance_RandomVsReference(t *testing.T) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	rng := rand.New(rand.NewSource(42))

	for i := 0; i < 1000; i++ {
		n := rng.Intn(1000) + 1
		m := rng.Intn(1000) + 1

		s1 := randomString(rng, charset, n)
		s2 := randomString(rng, charset, m)

		got := lev.Distance(s1, s2)
		want := levenshteinRef(s1, s2)

		if got != want {
			t.Errorf("Distance mismatch: len(%d,%d) got=%d want=%d", n, m, got, want)
			t.Errorf("  s1=%q... s2=%q...", s1[:min(20, len(s1))], s2[:min(20, len(s2))])
			return
		}
	}
}

// ---- Boundary tests (at all algorithm transition points) ----

func TestDistance_Boundaries(t *testing.T) {
	charset := "ab"

	// Single character
	if d := lev.Distance("a", "b"); d != 1 {
		t.Errorf("single char diff: got %d", d)
	}
	if d := lev.Distance("a", "a"); d != 0 {
		t.Errorf("single char same: got %d", d)
	}

	// 32-char boundary (myers32 ↔ myersX32 on 32-bit; still single-pass myers64 on 64-bit)
	for _, n := range []int{31, 32, 33} {
		s1 := randomString(rand.New(rand.NewSource(42)), charset, n)
		s2 := randomString(rand.New(rand.NewSource(43)), charset, n)
		d := lev.Distance(s1, s2)
		if d < 0 {
			t.Errorf("negative distance at n=%d: %d", n, d)
		}
	}

	// 64-char boundary (myers64 ↔ myersX64 on 64-bit)
	for _, n := range []int{63, 64, 65, 100, 200} {
		s1 := randomString(rand.New(rand.NewSource(50)), charset, n)
		s2 := randomString(rand.New(rand.NewSource(51)), charset, n)
		got := lev.Distance(s1, s2)
		want := levenshteinRef(s1, s2)
		if got != want {
			t.Errorf("boundary n=%d: got %d, want %d", n, got, want)
		}
	}

	// Long strings should still pass
	s200 := randomString(rand.New(rand.NewSource(5)), "abc", 200)
	t200 := randomString(rand.New(rand.NewSource(6)), "cba", 150)
	if d := lev.Distance(s200, t200); d != levenshteinRef(s200, t200) {
		t.Errorf("long strings mismatch: got %d, want %d", d, levenshteinRef(s200, t200))
	}
}

// ---- Helpers ----

func randomString(rng *rand.Rand, charset string, length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rng.Intn(len(charset))]
	}
	return string(b)
}