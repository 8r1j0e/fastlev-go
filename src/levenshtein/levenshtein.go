package levenshtein

// Distance returns the Levenshtein edit distance between strings a and b.
// Uses Myers' bit-parallel algorithm with automatic dispatch:
//   - 64-bit single-pass (myers64) for strings up to maxSinglePass=64 on 64-bit platforms
//   - 32-bit single-pass (myers32) for strings up to maxSinglePass=32 on 32-bit platforms
//   - Block-based extended algorithm for longer strings (myersX64 or myersX32)
//
// The specific implementation is selected at compile time via build tags:
//   - amd64, arm64, riscv64, etc. → 64-bit paths preferred
//   - 386, arm (32-bit)           → 32-bit paths preferred
//

func Distance(a, b string) int {
	if len(a) < len(b) {
		a, b = b, a
	}

	if len(b) == 0 {
		return len(a)
	}

	if is64bit {
		if len(a) <= 64 {
			return myers64(a, b)
		}
		return myersX64(a, b)
	}

	if len(a) <= 32 {
		return myers32(a, b)
	}
	return myersX32(a, b)
}

// Closest returns the string from arr with the lowest Levenshtein edit distance
// to str. Mirrors the original's linear-scan approach.
// Returns "" if arr is empty.
func Closest(str string, arr []string) string {
	if len(arr) == 0 {
		return ""
	}

	minIdx := 0
	minDist := Distance(str, arr[0])

	for i := 1; i < len(arr); i++ {
		dist := Distance(str, arr[i])
		if dist < minDist {
			minDist = dist
			minIdx = i
		}
	}

	return arr[minIdx]
}
