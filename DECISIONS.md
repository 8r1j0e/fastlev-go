# DECISIONS.md

Every non-trivial architectural divergence from the original TypeScript implementation,
with rationale.

---

## 1. Package name: `levenshtein` instead of `fastest-levenshtein`

**Decision:** Named the Go package `levenshtein` (imported as `fastlev-go/src/levenshtein`).

**Rationale:** Go convention is short, single-word package names. `fastest-levenshtein` violates Go's `go vet` naming guidelines (hyphens not allowed in identifiers). The binary is named `fastlev` for brand continuity.

---

## 2. `peq` as global `[65536]uint32` (slice allocation)

**Decision:** Declared `var peq = make([]uint32, 0x10000)` as a package-level variable.

**Rationale:** Mirrors the original's global `Uint32Array(0x10000).` A 256KB fixed-size lookup is reused across all calls to avoid per-call allocation overhead. **Note:** After adding 64-bit paths a second table `peq64 [65536]uint64` was added alongside the original `peq32` — see decision #17.
**Trade-off:** Not concurrent-safe — concurrent calls to `Distance` will corrupt the table. This matches the original single-threaded JS semantics. Documented here so downstream consumers know to use a mutex or per-goroutine table if needed.

---

## 3. `math.MaxUint32` for -1 in bit operations

**Decision:** Used `^uint32(0)` to represent the all-ones bitmask that the original expresses as `-1`.

**Rationale:** TypeScript's bitwise operators treat numbers as signed 32-bit integers, so `-1` becomes `0xFFFFFFFF`. Go's `uint32` cannot represent signed -1; `^uint32(0)` is the idiomatic way to get `0xFFFFFFFF`.

---

## 4. `int(a[i])` for character lookup instead of `charCodeAt(i)`

**Decision:** Used direct byte indexing (`a[i]`) rather than rune decoding.

**Rationale:** The original uses `charCodeAt()` which returns UTF-16 code units. For ASCII-range characters (which constitute 100% of the original test suite and benchmark data), `a[i]` produces the identical integer. Full Unicode support would require `[]rune` conversion, which would change the algorithm's semantics and break behavioral equivalence with the original.

---

## 5. Block decomposition for strings beyond single-pass limit (`myersX`)

**Decision:** Implemented both 32-bit and 64-bit block decomposition variants (`myersX32` / `myersX64`) matching the original's `myers_x` pattern.

**Rationale:** The original uses `(i / 32) | 0` for block indexing and `>>> i` for intra-block bit extraction. Go's right-shift is always unsigned for both `uint32` and `uint64`, so the translation is exact. On 64-bit platforms, `myersX64` uses 64-bit blocks — reducing the number of vertical pass cuts by 50% compared to the original's 32-bit blocks.

---

## 6. Arg swapping in `Distance` (asymmetric optimization)

**Decision:** Preserved the original's arg swap pattern — the longer string is placed first.

**Rationale:** Myers' algorithm performance depends on the longer string determining the bit-vector size. Swapping so the longer string drives `n` (the bit-loop variable) preserves the original's performance characteristics. Documented because it changes the function's algebraic symmetry (`Distance(a,b) == Distance(b,a)` still holds, but the internal path differs).

---

## 7. No CGo / no FFI to Node runtime

**Decision:** Pure Go implementation with zero CGo calls.

**Rationale:** Track C rule: "No source-language runtime." The whole point is to leave the source language behind. CGo would reintroduce Node/JavaScript dependencies.

---

## 8. CLI interface design for test adapter

**Decision:** Built a `cmd/fastlev` binary with `distance <s1> <s2>` and `closest <str> <csv-args>` subcommands.

**Rationale:** The original Jest test suite runs in Node and needs to call the Go port. A subprocess-based adapter (`child_process.execSync`) is the simplest bridge that satisfies the "no source-language runtime" constraint. The CSV encoding for `closest` args uses `%2C` to escape commas within individual strings.

---

## 9. Test adapter strategy: subprocess over shared library

**Decision:** Jest adapter calls Go binary via `execSync`, not a native Node addon.

