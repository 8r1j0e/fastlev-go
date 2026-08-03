/**
 * adapter.js — bridges the original Jest test suite to the Go port.
 *
 * Replaces the original `require("./mod.js")` with calls to the Go CLI binary.
 * The original test.ts imports { distance, closest } from "./mod.js".
 * We intercept that import by proxying through execSync to the Go binary.
 *
 * Usage: jest --no-cache (via npm test)
 */
const { execSync } = require("child_process");

// Path to the Go binary — relative to tests/original/ in the repo
const FASTLEV_BIN = process.env.FASTLEV_PATH || "../../fastlev";

function run(args) {
  const cmd = `${FASTLEV_BIN} ${args}`;
  const stdout = execSync(cmd, { encoding: "utf8", timeout: 5000 });
  return stdout.trim();
}

// The original mod.js exports { distance, closest }.
// We override the global require cache so require("./mod.js") returns this object.
const mod = {
  distance: (a, b) => {
    const out = run(`distance "${a.replace(/"/g, '\\"')}" "${b.replace(/"/g, '\\"')}"`);
    return parseInt(out, 10);
  },
  closest: (str, arr) => {
    const joined = arr.map((s) => s.replace(/,/g, "%2C")).join(",");
    const out = run(`closest "${str.replace(/"/g, '\\"')}" "${joined}"`);
    return out;
  },
};

module.exports = mod;