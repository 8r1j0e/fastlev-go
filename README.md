# fastlev-go

> **Track C: TypeScript → Go** · Port Mortem 2026

A port of the [ka-weihe/fastest-levenshtein](https://github.com/ka-weihe/fastest-levenshtein) library from TypeScript to Go. It computes the Levenshtein edit distance between two strings (minimum number of single-characteredits: insertions, deletions, substitutions)

Uses Myers' bit-parallel algorithm with **automatic dispatch between 32-bit and 64-bit paths**:

| Platform | Single-pass | Block-based |
|---|---|---|
| **64-bit** (amd64, arm64, riscv64, …) | `myers64` (up to 64 chars) | `myersX64` (64-bit blocks) |
| **32-bit** (386, arm) | `myers32` (up to 32 chars) | `myersX32` (32-bit blocks) |

Both families compile in; the dispatcher is a `//go:build` tag constant — zero runtime cost.
On 64-bit platforms the single-pass threshold doubles from 32 to 64 chars, eliminating
block-decomposition overhead for strings 33–64 characters long.

---

## Quick Start

```bash
# Build
go build -o fastlev ./src/cmd/fastlev/

# Run
./fastlev distance "fast" "faster"         # → 2
./fastlev closest "hello" "world,help,hell"   # → hell
```

## API

```go
import "fastlev-go/src/levenshtein"

levenshtein.Distance("fast", "faster")                             // → 2
levenshtein.Closest("fast", []string{"slow","faster","fastest"})   // → "faster"
```

---

## Test Suite

```bash
# Go port tests (6 tests — known values, closest, random-vs-ref, boundaries)
go test -v ./tests/port/

# Original Jest test suite (runs against Go port via adapter)
go build -o fastlev ./src/cmd/fastlev/
cd tests/original && npm install && npm test && cd ../..

# Quick fuzz (500 pairs, ~15s)
go test -v -run TestFuzz_Quick -timeout 60s ./fuzz/

# 60-second fuzz — Differential Fuzz Survivor bonus (+5)
go test -v -run TestFuzz_60Seconds -timeout 120s ./fuzz/
```

## Benchmark

```bash
# Full Go vs Node comparison (9 bands, ~2 minutes)
go run ./bench/cmd/bench.go -runs 3

# View results
cat bench/results_raw.txt
cat bench/results.json
```

Both `bench/results.json` and `bench/results_raw.txt` are **auto-generated** by the runner.
See [bench/methodology.md](bench/methodology.md) for measurement methodology.

## Docker

The image defaults to the CLI binary. `docker compose up` runs the program.

```bash
# Build
docker compose build

# Run the CLI
docker compose run --rm demo --help
docker compose run --rm demo distance "kitten" "sitting"
docker compose run --rm demo closest "cat" "sat,bat,cab"

# Go port tests
docker compose --profile test up test-port

# Original Jest suite (vs Go port)
docker compose --profile test up test-original

# Quick fuzz
docker compose --profile fuzz up fuzz-quick

# 60-second fuzz
docker compose --profile fuzz up fuzz-survivor

# Go vs Node benchmark
docker compose --profile bench up bench
```

---

## Performance Highlights

| Metric | Go Port | Node.js (original) |
|---|---|---|
| Startup (cold) | ~0.00s | ~0.03s |
| Memory (RSS) | 2.3 MB | 50 MB |
| Single-pass max | 64 chars (2× original) | 32 chars |
| Zero `unsafe` blocks | ✓ | — (JIT) |
| Zero allocs (N ≤ 256) | ✓ | GC-managed |
| Speedup vs original | **2.5×–29×** | — |

---

## Project Structure

```
.
├── src/
│   ├── cmd/fastlev/main.go          # CLI entrypoint
│   └── levenshtein/
│       ├── myers.go                 # Core algorithm (myers32/64, myersX32/X64)
│       ├── levenshtein.go           # Public API (Distance, Closest)
│       ├── arch_64bit.go            # Build tag: 64-bit dispatch
│       └── arch_32bit.go            # Build tag: 32-bit dispatch
├── tests/
│   ├── original/                    # Unmodified TS test suite + Jest adapter
│   └── port/levenshtein_test.go     # Go port tests
├── fuzz/                            # Differential fuzz harness (Go vs JS)
├── bench/                           # Benchmark runner, methodology, outputs
├── Dockerfile                       # Multi-stage → alpine + binary
├── docker-compose.yml               # Runnable artifact + test/fuzz/bench profiles
├── DECISIONS.md                     # 23 architectural decisions
├── LICENSE                          # MIT
├── .gitignore                       # Ignores node_modules/ + generated artifacts
└── README.md
```

---

## License

[MIT](LICENSE) — Copyright (c) 2026 8r1j0e