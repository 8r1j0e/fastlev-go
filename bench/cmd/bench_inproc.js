#!/usr/bin/env node

/**
 * bench_inproc.js
 *
 * In-process JS benchmark — measures the original fastest-levenshtein
 * distance function the same way Go's testing.B does (repeated calls
 * in the same process).  This avoids Node startup cost dominating the
 * measurement for small N.
 *
 * Usage: node bench/cmd/bench_inproc.js <testcase_file>
 *
 * Output: INPROC_OPS=<n> and INPROC_NS_PER_OP=<n>
 */

const fs = require("fs");

const inPath = process.argv[2];
if (!inPath) {
    console.error("usage: node bench/cmd/bench_inproc.js <testcase_file>");
    process.exit(1);
}

let distance;
try {
    ({ distance } = require("fastest-levenshtein"));
} catch {
    try {
        ({ distance } = require("./fastest-levenshtein/mod"));
    } catch {
        console.error("fastest-levenshtein not found");
        process.exit(2);
    }
}

const lines = fs.readFileSync(inPath, "utf-8").split("\n").filter(Boolean);
const pairs = lines.map((l) => l.split("|", 2));

// Warmup
for (const [a, b] of pairs) distance(a, b);

// Measure
const N_RUNS = 5;
const total = pairs.length;

let bestNsPerOp = Infinity;
for (let r = 0; r < N_RUNS; r++) {
    const t0 = process.hrtime.bigint();
    for (let i = 0; i < total; i++) {
        distance(pairs[i][0], pairs[i][1]);
    }
    const elapsed = Number(process.hrtime.bigint() - t0);
    const nsPerOp = elapsed / total;
    if (nsPerOp < bestNsPerOp) bestNsPerOp = nsPerOp;
}

const opsPerSec = 1e9 / bestNsPerOp;
console.log(`js inproc  ops/sec: ${opsPerSec.toFixed(2)}`);
console.log(`js inproc  ns/op:   ${bestNsPerOp.toFixed(2)}`);
console.log(`INPROC_OPS=${opsPerSec.toFixed(2)}`);
console.log(`INPROC_NS_PER_OP=${bestNsPerOp.toFixed(2)}`);