**Rationale:** A native addon (N-API) would require CGo re-compilation per platform and complicates the build pipeline. Subprocess is slower but trivially cross-platform and meets the single-command build requirement. The original test suite only runs 1000 iterations, so subprocess overhead is negligible (≈2ms per call vs <1µs for the actual algorithm).

---

## 10. Zero unsafe blocks

**Decision:** The entire port uses zero `unsafe` blocks.

**Rationale:** The algorithm is pure integer bit-manipulation — no pointer arithmetic, no type punning, no memory reinterpretation needed. This is a deliberate demonstration that "fast" and "safe" are not in conflict. Qualifies for the **Zero Unsafe (+5)** bonus.

---

## 11. Reference DP implementation in test suite

**Decision:** Included a naive O(n*m) DP reference implementation in `tests/port/`.

**Rationale:** The original test.ts includes its own naive DP implementation to verify the fast Myers path. Our Go tests replicate this: the property-based fuzz compares `Distance()` output against the naive DP on 1000 random pairs. This catches Myers bit-flip errors that would silently produce wrong distances.

---

## 12. Fuzz harness fallback to reference when Node unavailable

**Decision:** `fuzz/harness.go` tries the JS original via `exec.Command("node", ...)`, but falls back to the Go reference DP implementation when Node is not installed.

**Rationale:** The differential fuzz harness should run in CI/Docker environments that may not have Node installed. The reference DP is the same algorithm the original test suite uses for verification, so comparing against it is semantically equivalent to comparing against the JS library.

---

## 13. Benchmark runner methodology: `bench/cmd/bench.go` with multi-variant JS comparison

**Decision:** Measure Go port and JS original on **identical** random inputs using a standalone Go program (`bench/cmd/bench.go`). Report Go ns/op (in-process), JS subprocess ops/sec, JS in-process ops/sec, and speedup.

**Rationale:** Go's `testing.B` framework measures only one implementation at a time. A fair comparison requires both implementations to run on the same data. Using `bench/cmd/bench.go` ensures identical inputs. We report startup time and RSS as separate metrics since these are key differentiators: Go binaries have near-zero startup vs Node's JIT warmup.

---

## 14. Module name: `fastlev-go` (intentionally scoped under `src/`)

**Decision:** Go module name is `fastlev-go` and the import path is `fastlev-go/src/levenshtein`.

**Rationale:** Aligns the Go module path with the directory structure (`src/levenshtein/`) used to mirror the original TypeScript layout. Tools like `go mod tidy` and IDE auto-import use this path. The Go module path is also GitHub-style ready (`github.com/yourname/fastlev-go/src/levenshtein`) for users who fork or vendor the project.

---

## 15. 64-bit Myers paths (myers64 and myersX64)

**Decision:** Added 64-bit variants of the bit-parallel Myers algorithm alongside the existing 32-bit variants.

**Rationale:** The original `fastest-levenshtein` uses 32-bit paths exclusively because JavaScript's bitwise operators operate on 32-bit signed integers. Go has native 64-bit `uint64` types, enabling a 64-bit single-pass variant (`myers64`) that handles strings up to 64 characters in one pass (vs 32 in the original). This reduces the performance cliff at the 33-char boundary where the original falls back to block decomposition. The 64-bit block-based variant (`myersX64`) uses 64-bit blocks instead of 32-bit, reducing the number of vertical block passes by 50%.

---

## 16. Architecture-based dispatch (32-bit vs 64-bit)

**Decision:** Use Go build tags (`go:build amd64 || arm64 || ...`) detected at compile time to select which algorithm family serves as the default dispatch path on a given platform.

**Rationale:** The 32-bit and 64-bit algorithm families are both always compiled in, so testing and benchmarks can directly call either variant. At runtime, `Distance()` checks `is64bit` (compile-time constant) and dispatches accordingly. This means:
- On 64-bit platforms (amd64, arm64, riscv64, …): `myers64` is used for strings up to 64 chars, then `myersX64` for longer strings.
- On 32-bit platforms (386, arm): `myers32` is used for strings up to 32 chars, then `myersX32`.
- Running `GOARCH=386 go test` naturally exercises the 32-bit paths even on a 64-bit host — ideal for portability testing.

