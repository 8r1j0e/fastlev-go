//go:build !amd64 && !arm64 && !riscv64 && !loong64 && !mips64 && !mips64le && !ppc64 && !ppc64le && !s390x && !wasm
// +build !amd64,!arm64,!riscv64,!loong64,!mips64,!mips64le,!ppc64,!ppc64le,!s390x,!wasm

package levenshtein

import "runtime"

// maxSinglePass is the maximum string length handled by the single-pass Myers
// algorithm. On 32-bit platforms we can handle up to 32 characters in one 32-bit
// integer. Strings longer than this fall through to the block-based path.
const maxSinglePass = 32

// is64bit returns false on 32-bit platforms.
const is64bit = false

func init() {
	// On 32-bit platforms, we use myers32 for strings up to 32 chars
	// and myersX32 for longer strings (32-bit block decomposition).
	// The 64-bit paths are available if needed but not the default.
	_ = runtime.GOARCH
}