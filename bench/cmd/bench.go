package main

// bench.go — runs the full benchmark suite and produces comparison reports.
//
// Usage:
//   go run ./bench/cmd/bench.go            # default: 9 bands, 500 pairs each
//   go run ./bench/cmd/bench.go -n 4       # single band
//   go run ./bench/cmd/bench.go -runs 3    # repeat each band 3× and average
//
// Outputs:
//   bench/results_raw.txt  — human-readable table
//   bench/results.json     — structured JSON

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand/v2"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"fastlev-go/src/levenshtein"
)

var sizes = []int{4, 8, 16, 32, 64, 128, 256, 512, 1024}

type bandRow struct {
	N         int
	PortNs    int64
	PortOps   float64
	JsSubproc float64
	JsInproc  float64
	Speedup   float64
}

func main() {
	count := flag.Int("count", 500, "pairs per band")
	runs := flag.Int("runs", 3, "repeat each band K times (averaged)")
	singleN := flag.Int("n", -1, "single band (overrides default)")
	outFile := flag.String("out", "bench/results_raw.txt", "output file")
	flag.Parse()

	npmCmd := exec.Command("npm", "install")
	npmCmd.Dir = "./bench/cmd"
	_, npmErr := npmCmd.CombinedOutput()

	if npmErr != nil {
		fmt.Printf("npm install failed: %s\n", npmErr)
	}

	if err := os.MkdirAll("bench/cmd/testcases", 0755); err != nil {
		fail(err)
	}

	bands := sizes
	if *singleN > 0 {
		bands = []int{*singleN}
	}

	var rows []bandRow

	for _, n := range bands {
		inputFile := fmt.Sprintf("bench/cmd/testcases/input_n%d.txt", n)
		genTestcases(inputFile, n, *count)

		jsSubproc := runOriginalBenchmark(inputFile, "bench/cmd/bench_original.js", "ORIGINAL_OPS", *runs)
		jsInproc := runOriginalBenchmark(inputFile, "bench/cmd/bench_inproc.js", "INPROC_OPS", *runs)
		portNs := runGoBenchmark(inputFile, *runs)

		portOps := float64(1e9) / float64(portNs)
		speedup := 0.0
		if jsInproc > 0 {
			speedup = portOps / jsInproc
		}

		rows = append(rows, bandRow{
			N: n, PortNs: portNs, PortOps: portOps,
			JsSubproc: jsSubproc, JsInproc: jsInproc, Speedup: speedup,
		})

		fmt.Printf("N=%-4d %-12d %-13.2f %-12.2f %-13.2f %.2fx\n",
			n, portNs, portOps, jsSubproc, jsInproc, speedup)
	}

	// Write human-readable table
	lines := []string{
		"Band   Port ns/op   Port ops/sec   JS subproc   JS inproc    Speedup",
		"----------------------------------------------------------------------",
	}
	for _, r := range rows {
		lines = append(lines, fmt.Sprintf("N=%-4d %-12d %-13.2f %-12.2f %-13.2f %.2fx",
			r.N, r.PortNs, r.PortOps, r.JsSubproc, r.JsInproc, r.Speedup))
	}
	if err := os.WriteFile(*outFile, []byte(strings.Join(lines, "\n")+"\n"), 0644); err != nil {
		fail(err)
	}
	fmt.Printf("\nReport written to %s\n", *outFile)

	// Write structured JSON
	jsonPath := filepath.Join(filepath.Dir(*outFile), "results.json")
	jsonData, err := json.MarshalIndent(map[string]any{
		"generated_by": "bench/cmd/bench.go",
		"timestamp":    time.Now().Format(time.RFC3339),
		"bands":        rows,
		"methodology":  "See bench/methodology.md",
	}, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "warn: json marshal: %v\n", err)
	} else if err := os.WriteFile(jsonPath, jsonData, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "warn: write %s: %v\n", jsonPath, err)
	} else {
		fmt.Printf("JSON written to %s\n", jsonPath)
	}
}

func genTestcases(path string, n, count int) {
	f, err := os.Create(path)
	if err != nil {
		fail(err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	for i := 0; i < count; i++ {
		fmt.Fprintf(w, "%s|%s\n", randString(n), randString(n))
	}
	w.Flush()
}

func randString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = byte('a' + rand.IntN(26))
	}
	return string(b)
}

func runGoBenchmark(path string, runs int) int64 {
	pairs, err := readPairs(path)
	if err != nil {
		fail(err)
	}
	for _, p := range pairs {
		levenshtein.Distance(p.a, p.b)
	}
	var totalNs int64
	for r := 0; r < runs; r++ {
		t0 := time.Now()
		for _, p := range pairs {
			levenshtein.Distance(p.a, p.b)
		}
		totalNs += time.Since(t0).Nanoseconds()
	}
	return totalNs / int64(runs) / int64(len(pairs))
}

func runOriginalBenchmark(path, script, key string, runs int) float64 {
	var ops float64
	for r := 0; r < runs; r++ {
		cmd := exec.Command("node", script, path)
		out, err := cmd.Output()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s failed on run %d: %v\n", script, r, err)
			continue
		}
		for _, line := range strings.Split(string(out), "\n") {
			if strings.HasPrefix(line, key+"=") {
				v, err := strconv.ParseFloat(strings.TrimPrefix(line, key+"="), 64)
				if err == nil {
					ops = v
				}
			}
		}
	}
	return ops
}

func readPairs(path string) ([]pair, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []pair
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1024*1024), 1024*1024*10)
	for sc.Scan() {
		line := sc.Text()
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("bad line: %q", line)
		}
		out = append(out, pair{parts[0], parts[1]})
	}
	return out, sc.Err()
}

type pair struct{ a, b string }

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