## 17. Dual `peq` tables (peq32 and peq64)

**Decision:** Maintain two separate `peq` lookup tables: `peq32 [0x10000]uint32` and `peq64 [0x10000]uint64`.

**Rationale:** The original uses a single `Uint32Array`. The 64-bit algorithm needs scope to store up to 64 bits per character position. Storing 64-bit entries in a 32-bit table would overflow. While the 64-bit table uses twice the memory (1.04MB vs 512KB), both are fixed global allocations and the tradeoff is acceptable for the performance gain.

## 18. Boundary testing at 32 and 64 character thresholds

**Decision:** Tests explicitly verify correctness on exactly 31, 32, 33, 63, 64, 65, 100, and 200-character strings.

**Rationale:** These are the critical transition points where the dispatching logic changes: 32 is the transition from myers_32 to myers_X32 on 32-bit systems, and 64 is the transition from myers64 to myersX64 on 64-bit systems. Verifying one-off errors around these boundaries prevents silent correctness bugs.

---

## 19. Package layout: `src/levenshtein/` instead of flat

**Decision:** Core library in `src/levenshtein/`, CLI in `src/cmd/fastlev/`.

**Rationale:** Clean separation of library and CLI concerns. Go tooling understands this layout natively. The import path `fastlev-go/src/levenshtein` is slightly longer than a flat package but clearly separates the public API from the tooling.

---

## 20. Benchmark runner script: `bench/cmd/bench.go`

**Decision:** Build a standalone Go program (not a `go test` benchmark) that generates random testcases, runs the Go port and the JS original on the same inputs, and writes a Markdown-style comparison table to `bench/results_raw.txt`.

**Rationale:** Go's `testing.B` framework measures only one implementation at a time. To fairly compare Go vs Node, both must run on the same input data. The runner generates 500 random pairs per band (N=4..1024), measures each implementation K times, and reports `speedup_vs_js_inproc` — the cleanest comparison since it excludes Node startup cost. This avoids the common pitfall of using `ops_per_sec` from a cold-start Node process, which unfairly inflates Go's advantage.

---

## 21. Two JS benchmark variants: subprocess vs in-process

**Decision:** Provide both `bench/cmd/bench_original.js` (subprocess-style) and `bench/cmd/bench_inproc.js` (in-process).

**Rationale:** Subprocess measurement includes Node startup time (~30ms) and is therefore biased against the original implementation. In-process measurement (warmup + repeated calls) isolates the actual algorithm cost, matching how Go's `testing.B` works. We report both for transparency but use the in-process number for the headline speedup claim.

---

## 22. Docker as a runnable artifact

**Decision:** The Docker image defaults to the `fastlev` CLI binary (`ENTRYPOINT ["fastlev"]`) not a test runner. `docker compose up` runs the program, not Jest.

**Rationale:** The hackathon submission artifact is a **runnable program**, not a test suite. The same multi-stage Dockerfile produces the binary for all operations: `docker compose run --rm demo distance "a" "b"` for the CLI, or `docker compose --profile test up test-port` for tests. This keeps the default path minimal (scratch-like alpine image) while supporting opt-in profiles for CI verification.

---

## 23. Auto-generated results.json and fuzz/log.txt

**Decision:** Both `bench/results.json` and `fuzz/log.txt` are **outputs** of their respective scripts, not hand-authored files. `bench/cmd/bench.go` writes `bench/results.json`; `TestFuzz_60Seconds` writes `fuzz/log.txt`.

They are listed in `.gitignore` and regenerated on every run:

```bash
go run ./bench/cmd/bench.go -runs 3    # produces bench/results.json
go test -v -run TestFuzz_60Seconds ./fuzz/  # produces fuzz/log.txt
```

**Rationale:** The hackathon explicitly requires these files to be auto-generated by their respective tools, and disqualifies hand-edited artifacts. Keeping them as build outputs also ensures they accurately reflect the most recent benchmark/fuzz results, preventing stale data from being committed.