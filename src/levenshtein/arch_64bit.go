//go:build amd64 || arm64 || riscv64 || loong64 || mips64 || mips64le || ppc64 || ppc64le || s390x || wasm
// +build amd64 arm64 riscv64 loong64 mips64 mips64le ppc64 ppc64le s390x wasm

package levenshtein

import "runtime"

// maxSinglePass is the maximum string length handled by the single-pass Myers
// algorithm. On 64-bit platforms we can handle up to 64 characters in one 64-bit
// integer. Strings longer than this fall through to the block-based path.
const maxSinglePass = 64

// is64bit returns true on 64-bit platforms.
const is64bit = true

func init() {
	// Dispatch chosen algorithm once at startup.
	// On 64-bit platforms, we prefer myers64 for strings up to 64 chars.
	// myersX64 handles strings longer than that with 64-bit blocks.
	// Both 32-bit paths (myers32, myersX32) remain available for legacy use
	// but are not the default dispatch targets on this platform.
	_ = runtime.GOARCH // force compile-time link
}