#!/usr/bin/env node

/**
 * bench_original.js
 *
 * Reads a testcases file (pairs of strings separated by "|"), runs each pair
 * through the fastest-levenshtein package's distance function, and prints
 * ops/sec on a single line as ORIGINAL_OPS=<n>.
 *
 * Usage: node bench/cmd/bench_original.js <testcase_file>
 */

const fs = require("fs");

const inPath = process.argv[2];
if (!inPath) {
    console.error("usage: node bench/cmd/bench_original.js <testcase_file>");
    process.exit(1);
}

const lines = fs.readFileSync(inPath, "utf-8")
    .split("\n")
    .filter((l) => l.length > 0);

const pairs = lines.map((l) => l.split("|", 2));

// Try the npm module first; fall back to a local path.
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

// Warmup (JIT priming).
for (const [a, b] of pairs) distance(a, b);

// Measured run.
const start = process.hrtime.bigint();
for (const [a, b] of pairs) distance(a, b);
const elapsedNs = Number(process.hrtime.bigint() - start);

const opsPerSec = pairs.length / (elapsedNs / 1e9);
console.log(`original  ops/sec: ${opsPerSec.toFixed(2)}`);
console.log(`original  ns/op:   ${(elapsedNs / pairs.length).toFixed(2)}`);
console.log(`ORIGINAL_OPS=${opsPerSec.toFixed(2)}`);
console.log(`ORIGINAL_NS_PER_OP=${(elapsedNs / pairs.length).toFixed(2)}`);