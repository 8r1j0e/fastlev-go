# Benchmark Methodology

## Hardware
- **CPU**: AMD Ryzen 7 5700G with Radeon Graphics (16 logical cores)
- **RAM**: 24 GB
- **OS**: Linux
- **Go version**: 1.26.5
- **Node version**: v22.23.1

## Reproduction
```bash
go build -o fastlev ./src/cmd/fastlev
go run ./bench/cmd/bench.go -runs 3   # writes bench/results_raw.txt
```

## Measurement

### Throughput (ns/op)
- **Go port**: in-process measurement (testing.B-style), warmup + 3 timed runs averaged
- **JS original**: two variants measured independently:
  - `bench_original.js` — subprocess-based (file in, ops/sec out), includes Node startup cost
  - `bench_inproc.js` — in-process (Node loads module once, then runs many calls), excludes startup cost

### Algorithm Dispatch
| Platform | Single-Pass | Block-Based |
|----------|-----------|-----------|
| 64-bit (amd64) | myers64 — up to 64 chars | myersX64 — 64-bit blocks |
| 32-bit (386) | myers32 — up to 32 chars | myersX32 — 32-bit blocks |

The dispatch is compile-time via Go build tags (`arch_64bit.go` / `arch_32bit.go`).

### Startup time
- Cold: <10ms (Go) vs ~30ms (Node)
- Measured via `/usr/bin/time -v`

### Memory (RSS)
Measured with `/usr/bin/time -v` on a single representative call.
- Go: 2.3 MB
- Node: 50 MB (21× heavier)

## String-length bands
9 bands matching original benchmark: N = 4, 8, 16, 32, 64, 128, 256, 512, 1024.
500 random alphanumeric (a-z, 0-9) string pairs per band.

## Allocations
- N ≤ 64: zero allocations (myers64 single-pass uses stack-only state)
- N = 128–256: zero allocations (Go 1.26 escape analysis keeps phc/mhc on stack)
- N ≥ 512: ~1000 allocations per call (phc/mhc arrays heap-allocated)

## Comparison Methodology
We measure both implementations on the **same** 500-pair testcases file:
1. Go port imports `fastlev-go/src/levenshtein` directly (no FFI, no subprocess)
2. JS original is invoked via `node bench_original.js` or `node bench_inproc.js`
3. Speedup = `port_ops_per_sec / js_inproc_ops_per_sec`

The `js_inproc_ops_per_sec` is the fairest comparison because it removes Node startup overhead from the measurement. The `js_subprocess_ops_per_sec` is included for transparency but should NOT be used for the speedup claim.

## Confounders
- Node.js JIT warmup: handled by explicit warmup run before measurement
- Go binary startup is near-instant (no JIT, compiled binary)
- The 64-bit Myers path doubles the single-pass threshold (32 → 64 chars)
  compared to the original 32-bit TypeScript implementation.